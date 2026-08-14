package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/potibm/billedapparat/internal/app/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	importDirectory string
	importForce     bool
)

func NewImportSlidesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "slides",
		Short: "Import slides from a directory",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ensureAppInfrastructure()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			slog.Info("Importing slides from directory", "directory", importDirectory)

			if !importForce &&
				!confirm("This will import slides from the specified directory. No duplicate check. Continue?") {
				slog.Info("Import aborted by user")

				return nil
			}

			hubURL := viper.GetString("app.collector_url")
			if hubURL == "" {
				return fmt.Errorf("app.collector_url must be set in the configuration")
			}

			err := triggerImport(hubURL, Cfg.API.AdminAPIKey, importDirectory)
			if err != nil {
				return fmt.Errorf("import failed: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().
		StringVarP(&importDirectory, "directory", "d", "",
			fmt.Sprintf("Directory to import slides from, relative to %s", config.ImportDirname))

	cmd.Flags().BoolVarP(&importForce, "force", "f", false, "Skips the confirmation prompt")

	return cmd
}

func generateImportURL(hubURL string) string {
	return fmt.Sprintf("%s/api/internal/import", strings.TrimRight(hubURL, "/"))
}

func triggerImport(hubURL, apiKey, directory string) error {
	url := generateImportURL(hubURL)

	slog.Debug("Sending import request to server", "url", url, "directory", directory)

	payload := map[string]string{"directory": directory}

	bodyData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("network error during request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return fmt.Errorf("failed to read response body: %w", readErr)
	}

	if resp.StatusCode != http.StatusOK {
		slog.Error("Import API returned an error.",
			"status_code", resp.StatusCode,
			"response", string(bodyBytes))

		return fmt.Errorf("import failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	slog.Info("Import successful!", "status", resp.Status, "response", string(bodyBytes))

	return nil
}
