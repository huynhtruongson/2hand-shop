// Command es-migration runs Elasticsearch index migrations for the catalog service.
//
// Usage:
//
//	es-migration up       run all pending migrations
//	es-migration down [n] roll back the last n migrations (default 1)
//	es-migration status   show applied/pending status for every migration
//	es-migration target v migrate to a specific version v
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/logger"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/logger/zerolog"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/config"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/infrastructure/elasticsearch"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/infrastructure/elasticsearch/migrations"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/infrastructure/elasticsearch/migrator"
	"github.com/spf13/cobra"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

var (
	cfg config.Config
	log logger.Logger

	upCmd = &cobra.Command{
		Use:   "up",
		Short: "Run all pending migrations",
		Long:  "Run all pending Elasticsearch index migrations in ascending version order.",
		Run:   runUp,
	}

	downCmd = &cobra.Command{
		Use:   "down [n]",
		Short: "Roll back migrations",
		Long:  "Roll back the last n applied migrations (default 1).",
		Args:  cobra.MaximumNArgs(1),
		Run:   runDown,
	}

	statusCmd = &cobra.Command{
		Use:   "status",
		Short: "Show migration status",
		Long:  "Print the applied/pending state of every registered migration.",
		Run:   runStatus,
	}

	targetCmd = &cobra.Command{
		Use:   "target <version>",
		Short: "Migrate to a specific version",
		Long:  "Migrate up or down to reach the specified version.",
		Args:  cobra.ExactArgs(1),
		Run:   runTarget,
	}

	rootCmd = &cobra.Command{
		Use:   "es-migration",
		Short: "Elasticsearch index migration CLI for the catalog service",
		Long:  "Manages Elasticsearch index schema migrations for the catalog service.",
	}
)

func init() {
	rootCmd.AddCommand(upCmd)
	rootCmd.AddCommand(downCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(targetCmd)

	// Load config and initialise logger before any command runs.
	var err error
	cfg, err = config.Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	log = zerolog.NewZeroLogger(zerolog.Config{
		Level:       cfg.Logger.Level,
		Environment: cfg.App.Environment,
		ServiceName: cfg.App.ServiceName,
	})
}

func buildMigrator() *migrator.Migrator {
	esClient, err := elasticsearch.NewClient(cfg.Elasticsearch, log)
	if err != nil {
		log.Fatal("failed to connect to Elasticsearch", "error", err)
	}
	if esClient == nil {
		log.Fatal("Elasticsearch is unavailable")
	}

	m := migrator.NewMigrator(esClient.Elasticsearch(), log, "")
	m.Register(migrations.NewCreateProductIndexMigration())
	return m
}

func runUp(_ *cobra.Command, _ []string) {
	m := buildMigrator()
	if err := m.Run(context.Background()); err != nil {
		log.Fatal("migration up failed", "error", err)
	}
	log.Info("all migrations applied successfully")
}

func runDown(cmd *cobra.Command, args []string) {
	steps := 1
	if len(args) == 1 {
		var n int
		if _, err := fmt.Sscanf(args[0], "%d", &n); err != nil || n <= 0 {
			log.Fatal("invalid steps argument, must be a positive integer")
		}
		steps = n
	}

	m := buildMigrator()
	if err := m.Down(context.Background(), steps); err != nil {
		log.Fatal("migration down failed", "error", err)
	}
	log.Info("migration down completed")
}

func runStatus(_ *cobra.Command, _ []string) {
	m := buildMigrator()
	statuses, err := m.Status(context.Background())
	if err != nil {
		log.Fatal("failed to get migration status", "error", err)
	}

	if len(statuses) == 0 {
		log.Info("no migrations registered")
		return
	}

	_, _ = fmt.Fprintln(os.Stdout, "VERSION  DESCRIPTION            STATUS")
	_, _ = fmt.Fprintln(os.Stdout, "-------- ---------------------- ------------")
	for _, s := range statuses {
		status := "pending"
		if s.Applied {
			status = "applied"
		}
		_, _ = fmt.Fprintf(os.Stdout, "%-8s %-22s %s\n", s.Version, s.Description, status)
	}
}

func runTarget(_ *cobra.Command, args []string) {
	targetVersion := args[0]
	m := buildMigrator()
	if err := m.Target(context.Background(), targetVersion); err != nil {
		log.Fatal("migration target failed", "error", err)
	}
	log.Info("reached target version", "version", targetVersion)
}
