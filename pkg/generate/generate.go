package generate

import (
	"errors"
	"fmt"
	"os"
	"sort"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/newrelic/tutone/internal/schema"
)

// The big show
func Generate() error {
	fmt.Print("\n GENERATE..... \n")

	defFile := viper.GetString("definition")
	schemaFile := viper.GetString("schema_file")
	typesFile := viper.GetString("generate.types_file")
	// packageName := viper.GetString("package")

	log.WithFields(log.Fields{
		"definition_file": defFile,
		"schema_file":     schemaFile,
		"types_file":      typesFile,
	}).Info("Loading generation config")

	// load the config
	cfg, err := LoadConfig(defFile)
	if err != nil {
		return err
	}

	// package is required
	if len(cfg.Packages) == 0 {
		return errors.New("an array of packages is required")
	}

	// Load the schema
	s, err := schema.Load(schemaFile)
	if err != nil {
		return err
	}
	log.WithFields(log.Fields{
		"schema": s,
	}).Trace("loaded schema")

	if err := doGenerate(s, cfg); err != nil {
		return err
	}

	log.Printf("successfully generated types")

	return nil
}

func doGenerate(schemaInput *schema.Schema, config *Config) error {
	for _, pkg := range config.Packages {
		if err := schema.ResolveSchemaTypes(*schemaInput, pkg.Types); err != nil {
			return err
		}

		// TODO: Update return pattern to be tuple? - e.g. (result, err)
		if err := generateTypesForPackage(pkg, schemaInput); err != nil {
			return err
		}
	}

	return nil
}

func generateTypesForPackage(pkg Package, schemaInput *schema.Schema) error {
	// TODO: Putting the types in the specified path should be optional
	//       Should we use a flag or allow the user to omit that field in the config? ¿Por que no lost dos?

	// Default to project root for types
	destinationPath := "./"
	if pkg.Path != "" {
		destinationPath = pkg.Path
	}

	if _, err := os.Stat(destinationPath); os.IsNotExist(err) {
		if err := os.Mkdir(destinationPath, 0755); err != nil {
			log.Error(err)
		}
	}

	// Default file name is 'types.go'
	fileName := "types.go"
	if pkg.FileName != "" {
		fileName = pkg.FileName
	}

	filePath := fmt.Sprintf("%s/%s", destinationPath, fileName)
	f, err := os.Create(filePath)
	if err != nil {
		log.Error(err)
	}

	// Generate the types code via `WriteString`
	// TODO: template(s)
	_, err = f.WriteString(fmt.Sprintf("// Code generated by typegen; DO NOT EDIT.\n\npackage %s\n\n", pkg.Name))
	if err != nil {
		log.Error(err)
	}

	defer f.Close()

	// TODO: Imports?? Check old implementation

	keys := make([]string, 0, len(schema.Types))
	for k := range schema.Types {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		_, err := f.WriteString(schema.Types[k])
		if err != nil {
			log.Error(err)
		}
	}

	return nil
}