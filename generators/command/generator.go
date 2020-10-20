package command

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/huandu/xstrings"
	log "github.com/sirupsen/logrus"

	"github.com/newrelic/tutone/internal/codegen"
	"github.com/newrelic/tutone/internal/config"
	"github.com/newrelic/tutone/internal/schema"
	"github.com/newrelic/tutone/internal/util"
	"github.com/newrelic/tutone/pkg/lang"
)

type Generator struct {
	lang.CommandGenerator
}

var goTypesToCobraFlagMethodMap = map[string]string{
	"int":    "IntVar",
	"string": "StringVar",
}

func hydrateCommand(s *schema.Schema, command config.Command, pkgConfig *config.PackageConfig) lang.Command {
	isBaseCommand := true
	cmdVarName := "Command"

	// Handle case where this command is a subcommand of a parent entry point
	// Note: this doesn't do anything yet,
	if !isBaseCommand {
		cmdVarName = fmt.Sprintf("cmd%s", xstrings.ToCamelCase(command.Name))
	}

	cmd := lang.Command{
		Name:             command.Name,
		CmdVariableName:  cmdVarName,
		ShortDescription: command.ShortDescription,
		LongDescription:  command.LongDescription,
		Example:          command.Example,
		GraphQLPath:      command.GraphQLPath,
	}

	if len(command.Subcommands) == 0 {
		return lang.Command{}
	}

	cmd.Subcommands = make([]lang.Command, len(command.Subcommands))

	for i, subCmdConfig := range command.Subcommands {
		var err error
		// var mutationCmdData *schema.Field
		// var queryCmdData *schema.Field
		var queryPathTypes []*schema.Type
		var subcommandMetadata *schema.Field

		// Check to see if the commands CRUD action is a mutation.
		// If it's not a mutation, then it's a query (read) request.
		subcommandMetadata, err = s.LookupMutationByName(subCmdConfig.Name)

		// If the command is not a mutation, move forward with
		// generating a command to perform a query request.
		if err != nil {
			// TODO: This works for now, but schema.LookupQueryTypesByFieldPath is optimized
			// for generating types in the client. In the case of the CLI we only need the last
			// part of the path (the endpoint). A DFS lookup could be quicker, although the
			// performance hit might be negligible.
			queryPathTypes, err = s.LookupQueryTypesByFieldPath(subCmdConfig.GraphQLPath)
			if err != nil {
				log.Fatalf("query endpoint not found: %s", err)
			}

			graphQLEndpoint := subCmdConfig.GraphQLPath[len(subCmdConfig.GraphQLPath)-1]

			for _, queryStep := range queryPathTypes {
				for _, field := range queryStep.Fields {
					if field.Name == graphQLEndpoint {
						subcommandMetadata = &field
					}
				}
			}
		}

		subcommand := hydrateSubcommand(s, subcommandMetadata, subCmdConfig)

		exampleData := lang.CommandExampleData{
			CLIName:     "newrelic",
			PackageName: pkgConfig.Name,
			Command:     cmd.Name,
			Subcommand:  subcommand.Name,
			Flags:       subcommand.Flags,
		}

		subcommand.Example = subCmdConfig.Example
		if subCmdConfig.Example == "" {
			sCmdExample, err := generateCommandExample(subcommandMetadata, exampleData)
			if err != nil {
				log.Fatal(err)
			}

			subcommand.Example = sCmdExample
		}

		cmd.Subcommands[i] = *subcommand
	}

	return cmd
}

func generateCommandExample(sCmd *schema.Field, data lang.CommandExampleData) (string, error) {
	t := `{{ .CLIName }} {{ .PackageName }} {{ .Command }} {{ .Subcommand }}{{- range .Flags }} --{{ .Name }}{{ end }}`

	return codegen.RenderTemplate("commandExample", t, data)
}

func hydrateFlagsFromSchema(args []schema.Field, cmdConfig config.Command) []lang.CommandFlag {
	var flags []lang.CommandFlag

	for _, arg := range args {
		var variableName string

		isInputObject := arg.Type.IsInputObject()
		if isInputObject {
			// Add 'Input' suffix to the input variable name
			variableName = fmt.Sprintf("%sInput", cmdConfig.Name)
		} else {
			variableName = fmt.Sprintf("%s%s", cmdConfig.Name, arg.Name)
		}

		typ, _, _ := arg.Type.GetType()
		typeName := arg.Type.GetTypeName()

		// TODO: Put this into a helper method or find an existing helper method
		isOfTypeScalarID := typeName == "ID" && arg.Type.OfType != nil && arg.Type.OfType.Kind == schema.KindScalar
		if isOfTypeScalarID {
			typ = "string"
		}

		variableType := "string"
		if arg.IsGoType() {
			variableType = typ
		}

		isRequired := arg.Type.Kind == schema.KindNonNull
		// isInputObject := arg.Type.OfType.Kind == schema.KindInputObject
		clientType := fmt.Sprintf("%s.%s", cmdConfig.ClientPackageName, typ)

		flag := lang.CommandFlag{
			Name:           arg.Name,
			Type:           typ,
			FlagMethodName: getCobraFlagMethodName(typ),
			DefaultValue:   "",
			Description:    arg.Description,
			VariableName:   variableName,
			VariableType:   variableType,
			Required:       isRequired,
			IsInputType:    isInputObject,
			ClientType:     clientType,
		}

		flags = append(flags, flag)
	}

	return flags
}

