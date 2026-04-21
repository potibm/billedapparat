package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/potibm/billedapparat/internal/app/config"
	"github.com/potibm/billedapparat/internal/app/seeder"
	store "github.com/potibm/billedapparat/internal/app/store/gorm"
	"github.com/spf13/cobra"
)

var (
	resetSeed  bool
	resetForce bool
)

func NewDatabaseResetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset database and media files",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ensureAppInfrastructure()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if !resetForce && !confirmReset(Cfg.App.DbFilename) {
				slog.Info("Reset aborted by user")

				return nil
			}

			err := performDatabaseReset(Cfg.App.DbFilename)
			if err != nil {
				slog.Error("Database reset failed", "error", err)

				return fmt.Errorf("database reset failed: %w", err)
			}

			err = performMediaReset(config.MediaDirname)
			if err != nil {
				slog.Error("Media reset failed", "error", err)

				return fmt.Errorf("media reset failed: %w", err)
			}

			if resetSeed {
				if err := seedDatabase(); err != nil {
					slog.Error("Database seeding failed after reset", "error", err)

					return fmt.Errorf("database seeding failed after reset: %w", err)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&resetSeed, "seed", "s", false, "Runs a seed after the reset")
	cmd.Flags().BoolVarP(&resetForce, "force", "f", false, "Skips the confirmation prompt (for CI/CD)")

	return cmd
}

func confirmReset(dbName string) bool {
	return confirm(fmt.Sprintf("WARNING: The database '%s' and all media files will be COMPLETELY deleted!", dbName))
}

func performDatabaseReset(dbName string) error {
	slog.Info("Performing database reset...")

	dbStore, err := store.NewSqliteStore(dbName)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	defer func() {
		if err := dbStore.Close(); err != nil {
			slog.Error("failed to close database", "error", err)
		}
	}()

	if err := dbStore.PurgeAll(); err != nil {
		return fmt.Errorf("failed to purge database: %w", err)
	}

	slog.Info("Database reset completed successfully!")

	return nil
}

func performMediaReset(directory string) error {
	slog.Info("Performing media reset...")

	if err := os.RemoveAll(directory); err != nil {
		return fmt.Errorf("failed to delete media directory: %w", err)
	}

	slog.Info("Media reset completed successfully!")

	return nil
}

func seedDatabase() error {
	slog.Info("Starting database seeding...")

	dbStore, err := store.NewSqliteStore(Cfg.App.DbFilename)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	defer func() {
		if err := dbStore.Close(); err != nil {
			slog.Error("failed to close database", "error", err)
		}
	}()

	s := seeder.NewSeeder(dbStore.NewSlideRepository())
	if err := s.Run(); err != nil {
		return fmt.Errorf("seeding error: %w", err)
	}

	slog.Info("Database seeding completed successfully!")

	return nil
}
