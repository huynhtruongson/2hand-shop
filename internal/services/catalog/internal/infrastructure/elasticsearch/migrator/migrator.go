package migrator

import (
	"context"
	"fmt"
	"sort"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/logger"
)

// Migration represents a single versioned Elasticsearch migration. Versions
// are zero-padded sortable strings (e.g. "001", "002") sorted in ascending
// lexicographic order.
//
// The Up function should be idempotent: if the change it represents is already
// present, it should succeed without error so that re-running a migration after
// a crash is safe.
//
// The Down function is not called automatically on failure; it must be invoked
// manually via Migrator.Down.
type Migration interface {
	// Version is the zero-padded version identifier, e.g. "001". It is used as
	// the document _id in the .migrations index, making it also the sort key.
	Version() string
	// Description is a human-readable description of what this migration does.
	Description() string
	// Up applies the migration. It receives the raw *elasticsearch.Client so that
	// migrations have full access to the ES API.
	Up(ctx context.Context, es *elasticsearch.Client) error
	// Down rolls back the migration. It should be the inverse of Up.
	Down(ctx context.Context, es *elasticsearch.Client) error
}

// MigrationStatus represents the current state of a single migration version.
type MigrationStatus struct {
	Version     string
	Description string
	Applied     bool
}

// Migrator runs versioned migrations against an Elasticsearch cluster. It is
// safe to re-run: any version already recorded in the .migrations index is
// skipped. If Up fails partway through, Migrator does not roll back already-
// applied migrations — that must be done manually via Down.
type Migrator struct {
	es              *elasticsearch.Client
	lg              logger.Logger
	migrationsIndex string
	appliedStore    *AppliedStore
	registered      []Migration
}

// NewMigrator constructs a Migrator. The migrationsIndex parameter names the
// Elasticsearch index used to track applied versions; if empty, ".migrations"
// is used. Optional migrations may be passed at construction; additional
// migrations can be registered later via Register.
func NewMigrator(es *elasticsearch.Client, lg logger.Logger, migrationsIndex string, registered ...Migration) *Migrator {
	m := &Migrator{
		es:              es,
		lg:              lg,
		migrationsIndex: migrationsIndex,
		registered:      registered,
	}
	m.appliedStore = NewAppliedStore(es, migrationsIndex, lg)
	return m
}

// Register appends migrations to the migrator's registry. They are sorted by
// Version string in ascending order each time Register is called, so migrations
// can be registered in any order. It is safe to call Register before or after
// NewMigrator.
func (m *Migrator) Register(ms ...Migration) {
	m.registered = append(m.registered, ms...)
	sort.Slice(m.registered, func(i, j int) bool {
		return m.registered[i].Version() < m.registered[j].Version()
	})
}

// Run executes all pending Up migrations in ascending version order. A migration
// is considered pending if its version is not present in the .migrations index.
// If Up returns an error, Run stops immediately and returns that error — it
// does not roll back already-applied migrations.
func (m *Migrator) Run(ctx context.Context) error {
	if err := m.appliedStore.Ensure(ctx); err != nil {
		return fmt.Errorf("ensure migrations index: %w", err)
	}

	applied, err := m.appliedStore.Applied(ctx)
	if err != nil {
		return fmt.Errorf("load applied migrations: %w", err)
	}

	for _, mig := range m.registered {
		if applied[mig.Version()] {
			m.lg.Debug("migration already applied, skipping", "version", mig.Version)
			continue
		}
		m.lg.Info("applying migration", "version", mig.Version, "description", mig.Description)
		if err := mig.Up(ctx, m.es); err != nil {
			return fmt.Errorf("migration %s up failed: %w", mig.Version(), err)
		}
		if err := m.appliedStore.Record(ctx, mig); err != nil {
			return fmt.Errorf("record migration %s: %w", mig.Version(), err)
		}
		m.lg.Info("migration applied", "version", mig.Version)
	}

	return nil
}

