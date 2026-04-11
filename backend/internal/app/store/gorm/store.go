package gorm

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/potibm/billedapparat/internal/app/config"
)

const (
	defaultDirectory = "data/db"
	defaultDirMode   = 0o755
)

type Store struct {
	db *gorm.DB
}

var allModels = []interface{}{
	&dbSlide{},
}

func NewSqliteStore(filename string) (*Store, error) {
	if filename == "" {
		filename = config.DefaultDBFilename
	}

	dbPath := filepath.Join(defaultDirectory, filename+".db")
	if err := os.MkdirAll(defaultDirectory, defaultDirMode); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	dsn := fmt.Sprintf("%s?_busy_timeout=5000", dbPath)

	return newStore(dsn)
}

func NewSqliteInMemoryStore() (*Store, error) {
	dsn := "file::memory:?cache=shared"

	return newStore(dsn)
}

func (s *Store) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying database connection: %w", err)
	}

	return sqlDB.Close()
}

func newStore(dsn string) (*Store, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err == nil {
		_, err = sqlDB.Exec("PRAGMA journal_mode = WAL;")
		if err != nil {
			return nil, fmt.Errorf("failed to set journal mode: %w", err)
		}

		_, err = sqlDB.Exec("PRAGMA foreign_keys = ON;")
		if err != nil {
			return nil, fmt.Errorf("failed to set foreign keys: %w", err)
		}
	}

	if err := db.AutoMigrate(allModels...); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return &Store{db: db}, nil
}
