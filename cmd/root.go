package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func RootCmd() *cobra.Command {
	command := &cobra.Command{
		Use:   "genjsonschema-cli",
		Short: "Generate JSON Schemas from one or more YAML or JSON files",
		Long: `This application is used to generate JSON Schemas from YAML or JSON files.
For more information, see genjsonschema-cli create --help 
`,
		// Uncomment the following line if your bare application
		// has an action associated with it:
		// Run: func(cmd *cobra.Command, args []string) { },
	}
	command.AddCommand(
		generateCreateCommand(),
	)
	return command

}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := RootCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}
