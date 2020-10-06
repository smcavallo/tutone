package lang

import (
	"fmt"
	"sort"
	"strings"

	"github.com/newrelic/tutone/internal/config"
	"github.com/newrelic/tutone/internal/schema"

	log "github.com/sirupsen/logrus"
)

// TODO: Move CommandGenerator and its friends to a proper home
type CommandGenerator struct {
	PackageName string
	Imports     []string
	Commands    []Command
}

type InputObject struct {
	Name   string
	GoType string
}

type Command struct {
	Name             string
	ShortDescription string
	LongDescription  string
	Example          string
	InputType        string
	ClientMethod     string
	ClientMethodArgs []string
	InputObjects     []InputObject
	Flags            []CommandFlag
	Subcommands      []Command

	GraphQLPath []string // Should mutations also use this? Probably
}

type CommandFlag struct {
	Name           string
	Type           string
	FlagMethodName string
	DefaultValue   string
	Description    string
	VariableName   string
	VariableType   string
	ClientType     string
	Required       bool
	IsInputType    bool
}

type CommandExampleData struct {
	CLIName     string
	PackageName string
	Command     string
	Subcommand  string
	Flags       []CommandFlag
}

// GolangGenerator is enough information to generate Go code for a single package.
type GolangGenerator struct {
	Types       []GoStruct
	PackageName string
	Enums       []GoEnum
	Imports     []string
	Scalars     []GoScalar
	Interfaces  []GoInterface
	Mutations   []GoMethod
	Queries     []GoMethod
}

type GoStruct struct {
	Name             string
	Description      string
	Fields           []GoStructField
	Implements       []string
	SpecialUnmarshal bool
}

type GoStructField struct {
	Name        string
	Type        string
	TypeName    string
	Tags        string
	TagKey      string
	Description string
	IsInterface bool
	IsList      bool
}

type GoEnum struct {
	Name        string
	Description string
	Values      []GoEnumValue
}

type GoEnumValue struct {
	Name        string
	Description string
}

type GoScalar struct {
	Name        string
	Description string
	Type        string
}

type GoInterface struct {
	Name          string
	Description   string
	Type          string
	PossibleTypes []GoInterfacePossibleType
}

type GoInterfacePossibleType struct {
	GoName      string
	GraphQLName string
}

type GoMethod struct {
	Description string
	Name        string
	QueryVars   []QueryVar
	Signature   GoMethodSignature
	QueryString string
	// ResponseObjectType is the name of the type for the API response.  Note that this is not the method return, but the API call response.
	ResponseObjectType string
}

type GoMethodSignature struct {
	Input  []GoMethodInputType
	Return []string
	// ReturnSlice indicates if the response is a slice of objects or not.  Used to flag KindList.
	ReturnSlice bool
	// Return path is the fields on the response object that nest the results the given method will return.
	ReturnPath []string
}

type GoMethodInputType struct {
	Name string
	Type string
}

type QueryVar struct {
	Key   string
	Value string
	Type  string
}

// GenerateGoMethodQueriesForPackage uses the provided configuration to generate the GoMethod structs that contain the information about performing GraphQL queries.
func GenerateGoMethodQueriesForPackage(s *schema.Schema, genConfig *config.GeneratorConfig, pkgConfig *config.PackageConfig) (*[]GoMethod, error) {
	var methods []GoMethod

	for _, pkgQuery := range pkgConfig.Queries {

		typePath, err := s.LookupQueryTypesByFieldPath(pkgQuery.Path)
		if err != nil {
			log.Error(err)
			continue
		}

		// TODO this will eventually break when a field name of the struct is not a
		// simple capitalization.  We'd need to loop over the fields for the type
		// and grab the name as is done in constrainedResponseStructs().
		returnPath := []string{}
		for _, t := range pkgQuery.Path {
			returnPath = append(returnPath, strings.Title(t))
		}

		// The endpoint we care about will always be on the last of the path elements specified.
		t := typePath[len(typePath)-1]

		// Find the intersection of the received endpoint and the field name on the last type in the path.
		// For example, given the following path...
		// actor { cloud { } }
		// ... we want to generate the method based on the Type of the field 'cloud'.
		for _, endpoint := range pkgQuery.Endpoints {
			for _, field := range t.Fields {
				if field.Name == endpoint.Name {

					method := goMethodForField(s, field, pkgConfig)

					method.QueryString = s.GetQueryStringForEndpoint(typePath, pkgQuery.Path, endpoint.Name, endpoint.MaxQueryFieldDepth)
					method.ResponseObjectType = fmt.Sprintf("%sResponse", endpoint.Name)
					method.Signature.ReturnPath = returnPath

					methods = append(methods, method)
				}
			}
		}
	}

	if len(methods) > 0 {
		sort.SliceStable(methods, func(i, j int) bool {
			return methods[i].Name < methods[j].Name
		})
		return &methods, nil
	}

	return &methods, nil
}

