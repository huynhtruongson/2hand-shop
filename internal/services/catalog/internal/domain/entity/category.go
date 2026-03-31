package entity

import (
	"strings"
	"time"

	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/errors"
)

// Category represents a product category in the catalog bounded context.
// It is a domain entity: it has identity but its value fields may be compared.
// Categories form a flat hierarchy (no parent/child in this implementation).
type Category struct {
	id          string
	name        string
	description string
	slug        string
	iconURL     string
	sortOrder   int
	createdAt   time.Time
	updatedAt   time.Time
	deletedAt   *time.Time
}

// NewCategory constructs a new Category. It validates required fields and
// auto-generates a slug from the name if one is not supplied.
func NewCategory(id, name, description, slug, iconURL string, sortOrder int) (*Category, error) {
	c := &Category{
		id:          id,
		name:        name,
		description: description,
		iconURL:     iconURL,
		sortOrder:   sortOrder,
		createdAt:   time.Now().UTC(),
		updatedAt:   time.Now().UTC(),
	}

	// Auto-generate slug from name if slug is empty.
	if strings.TrimSpace(slug) == "" {
		c.slug = generateSlug(name)
	} else {
		c.slug = slug
	}

	if err := c.validate(); err != nil {
		return nil, err
	}
	return c, nil
}

// ── Getters ───────────────────────────────────────────────────────────────────

func (c *Category) ID() string           { return c.id }
func (c *Category) Name() string         { return c.name }
func (c *Category) Description() string  { return c.description }
func (c *Category) Slug() string         { return c.slug }
func (c *Category) IconURL() string      { return c.iconURL }
func (c *Category) SortOrder() int       { return c.sortOrder }
func (c *Category) CreatedAt() time.Time  { return c.createdAt }
func (c *Category) UpdatedAt() time.Time  { return c.updatedAt }
func (c *Category) DeletedAt() *time.Time { return c.deletedAt }

// ── Business logic ───────────────────────────────────────────────────────────

// Update updates the mutable name and description fields.
func (c *Category) Update(name, description string) error {
	c.name = name
	c.description = description
	c.updatedAt = time.Now().UTC()
	return c.validate()
}

// ── Factory / DB reconstruction ─────────────────────────────────────────────

// UnmarshalCategoryFromDB reconstructs a Category from persistence storage.
// Use this when loading from PostgreSQL; it skips validation so that stored
// (potentially legacy) data can still be loaded.
func UnmarshalCategoryFromDB(
	id, name, description, slug, iconURL string,
	sortOrder int,
	createdAt, updatedAt time.Time, deletedAt *time.Time,
) *Category {
	return &Category{
		id:          id,
		name:        name,
		description: description,
		slug:        slug,
		iconURL:     iconURL,
		sortOrder:   sortOrder,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
		deletedAt:   deletedAt,
	}
}

// ── Validation ───────────────────────────────────────────────────────────────

func (c *Category) validate() error {
	switch {
	case strings.TrimSpace(c.id) == "":
		return errors.ErrValidation.WithDetail("id", "id is empty")
	case strings.TrimSpace(c.name) == "":
		return errors.ErrValidation.WithDetail("name", "name is empty")
	case strings.TrimSpace(c.slug) == "":
		return errors.ErrValidation.WithDetail("slug", "slug is empty")
	}
	return nil
}

// generateSlug creates a URL-safe slug from a category name.
// It collapses consecutive hyphens and trims leading/trailing hyphens.
func generateSlug(name string) string {
	name = strings.ToLower(name)
	// Replace spaces and ampersands with hyphens.
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "&", "-")
	// Strip non-alphanumeric characters (keep hyphens).
	var result strings.Builder
	var prevHyphen bool
	for _, ch := range name {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') {
			result.WriteRune(ch)
			prevHyphen = false
		} else if ch == '-' {
			if !prevHyphen {
				result.WriteRune(ch)
				prevHyphen = true
			}
			// skip consecutive hyphens
		}
	}
	return strings.Trim(result.String(), "-")
}
