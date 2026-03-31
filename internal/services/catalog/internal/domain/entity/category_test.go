package entity

import (
	stderrors "errors"
	"testing"
	"time"

	apperrors "github.com/huynhtruongson/2hand-shop/internal/pkg/errors"
)

func TestNewCategory(t *testing.T) {
	t.Parallel()

	valid := []any{"cat-1", "Electronics", "All kinds of electronics", "", "", 1}

	tests := []struct {
		name      string
		mutate    func(a []any)
		wantErr   bool
		errDetail string
	}{
		{
			name:   "valid category",
			mutate: func(a []any) {},
			wantErr: false,
		},
		{
			name:      "empty id",
			mutate:    func(a []any) { a[0] = "" },
			wantErr:   true,
			errDetail: "id",
		},
		{
			name:      "empty name",
			mutate:    func(a []any) { a[1] = "" },
			wantErr:   true,
			errDetail: "name",
		},
		{
			name: "auto-generates slug from name when empty",
			mutate: func(a []any) { a[3] = "" },
			wantErr: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			args := make([]any, len(valid))
			copy(args, valid)
			tc.mutate(args)

			c, err := NewCategory(
				args[0].(string), args[1].(string),
				args[2].(string), args[3].(string),
				args[4].(string), args[5].(int),
			)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tc.errDetail != "" {
					var appErr *apperrors.AppError
					if stderrors.As(err, &appErr) {
						if _, ok := appErr.Details()[tc.errDetail]; !ok {
							t.Errorf("error details %v does not contain key %q", appErr.Details(), tc.errDetail)
						}
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if c == nil {
				t.Fatal("expected category, got nil")
			}
			// A newly created category has no deletedAt, so it is active.
			if c.DeletedAt() != nil {
				t.Error("expected new category to be active (deletedAt = nil)")
			}
		})
	}
}

func TestCategory_IsActive(t *testing.T) {
	t.Parallel()

	c, err := NewCategory("cat-1", "Electronics", "Desc", "", "", 1)
	if err != nil {
		t.Fatalf("NewCategory failed: %v", err)
	}

	// Newly created category is active (deletedAt is nil).
	if c.DeletedAt() != nil {
		t.Error("expected new category to be active")
	}
}

func TestCategory_Update(t *testing.T) {
	t.Parallel()

	c, err := NewCategory("cat-1", "Electronics", "Old desc", "", "", 1)
	if err != nil {
		t.Fatalf("NewCategory failed: %v", err)
	}

	err = c.Update("New Name", "New Desc")
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if c.Name() != "New Name" {
		t.Errorf("Name = %q, want %q", c.Name(), "New Name")
	}
	if c.Description() != "New Desc" {
		t.Errorf("Description = %q, want %q", c.Description(), "New Desc")
	}
}

func TestCategory_Update_EmptyName(t *testing.T) {
	t.Parallel()

	c, err := NewCategory("cat-1", "Electronics", "Desc", "", "", 1)
	if err != nil {
		t.Fatalf("NewCategory failed: %v", err)
	}

	err = c.Update("", "New Desc")
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestUnmarshalCategoryFromDB(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	deletedAt := now.Add(-time.Hour)
	c := UnmarshalCategoryFromDB(
		"id", "Name", "Desc", "slug", "icon.png",
		5, now, now, &deletedAt,
	)

	if c.ID() != "id" {
		t.Errorf("ID = %q, want %q", c.ID(), "id")
	}
	// deletedAt is set, so the category is inactive.
	if c.DeletedAt() == nil {
		t.Error("expected DeletedAt to be set")
	}
}

func TestGenerateSlug(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  string
	}{
		{"Electronics", "electronics"},
		{"Home & Garden", "home-garden"},
		{"Sports Equipment", "sports-equipment"},
		{"  Spaces  ", "spaces"},
		{"Kids & Baby", "kids-baby"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			got := generateSlug(tc.input)
			if got != tc.want {
				t.Errorf("generateSlug(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