// GenerateGoMethodMutationsForPackage uses the provided configuration to generate the GoMethod structs that contain the information about performing GraphQL mutations.
func GenerateGoMethodMutationsForPackage(s *schema.Schema, genConfig *config.GeneratorConfig, pkgConfig *config.PackageConfig) (*[]GoMethod, error) {
	var methods []GoMethod

	if len(pkgConfig.Mutations) == 0 {
		return nil, nil
	}

	// for _, field := range s.MutationType.Fields {
	for _, pkgMutation := range pkgConfig.Mutations {
		field, err := s.LookupMutationByName(pkgMutation.Name)
		if err != nil {
			log.Error(err)
			continue
		}

		if field == nil {
			log.Errorf("unable to generate mutation from nil field, %s", pkgMutation.Name)
			continue
		}

		// if field.Name == pkgMutation.Name {
		method := goMethodForField(s, *field, pkgConfig)
		// method.QueryString = schema.PrefixLineTab(s.QueryFieldsForTypeName(field.Type.GetTypeName(), pkgMutation.MaxQueryFieldDepth))
		method.QueryString = s.GetQueryStringForMutation(field, pkgMutation.MaxQueryFieldDepth)

		methods = append(methods, method)
		// }
	}
	// }

	if len(methods) > 0 {
		sort.SliceStable(methods, func(i, j int) bool {
			return methods[i].Name < methods[j].Name
		})
		return &methods, nil
	}

	return nil, fmt.Errorf("no methods for package")
}

