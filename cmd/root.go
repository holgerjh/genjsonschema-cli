package cmd

import (
	"github.com/spf13/cobra"
)

func RootCmd() *cobra.Command {
	command := &cobra.Command{
		Use:   "genjsonschema-cli",
		Short: "Generate JSON Schemas from one or more YAML or JSON files",
		Long: `This application is used to generate JSON Schemas from YAML or JSON files.
For more information, see genjsonschema-cli create --help 
`,
	}
	command.AddCommand(
		generateCreateCommand(),
	)
	return command

}
