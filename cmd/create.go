package cmd

import (
	"fmt"
	"os"

	"github.com/holgerjh/genjsonschema"
	"github.com/holgerjh/genjsonschema-cli/internal/createschema"
	"github.com/spf13/cobra"
)

func generateCreateCommand() *cobra.Command {
	app := &createschema.CreateSchemaApp{}

	files := []string{}

	command := &cobra.Command{
		Use:   "create file",
		Short: "Creates a JSON Schema from one or multiple YAML and/or JSON file(s)",
		Long: `
	This command creates a JSON Schema from one or multiple YAML and/or JSON file(s).
	YAML files are only supported insofar they have an equivalent JSON representation.
	Among others, this means they must only contain mappings with string keys.

	List length and order is not preserved and datatypes encountered during schema generation
	are allowed to occur an undefined number of times. This e.g. means that a schema generated
	from '[1, "foo"]' accepts '[]', '[1, 2]' and ["bar", 5, 6, "baz"].

	Example:
	  Generate a schema from "example.yaml" and write it to STDOUT:
	    genjsonschema example.yaml

	  Generate a schema from "example.yaml" that requires all object properties to be set
	  and that disallows additional object properties. Store it in "out.yaml":
	  	genjsonschema -o out.yaml -r -a example.yaml


	To read from STDIN, specify "-" as filename.
		Example:
		  echo '{"foo": "bar"}' | genjsonschema create -

	Use -f to specify additional input files. Files are merged together as follows:

		* Objects are deeply merged.
		* Lists are merged constructively. 
		
		Example:
		  Given:
		  	file1: {"foo": "aaa", "bar": [41], "baz": [41]}
		  	file2: {"foo": "bbb",              "baz": [42]}
		  Then "genjsonschema create -f file2 file1" effectively works on the following input:
		  	       {"foo": "bbb", "bar": [41], "baz": [41, 42]}
		
		Merging is only supported for the same datatypes.
		Example:
		  * strings: "foo" with "bar" gives "bar"
		  * lists: [11] with ["foo"] gives [11, "foo"]
		  * objects: {"foo": "bar"} with {"bar": "baz"} gives {"foo": "bar", "bar": "baz"}
		  * error: [42] with {"foo": "bar"} gives an error
		
		Note that this holds for deeper structures as well:
		  Given
			file1: {"foo": 42}
			file2: {"foo": {"bar": "baz"}}
		  then "genjsonschema create -f file2 file1" fails with an error (42 and type object cannot be merged)
`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return parseArguments(cmd, args, files, app)
		},

		Run: func(cmd *cobra.Command, args []string) {
			if err := app.Run(); err != nil {
				fmt.Printf("Encountered an error: %v", err)
				os.Exit(1)
			}
		}}

	command.Flags().StringP("output", "o", "", "Output file. Default is STDOUT.")
	command.Flags().StringP("id", "d", "", "Fill the schema $id field.")
	command.Flags().BoolP("require-all", "r", false, "Generates a schema that requires all object properties to be set. Default: false")
	command.Flags().BoolP("allow-additional", "a", false, "Generates a schema that allows unknown object properties that were not encountered during schema generation. Default: false")
	command.Flags().BoolP("merge-only", "m", false, "Do not generate a schema. Instead, output the YAML result of the merge operation. Default: false")
	command.Flags().StringArrayVarP(&files, "file", "f", []string{}, "Additional file that will be merged into main file before creating the schema. Can be specified mulitple times.")

	return command

}

func parseArguments(cmd *cobra.Command, args []string, files []string, app *createschema.CreateSchemaApp) error {
	if len(args) != 1 {
		return fmt.Errorf("you need to specify exactly one input file. See --help for detailed information")
	}
	schemaConfig, err := schemaConfigFromCmd(cmd)
	if err != nil {
		return fmt.Errorf("unexpected error parsing command line: %v", err)
	}
	inputFiles := append(args, files...)
	outFile, err := cmd.Flags().GetString("output")
	if err != nil {
		return fmt.Errorf("unexpected error parsing command line: %v", err)
	}
	mergeOnly, err := cmd.Flags().GetBool("merge-only")
	if err != nil {
		return fmt.Errorf("unexpected error parsing command line: %v", err)
	}
	app.Arguments = &createschema.Arguments{
		SchemaConfig: *schemaConfig,
		InputFiles:   inputFiles,
		OutputFile:   outFile,
		MergeOnly:    mergeOnly,
	}
	return nil
}

func schemaConfigFromCmd(cmd *cobra.Command) (*genjsonschema.SchemaConfig, error) {
	id, err := cmd.Flags().GetString("id")
	if err != nil {
		return nil, err
	}
	additionalProperties, err := cmd.Flags().GetBool("allow-additional")
	if err != nil {
		return nil, err
	}
	requireAllProperties, err := cmd.Flags().GetBool("require-all")
	if err != nil {
		return nil, err
	}

	return genjsonschema.NewSchemaConfig(id, additionalProperties, requireAllProperties), nil
}
