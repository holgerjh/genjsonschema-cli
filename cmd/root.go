package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

const envBinaryName = "GENSCHEMA_BINARY_NAME"
const defaultBinaryName = "genjsonschema-cli"

func RootCmd() *cobra.Command {
	binaryName := os.Getenv(envBinaryName)
	if binaryName == "" {
		binaryName = defaultBinaryName
	}
	command := &cobra.Command{
		Use:   binaryName,
		Short: "Generate JSON Schemas from one or more YAML or JSON files",
		Long: `This application is used to generate JSON Schemas from YAML or JSON files.
For more information, see create --help 
`,
	}
	command.AddCommand(
		generateCreateCommand(binaryName),
	)
	return command

}
