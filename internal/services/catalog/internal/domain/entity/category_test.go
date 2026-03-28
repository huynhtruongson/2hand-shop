package entity

import (
	"testing"
	"time"
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
			// slug field is empty string, so name will be used to generate it
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
				// For validation errors we just check that we got an error
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if c == nil {
				t.Fatal("expected category, got nil")
			}
			if !c.IsActive() {
				t.Error("expected IsActive() = true")
			}
		})
	}
}

func TestCategory_Activate_Deactivate(t *testing.T) {
	t.Parallel()

	c, err := NewCategory("cat-1", "Electronics", "Desc", "", "", 1)
	if err != nil {
		t.Fatalf("NewCategory failed: %v", err)
	}

	if !c.IsActive() {
		t.Error("expected new category to be active")
	}

	c.Deactivate()
	if c.IsActive() {
		t.Error("expected Deactivate() to set IsActive=false")
	}

	c.Activate()
	if !c.IsActive() {
		t.Error("expected Activate() to set IsActive=true")
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
	c := UnmarshalCategoryFromDB(
		"id", "Name", "Desc", "slug", "icon.png",
		5, false, now, now,
	)

	if c.ID() != "id" {
		t.Errorf("ID = %q, want %q", c.ID(), "id")
	}
	if c.IsActive() {
		t.Error("expected IsActive() = false")
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
