package cmd

import (
	"github.com/spf13/cobra"
)

var importPort int

func NewImportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import commands",
		Annotations: map[string]string{
			skipConfigValidationAnnotation: "true",
		},
	}

	cmd.PersistentFlags().
		IntVarP(&importPort, "port", "p", defaultPort, "Set the port number where the server to listens on")

	return cmd
}
