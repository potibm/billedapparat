package gorm

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/potibm/billedapparat/internal/app/config"
)

// maxOpenConns caps the SQLite connection pool, see configureConnectionPool.
const maxOpenConns = 4

type Store struct {
	db *gorm.DB
}

var allModels = []interface{}{
	&dbSlide{},
	&dbFilterRule{},
	&dbNews{},
	&dbTimetableEvent{},
}

func NewSqliteStore(filename string) (*Store, error) {
	if filename == "" {
		filename = config.DefaultDBFilename
	}

	dbPath := filepath.Join(config.DatabaseDirname, filename+".db")

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

// Ping reports whether the database connection is alive. It is a connection-level
// check only, on purpose; see doc/architecture/007-health-readiness-endpoints.md.
func (s *Store) Ping(ctx context.Context) error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying database connection: %w", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
}

func (s *Store) PurgeAll() error {
	if err := s.db.Migrator().DropTable(allModels...); err != nil {
		return fmt.Errorf("failed to drop tables: %w", err)
	}

	if err := s.db.AutoMigrate(allModels...); err != nil {
		return fmt.Errorf("failed to recreate tables: %w", err)
	}

	return nil
}

func newStore(dsn string) (*Store, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying database connection: %w", err)
	}

	if _, err := sqlDB.Exec("PRAGMA journal_mode = WAL;"); err != nil {
		return nil, fmt.Errorf("failed to set journal mode: %w", err)
	}

	if _, err := sqlDB.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		return nil, fmt.Errorf("failed to set foreign keys: %w", err)
	}

	configureConnectionPool(sqlDB)

	if err := db.AutoMigrate(allModels...); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	if err := cleanupSoftDeletes(db); err != nil {
		return nil, fmt.Errorf("failed to cleanup soft deleted records: %w", err)
	}

	err = RegisterAuditCallbacks(db)
	if err != nil {
		return nil, fmt.Errorf("failed to register audit callbacks: %w", err)
	}

	return &Store{db: db}, nil
}

// configureConnectionPool bounds the pool, see ADR 007. Kept above one so a slow
// query cannot occupy the only connection and stall the probe.
func configureConnectionPool(sqlDB *sql.DB) {
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxOpenConns)
}

func cleanupSoftDeletes(db *gorm.DB) error {
	if err := db.Unscoped().Where("deleted_at IS NOT NULL").Delete(&dbNews{}).Error; err != nil {
		return fmt.Errorf("failed to hard delete news: %w", err)
	}

	if err := db.Unscoped().Where("deleted_at IS NOT NULL").Delete(&dbTimetableEvent{}).Error; err != nil {
		return fmt.Errorf("failed to hard delete timetable: %w", err)
	}

	query := "(type = 'news' OR type = 'timetable') AND deleted_at IS NOT NULL"
	if err := db.Unscoped().Where(query).Delete(&dbSlide{}).Error; err != nil {
		return fmt.Errorf("failed to hard delete slides: %w", err)
	}

	return nil
}
