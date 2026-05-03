package cmd

import (
	"github.com/spf13/cobra"
)

func NewDatabaseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "database",
		Short: "Database commands",
		Annotations: map[string]string{
			skipConfigValidationAnnotation: "true",
		},
	}

	return cmd
}
