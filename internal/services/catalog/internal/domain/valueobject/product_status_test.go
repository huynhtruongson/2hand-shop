package valueobject

import "testing"

func TestProductStatus_CanTransitionTo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		from   ProductStatus
		to     ProductStatus
		expect bool
	}{
		// draft → published (allowed)
		{ProductStatusDraft, ProductStatusPublished, true},
		// draft → sold (not allowed)
		{ProductStatusDraft, ProductStatusSold, false},
		// draft → archived (not allowed)
		{ProductStatusDraft, ProductStatusArchived, false},
		// published → sold (allowed)
		{ProductStatusPublished, ProductStatusSold, true},
		// published → archived (allowed)
		{ProductStatusPublished, ProductStatusArchived, true},
		// published → draft (not allowed — backward transition)
		{ProductStatusPublished, ProductStatusDraft, false},
		// sold is terminal
		{ProductStatusSold, ProductStatusDraft, false},
		{ProductStatusSold, ProductStatusPublished, false},
		{ProductStatusSold, ProductStatusSold, false},
		{ProductStatusSold, ProductStatusArchived, false},
		// archived is terminal
		{ProductStatusArchived, ProductStatusDraft, false},
		{ProductStatusArchived, ProductStatusPublished, false},
		{ProductStatusArchived, ProductStatusSold, false},
		{ProductStatusArchived, ProductStatusArchived, false},
	}

	for _, tc := range tests {
		t.Run(tc.from.String()+"_to_"+tc.to.String(), func(t *testing.T) {
			t.Parallel()
			got := tc.from.CanTransitionTo(tc.to)
			if got != tc.expect {
				t.Errorf("CanTransitionTo(%s) from %s = %v, want %v",
					tc.to, tc.from, got, tc.expect)
			}
		})
	}
}

func TestAllProductStatuses(t *testing.T) {
	t.Parallel()

	all := AllProductStatuses()
	if len(all) != 4 {
		t.Errorf("AllProductStatuses returned %d items, want 4", len(all))
	}

	for _, s := range all {
		r, err := NewProductStatusFromString(s.String())
		if err != nil {
			t.Errorf("NewProductStatusFromString(%q) failed: %v", s.String(), err)
		}
		if r != s {
			t.Errorf("round-trip status %s != %s", r, s)
		}
	}
}

func TestProductStatus_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		s    ProductStatus
		want string
	}{
		{ProductStatusDraft, "draft"},
		{ProductStatusPublished, "published"},
		{ProductStatusSold, "sold"},
		{ProductStatusArchived, "archived"},
	}

	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			t.Parallel()
			if got := tc.s.String(); got != tc.want {
				t.Errorf("String() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestNewProductStatusFromString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input   string
		wantErr bool
	}{
		{"draft", false},
		{"published", false},
		{"sold", false},
		{"archived", false},
		{"invalid", true},
		{"", true},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			_, err := NewProductStatusFromString(tc.input)
			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