func GenerateGoTypesForPackage(s *schema.Schema, genConfig *config.GeneratorConfig, pkgConfig *config.PackageConfig, expandedTypes *[]*schema.Type) (*[]GoStruct, *[]GoEnum, *[]GoScalar, *[]GoInterface, error) {
	// TODO: Putting the types in the specified path should be optional
	//       Should we use a flag or allow the user to omit that field in the config? ¿Por que no lost dos?

	var structsForGen []GoStruct
	var enumsForGen []GoEnum
	var scalarsForGen []GoScalar
	var interfacesForGen []GoInterface

	for _, t := range *expandedTypes {
		switch t.Kind {
		case schema.KindInputObject, schema.KindObject, schema.KindInterface:
			xxx := GoStruct{
				Name:        t.GetName(),
				Description: t.GetDescription(),
			}

			var fields []schema.Field
			fields = append(fields, t.Fields...)
			fields = append(fields, t.InputFields...)

			fieldErrs := []error{}
			for _, f := range fields {

				// If any of the fields for this type are an interface type, then we
				// need to signal to the template an UnmarshalJSON() should be
				// rendered.
				for _, k := range f.Type.GetKinds() {
					if k == schema.KindInterface {
						xxx.SpecialUnmarshal = true
					}
				}

				xxx.Fields = append(xxx.Fields, getStructField(f, pkgConfig))
			}

			if len(fieldErrs) > 0 {
				log.Error(fieldErrs)
			}

			var implements []string
			for _, x := range t.Interfaces {
				implements = append(implements, x.GetName())
			}

			xxx.Implements = implements

			if t.Kind == schema.KindInterface {
				// Modify the struct type to avoid conflict with the interface type by the same name.
				// xxx.Name += "Type"

				// Ensure that the struct for the graphql interface implements the go interface
				xxx.Implements = append(xxx.Implements, t.GetName())

				// Handle the interface
				yyy := GoInterface{
					Description: t.GetDescription(),
					Name:        t.GetName(),
				}

				// Inform the template about which possible implementations exist for
				// this interface.  We need to know about both the name that GraphQL
				// uses and the name that Go uses.  This is to allow some flexibility
				// in the template for how to reference the implementation information.
				for _, x := range t.PossibleTypes {
					ttt := GoInterfacePossibleType{
						GraphQLName: x.Name,
						GoName:      x.GetName(),
					}

					yyy.PossibleTypes = append(yyy.PossibleTypes, ttt)
				}

				interfacesForGen = append(interfacesForGen, yyy)
			}

			sort.SliceStable(xxx.Fields, func(i, j int) bool {
				return xxx.Fields[i].Name < xxx.Fields[j].Name
			})

			structsForGen = append(structsForGen, xxx)
		case schema.KindENUM:
			xxx := GoEnum{
				Name:        t.GetName(),
				Description: t.GetDescription(),
			}

			for _, v := range t.EnumValues {
				value := GoEnumValue{
					Name:        v.GetName(),
					Description: v.GetDescription(),
				}

				xxx.Values = append(xxx.Values, value)
			}

			enumsForGen = append(enumsForGen, xxx)
		case schema.KindScalar:
			// Default scalars to string
			createAs := "string"
			skipTypeCreate := false
			nameToMatch := t.GetName()

			var seenNames []string
			for _, p := range pkgConfig.Types {
				if stringInStrings(p.Name, seenNames) {
					log.Warnf("duplicate package config name detected: %s", p.Name)
					continue
				}
				seenNames = append(seenNames, p.Name)

				if p.Name == nameToMatch {
					if p.CreateAs != "" {
						createAs = p.CreateAs
					}

					if p.SkipTypeCreate {
						skipTypeCreate = true
					}
				}
			}

			if !t.IsGoType() && !skipTypeCreate {
				xxx := GoScalar{
					Description: t.GetDescription(),
					Name:        t.GetName(),
					Type:        createAs,
				}

				scalarsForGen = append(scalarsForGen, xxx)
			}
		// case schema.KindInterface:
		// 	xxx := GoInterface{
		// 		Description: t.GetDescription(),
		// 		Name:        t.GetName(),
		// 	}
		//
		// 	interfacesForGen = append(interfacesForGen, xxx)
		default:
			log.WithFields(log.Fields{
				"name": t.Name,
				"kind": t.Kind,
			}).Warn("kind not implemented")
		}
	}

	structsForGen = append(structsForGen, constrainedResponseStructs(s, pkgConfig, expandedTypes)...)

	sort.SliceStable(structsForGen, func(i, j int) bool {
		return structsForGen[i].Name < structsForGen[j].Name
	})

	sort.SliceStable(enumsForGen, func(i, j int) bool {
		return enumsForGen[i].Name < enumsForGen[j].Name
	})

	sort.SliceStable(scalarsForGen, func(i, j int) bool {
		return scalarsForGen[i].Name < scalarsForGen[j].Name
	})

	sort.SliceStable(interfacesForGen, func(i, j int) bool {
		return interfacesForGen[i].Name < interfacesForGen[j].Name
	})

	return &structsForGen, &enumsForGen, &scalarsForGen, &interfacesForGen, nil
}

func getStructField(f schema.Field, pkgConfig *config.PackageConfig) GoStructField {
	var typeName string
	var typeNamePrefix string
	var typeNameSuffix string
	var err error

	typeName, err = f.GetTypeNameWithOverride(pkgConfig)
	if err != nil {
		log.Error(err)
	}

	kinds := f.Type.GetKinds()

	var isList bool

	// In the case we have a LIST type, we need to prefix the type with the slice
	// descriptor.  This can appear pretty much anywhere in a list of kinds, but
	// we ignore the order here.
	for _, k := range kinds {
		if k == schema.KindList {
			typeNamePrefix = "[]"
			isList = true
			break
		}
	}

	// Used to signal the template that the UnmarshalJSON should handle this field as an Interface.
	var isInterface bool

	// In the case a field type is of type Interface, we need to ensure we
	// append the term "Interface" to it, as is done in the "Implements"
	// below.
	if kinds[len(kinds)-1] == schema.KindInterface {
		typeNameSuffix = "Interface"
		isInterface = true
	}

	return GoStructField{
		Description: f.GetDescription(),
		Name:        f.GetName(),
		TagKey:      f.Name,
		Tags:        f.GetTags(),
		IsInterface: isInterface,
		IsList:      isList,
		Type:        fmt.Sprintf("%s%s%s", typeNamePrefix, typeName, typeNameSuffix),
		TypeName:    typeName,
	}
}