func getCobraFlagMethodName(typeString string) string {
	if v, ok := goTypesToCobraFlagMethodMap[typeString]; ok {
		return v
	}

	// Almost all CRUD inputs will be a JSON string
	return "StringVar"
}

func (g *Generator) Generate(s *schema.Schema, genConfig *config.GeneratorConfig, pkgConfig *config.PackageConfig) error {
	log.Debugf("Generate...")

	g.PackageName = pkgConfig.Name
	g.Imports = pkgConfig.Imports

	cmds := make([]lang.Command, len(pkgConfig.Commands))
	for i, c := range pkgConfig.Commands {
		cmds[i] = hydrateCommand(s, c, pkgConfig)
	}

	g.Commands = cmds

	return nil
}

// const templateURL = "https://raw.githubusercontent.com/newrelic/tutone/master/templates/command/command.go.tmpl"

func getRemoteTemplateAsString(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var respString string
	if resp.StatusCode == http.StatusOK {
		respBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		respString = string(respBytes)
	}

	return respString, nil
}

func (g *Generator) Execute(genConfig *config.GeneratorConfig, pkgConfig *config.PackageConfig) error {
	log.Debugf("Execute...")

	var templateStr string
	var err error

	destinationPath := GetDestinationPath(pkgConfig)
	if err = MakeDir(destinationPath); err != nil {
		return err
	}

	hasTemplateURL := genConfig.TemplateURL != ""
	hasTemplateDir := genConfig.TemplateDir != ""

	if hasTemplateURL {
		// TODO: handle error and optimize
		templateStr, _ = getRemoteTemplateAsString(genConfig.TemplateURL)
	}

	for _, command := range pkgConfig.Commands {
		fileName := fmt.Sprintf("command_%s.go", util.ToSnakeCase(command.Name))

		// Default template name is '{{ packageName }}.go.tmpl'
		templateName := "command.go.tmpl"
		if genConfig.TemplateName != "" {
			templateName = genConfig.TemplateName
		}

		fPath := fmt.Sprintf("%s/%s", destinationPath, fileName)
		destinationFile, err := codegen.RenderStringFromGenerator(fPath, g)
		if err != nil {
			return err
		}

		if hasTemplateURL && hasTemplateDir {
			return fmt.Errorf("generator configuration error: please set `templateDir` or `templateURL`, but not both")
		}

		templateDir := "templates/command"
		if hasTemplateDir {
			templateDir, err = codegen.RenderStringFromGenerator(genConfig.TemplateDir, g)
			if err != nil {
				return err
			}
		}

		c := codegen.CodeGen{
			TemplateDir:     templateDir,
			TemplateName:    templateName,
			DestinationFile: destinationFile,
			DestinationDir:  destinationPath,
		}

		// TODO: Provide a default template URL

		if templateStr != "" {
			if err := c.WriteFileFromTemplateString(g, templateStr); err != nil {
				return err
			}
		} else {
			if err := c.WriteFile(g); err != nil {
				return err
			}
		}

		printSuccessMessage(pkgConfig, destinationFile)
	}

	return nil
}

/////
// HELPER FUNCTIONS - TODO: Move to proper home
/////
func GetDestinationPath(pkgConfig *config.PackageConfig) string {
	if pkgConfig.Path != "" {
		return pkgConfig.Path
	}

	return "./"
}

// MakeDir creates a directory if it does't exist yet and sets
// folder permissions to 0755.
func MakeDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.Mkdir(path, 0755); err != nil {
			return err
		}
	}

	return nil
}

// printSuccessMessage prints a message to the terminal letting the user know
// that code generation was a success and outputs the package and file path for reference.
//
// Emoji unicode reference: http://www.unicode.org/emoji/charts/emoji-list.html
func printSuccessMessage(pkgConfig *config.PackageConfig, filePath string) {
	fmt.Print("\n\u2705 Code generation complete: \n\n")
	fmt.Printf("   Package:   %v \n", pkgConfig.Path)
	fmt.Printf("   File:      %v \n", filePath)
	fmt.Println("")
}
