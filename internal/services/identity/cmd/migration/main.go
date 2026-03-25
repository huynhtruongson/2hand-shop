package main

import (
	"fmt"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/logger/zerolog"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/migration/gomigrate"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(upCmd)
	rootCmd.AddCommand(downCmd)
}

var rootCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Database migration CLI",
	Long:  "A CLI tool to run Postgres migrations up and down using golang-migrate.",
}

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Run migrations up",
	Long:  "Run migrations up",
	Run: func(cmd *cobra.Command, args []string) {
		executeMigration(gomigrate.MigrationTypeUp)
	},
}

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Run migrations down",
	Long:  "Run migrations down",
	Run: func(cmd *cobra.Command, args []string) {
		executeMigration(gomigrate.MigrationTypeDown)
	},
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
}

func executeMigration(migrationType gomigrate.MigrationType) {
	config, err := config.Load()
	if err != nil {
		panic(err)
	}
	logger := zerolog.NewZeroLogger(zerolog.Config{Level: config.Logger.Level, Environment: config.App.Environment, ServiceName: config.App.ServiceName})

	dataSource := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		config.Postgres.User,
		config.Postgres.Password,
		config.Postgres.Host,
		config.Postgres.Port,
		config.Postgres.DBName,
		config.Postgres.SSLMode,
	)
	migeationDir := "/migrations"
	migrator, err := gomigrate.NewGoMigrateMigrator(logger, dataSource, migeationDir)
	if err != nil {
		logger.Fatal("Failed to initialize migrator", "error", err)
	}

	if migrationType == gomigrate.MigrationTypeUp {
		if err := migrator.Up(); err != nil {
			logger.Fatal("Failed to migrate up", "error", err)
		}
	} else if migrationType == gomigrate.MigrationTypeDown {
		if err := migrator.Down(); err != nil {
			logger.Fatal("Failed to migrate down", "error", err)
		}
	}
}