// Down rolls back the last n applied migrations in descending version order.
// It returns an error if any Down function fails; already-rolled-back migrations
// are not re-applied. If n is zero or negative, Down is a no-op.
func (m *Migrator) Down(ctx context.Context, steps int) error {
	if steps <= 0 {
		return nil
	}

	applied, err := m.appliedStore.Applied(ctx)
	if err != nil {
		return fmt.Errorf("load applied migrations: %w", err)
	}

	// Sort descending so we roll back the newest migration first.
	sorted := make([]Migration, len(m.registered))
	copy(sorted, m.registered)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Version() > sorted[j].Version()
	})

	rolledBack := 0
	for _, mig := range sorted {
		if !applied[mig.Version()] {
			continue
		}
		if rolledBack >= steps {
			break
		}
		m.lg.Info("rolling back migration", "version", mig.Version(), "description", mig.Description())
		if err := mig.Down(ctx, m.es); err != nil {
			return fmt.Errorf("migration %s down failed: %w", mig.Version(), err)
		}
		if err := m.appliedStore.Remove(ctx, mig.Version()); err != nil {
			return fmt.Errorf("remove migration record %s: %w", mig.Version(), err)
		}
		rolledBack++
	}

	return nil
}

// Target migrates up or down to reach the specified version. It determines the
// current applied version (the highest applied), then applies or rolls back as
// needed. If targetVersion equals the current version, Target is a no-op.
// If targetVersion is not found in the registered migrations, Target returns
// an error.
func (m *Migrator) Target(ctx context.Context, targetVersion string) error {
	applied, err := m.appliedStore.Applied(ctx)
	if err != nil {
		return fmt.Errorf("load applied migrations: %w", err)
	}

	// Determine current version (highest applied).
	current := "000"
	for _, mig := range m.registered {
		if applied[mig.Version()] && mig.Version() > current {
			current = mig.Version()
		}
	}

	if targetVersion == current {
		m.lg.Info("already at target version", "version", targetVersion)
		return nil
	}

	// Check that targetVersion exists in our registry.
	found := false
	for _, mig := range m.registered {
		if mig.Version() == targetVersion {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("target version %q not found in migration registry", targetVersion)
	}

	if targetVersion > current {
		// Apply all pending versions up to and including targetVersion.
		for _, mig := range m.registered {
			if mig.Version() > current && mig.Version() <= targetVersion {
				m.lg.Info("applying migration", "version", mig.Version(), "description", mig.Description())
				if err := mig.Up(ctx, m.es); err != nil {
					return fmt.Errorf("migration %s up failed: %w", mig.Version(), err)
				}
				if err := m.appliedStore.Record(ctx, mig); err != nil {
					return fmt.Errorf("record migration %s: %w", mig.Version(), err)
				}
			}
		}
	} else {
		// Roll back from current down to (but not including) targetVersion.
		sorted := make([]Migration, len(m.registered))
		copy(sorted, m.registered)
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Version() > sorted[j].Version()
		})
		for _, mig := range sorted {
			if mig.Version() <= current && mig.Version() > targetVersion {
				m.lg.Info("rolling back migration", "version", mig.Version(), "description", mig.Description())
				if err := mig.Down(ctx, m.es); err != nil {
					return fmt.Errorf("migration %s down failed: %w", mig.Version(), err)
				}
				if err := m.appliedStore.Remove(ctx, mig.Version()); err != nil {
					return fmt.Errorf("remove migration record %s: %w", mig.Version(), err)
				}
			}
		}
	}

	return nil
}

// Status returns the status of every registered migration.
func (m *Migrator) Status(ctx context.Context) ([]MigrationStatus, error) {
	applied, err := m.appliedStore.Applied(ctx)
	if err != nil {
		return nil, fmt.Errorf("load applied migrations: %w", err)
	}

	out := make([]MigrationStatus, 0, len(m.registered))
	for _, mig := range m.registered {
		out = append(out, MigrationStatus{
			Version:     mig.Version(),
			Description: mig.Description(),
			Applied:     applied[mig.Version()],
		})
	}
	return out, nil
}
