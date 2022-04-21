package cmd

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/holgerjh/genjsonschema"
	merge "github.com/holgerjh/genjsonschema-cli/internal/merge"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create file1 [file2 [...]]",
	Short: "Creates a JSON Schema from one or multiple YAML and/or JSON file(s)",
	Long: `
	This command creates a JSON Schema from one or multiple YAML and/or JSON file(s).
	YAML files are only supported insofar they have an equivalent JSON representation,
	this e.g. means they must not contain mappings where the key is a number.
	
	To read from STDIN, specify "-" as filename (only allowed once).
		Example:
		  echo '{"foo": "bar"}' | genjsonschema create -

	If more than one input file is specified, they are merged together as follows:

		* Objects are deeply merged. If keys conflict, then the values from latter files overwrites those of previous files.
		* Lists are always constructively merged. List order is not preserved. 
		
		Example:
		  Given:
		  	file1: {"foo": "aaa", "bar": [41], "baz": [41]}
		  	file2: {"foo": "bbb",              "baz": [42]}
		  Then "genjsonschema create file1 file2" effectively works on the following input:
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
		  then "genjsonschema create file1 file2" fails with an error (42 and type object cannot be merged)
		
		

`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Fprintf(os.Stderr, "You need to specify at least one input file.")
			os.Exit(2)
		}

		result, err := CreateSchemaFromFiles(args...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create schema: %s", err)
			os.Exit(1)
		}

		outFile, _ := cmd.Flags().GetString("output")
		if outFile == "" {
			fmt.Fprintf(os.Stdout, "%s", result)
		} else {
			if err := os.WriteFile(outFile, result, fs.ModePerm); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to write result into output file: %s", err)
				os.Exit(120)
			}
		}

	},
}

func CreateSchemaFromFiles(files ...string) ([]byte, error) {
	merged, err := LoadAndMergeFiles(files...)
	if err != nil {
		return nil, err
	}
	b, err := yaml.Marshal(merged)
	if err != nil {
		return nil, err
	}
	return genjsonschema.GenerateFromYAML(b, nil)
}

// LoadAndMergeFiles
func LoadAndMergeFiles(files ...string) (interface{}, error) {
	loadedFiles, err := loadAllFiles(files...)
	if err != nil {
		return nil, err
	}
	merged, err := merge.MergeAllYAML(loadedFiles...)
	if err != nil {
		return nil, err
	}
	return merged, nil
}

func loadAllFiles(files ...string) ([][]byte, error) {
	loadedFiles := make([][]byte, len(files))
	for i, v := range files {
		var file *os.File = nil
		var handle io.Reader
		var err error = nil
		if v == "-" {
			handle = bufio.NewReader(os.Stdin)
		} else {
			file, err = os.Open(v)
			if err != nil {
				return nil, err
			}
			handle = file
		}
		loadedFiles[i], err = ioutil.ReadAll(handle)
		if file != nil {
			file.Close()
		}
		if err != nil {
			return nil, err
		}
	}
	return loadedFiles, nil
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().StringP("output", "o", "", "Output file. Default is STDOUT.")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
