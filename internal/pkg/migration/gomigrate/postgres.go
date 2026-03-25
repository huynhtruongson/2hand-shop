package gomigrate

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/logger"
)

type GoMigrateMigrator struct {
	logger        logger.Logger
	migrate       *migrate.Migrate
	dataSource    string
	migrationsDir string
}

type MigrationType string

const (
	MigrationTypeUp   MigrationType = "up"
	MigrationTypeDown MigrationType = "down"
)

func NewGoMigrateMigrator(logger logger.Logger, dataSource string, migrationsDir string) (*GoMigrateMigrator, error) {
	migrate, err := migrate.New(fmt.Sprintf("file://%s", migrationsDir), dataSource)
	if err != nil {
		return nil, err
	}
	return &GoMigrateMigrator{
		logger:        logger,
		migrate:       migrate,
		dataSource:    dataSource,
		migrationsDir: migrationsDir,
	}, nil
}

func (m *GoMigrateMigrator) Up() error {
	m.logger.Info("Migrating up")
	if err := m.migrate.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			m.logger.Info("No migrations to apply — already up to date.")
			return nil
		}
		return fmt.Errorf("migration up failed: %w", err)
	}
	v, dirty, _ := m.migrate.Version()
	m.logger.Info("Migrated up.", "Current version:", v, "(dirty:", dirty, ")")
	return nil
}

func (m *GoMigrateMigrator) Down() error {
	m.logger.Info("Migrating down")
	if err := m.migrate.Down(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			m.logger.Info("No migrations to apply — already up to date.")
			return nil
		}
		return fmt.Errorf("migration down failed: %w", err)
	}
	v, dirty, _ := m.migrate.Version()
	m.logger.Info("Migrated down.", "Current version:", v, "(dirty:", dirty, ")")
	return nil
}
