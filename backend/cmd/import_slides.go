package cmd

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/potibm/billedapparat/internal/app/config"
	"github.com/spf13/cobra"
)

var (
	importDirectory string
	importForce     bool
)

func NewImportSlidesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "slides",
		Short: "Import slides from a directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			slog.Info("Importing slides from directory", "directory", importDirectory)

			if !importForce &&
				!confirm("This will import slides from the specified directory. No duplicate check. Continue?") {
				slog.Info("Import aborted by user")

				return nil
			}

			url := fmt.Sprintf("http://localhost:%d/api/internal/import", importPort)
			slog.Debug("Sending import request to server", "url", url, "directory", importDirectory)

			req, _ := http.NewRequest(http.MethodPost, url, strings.NewReader(`{"directory":"`+importDirectory+`"}`))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-API-Key", Cfg.API.AdminAPIKey)

			client := &http.Client{}
			resp, err := client.Do(req)

			if err != nil || resp.StatusCode != http.StatusOK {
				slog.Error("Import failed.", "error", err, "statusCode", resp.StatusCode)

				return fmt.Errorf("import failed: %w", err)
			}

			slog.Info("Import successful!", "status", resp.Status, "response", resp.Body)

			return nil
		},
	}

	cmd.Flags().
		StringVarP(&importDirectory, "directory", "d", "",
			fmt.Sprintf("Directory to import slides from, relative to %s", config.ImportDirname))

	cmd.Flags().BoolVarP(&importForce, "force", "f", false, "Skips the confirmation prompt")

	return cmd
}