// constrainedResponseStructs is used to create response objects that contain
// fields that already exist in the expandedTypes.  This avoids creating full
// structs, and limits response objects to those types that are already
// referenced in the expandedTypes.
func constrainedResponseStructs(s *schema.Schema, pkgConfig *config.PackageConfig, expandedTypes *[]*schema.Type) []GoStruct {
	var goStructs []GoStruct

	// Determine if the typeName received exists in the received expandedTypes list.
	isExpanded := func(expandedTypes *[]*schema.Type, typeName string) bool {
		for _, t := range *expandedTypes {
			if t.GetName() == typeName {
				return true
			}
		}

		return false
	}

	// Determine if the received typeName is in the  received schema.Type list.
	isInPath := func(types []*schema.Type, typeName string) bool {
		for _, t := range types {
			if t.GetName() == typeName {
				return true
			}
		}

		return false
	}

	// Build a response object for each one of the queries in the configuration.
	for _, query := range pkgConfig.Queries {

		// Retrieve the corresponding types for each of the field names in the query config.
		pathTypes, err := s.LookupQueryTypesByFieldPath(query.Path)
		if err != nil {
			log.Error(err)
			continue
		}

		// Ensure that all of the types that we will depend on in our response struct below are present.
		for _, t := range pathTypes {
			// Skip doing anything with this type if it has already been expanded.
			if isExpanded(expandedTypes, t.GetName()) {
				continue
			}

			xxx := GoStruct{
				Name:        t.GetName(),
				Description: t.GetDescription(),
			}

			for _, f := range t.Fields {
				if isExpanded(expandedTypes, f.Type.GetTypeName()) || isInPath(pathTypes, f.Type.GetName()) {

					xxx.Fields = append(xxx.Fields, getStructField(f, pkgConfig))
				}
			}

			goStructs = append(goStructs, xxx)
		}

		// Ensure we have a response struct for each of the endpoints in our config.
		for _, endpoint := range query.Endpoints {

			xxx := GoStruct{
				Name: fmt.Sprintf("%sResponse", endpoint.Name),
			}

			// For the top level response object, we only use the first field path that is received from the user.
			firstType := pathTypes[0]

			field := GoStructField{
				Name: firstType.GetName(),
				Type: firstType.GetName(),
				Tags: fmt.Sprintf("`json:\"%s\"`", query.Path[0]),
			}

			xxx.Fields = append(xxx.Fields, field)
			goStructs = append(goStructs, xxx)
		}

	}

	return goStructs
}

// goMethodForField creates a new GoMethod based on a field.  Note that the
// implementation specific information like QueryString are not added to the
// method, and it is up to the caller to flavor the method accordingly.
func goMethodForField(s *schema.Schema, field schema.Field, pkgConfig *config.PackageConfig) GoMethod {

	method := GoMethod{
		Name:        field.GetName(),
		Description: field.GetDescription(),
	}

	var prefix string
	kinds := field.Type.GetKinds()
	if kinds[0] == schema.KindList {
		prefix = "[]"
		method.Signature.ReturnSlice = true
	}

	pointerReturn := fmt.Sprintf("%s%s", prefix, field.Type.GetTypeName())
	method.Signature.Return = []string{pointerReturn, "error"}

	for _, methodArg := range field.Args {
		typeName, err := methodArg.GetTypeNameWithOverride(pkgConfig)
		if err != nil {
			log.Error(err)
			continue
		}

		inputType := GoMethodInputType{
			Name: methodArg.GetName(),
			Type: typeName,
		}

		queryVar := QueryVar{
			Key:   methodArg.Name,
			Value: inputType.Name,
			Type:  methodArg.Type.GetTypeName(),
		}

		method.QueryVars = append(method.QueryVars, queryVar)

		method.Signature.Input = append(method.Signature.Input, inputType)
	}

	return method
}
