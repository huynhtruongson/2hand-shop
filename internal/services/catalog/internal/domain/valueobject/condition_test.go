package valueobject

import "testing"

func TestCondition_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value    string
		expected bool
	}{
		{"new", true},
		{"like_new", true},
		{"good", true},
		{"fair", true},
		{"poor", true},
		{"invalid", false},
		{"", false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.value, func(t *testing.T) {
			t.Parallel()
			c, err := NewConditionFromString(tc.value)
			if tc.expected {
				if err != nil {
					t.Errorf("NewConditionFromString(%q) returned error: %v", tc.value, err)
				}
				if !c.IsValid() {
					t.Errorf("Condition(%q).IsValid() = false, want true", tc.value)
				}
			} else {
				if err == nil {
					t.Errorf("NewConditionFromString(%q) expected error, got nil", tc.value)
				}
			}
		})
	}
}

func TestAllConditions(t *testing.T) {
	t.Parallel()

	all := AllConditions()
	if len(all) != 5 {
		t.Errorf("AllConditions returned %d items, want 5", len(all))
	}

	// Verify all returned conditions are valid.
	for _, c := range all {
		if !c.IsValid() {
			t.Errorf("AllConditions contained invalid condition: %s", c)
		}
	}

	// Verify each condition round-trips correctly.
	for _, c := range all {
		r, err := NewConditionFromString(c.String())
		if err != nil {
			t.Errorf("NewConditionFromString(%q) failed: %v", c.String(), err)
		}
		if r != c {
			t.Errorf("round-trip condition %s != %s", r, c)
		}
	}
}

func TestCondition_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		c    Condition
		want string
	}{
		{ConditionNew, "new"},
		{ConditionLikeNew, "like_new"},
		{ConditionGood, "good"},
		{ConditionFair, "fair"},
		{ConditionPoor, "poor"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.want, func(t *testing.T) {
			t.Parallel()
			if got := tc.c.String(); got != tc.want {
				t.Errorf("String() = %q, want %q", got, tc.want)
			}
		})
	}
}
