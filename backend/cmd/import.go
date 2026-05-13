package cmd

import (
	"github.com/potibm/billedapparat/internal/app/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
		IntVarP(&importPort, portFlagName, "p", config.DefaultPort, "Set the port number where the server to listens on")
	_ = viper.BindPFlag("app.port", cmd.Flags().Lookup(portFlagName))

	return cmd
}
