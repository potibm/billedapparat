package cmd

import (
	"github.com/spf13/cobra"
)

func NewImportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import commands",
		Annotations: map[string]string{
			skipConfigValidationAnnotation: "true",
		},
	}

	return cmd
}
