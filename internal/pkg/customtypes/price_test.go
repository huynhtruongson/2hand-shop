package customtypes

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestNewPrice(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string // decimal string representation of expected Price
		wantErr require.ErrorAssertionFunc
	}{
		// Valid cases
		{name: "integer", input: "10", want: "10", wantErr: nil},
		{name: "1 decimal place", input: "10.5", want: "10.5", wantErr: nil},
		{name: "2 decimal places", input: "10.99", want: "10.99", wantErr: nil},
		{name: "zero integer", input: "0", want: "0", wantErr: nil},
		{name: "zero 2dp", input: "0.00", want: "0", wantErr: nil},
		{name: "large value", input: "999999.99", want: "999999.99", wantErr: nil},
		{name: "whitespace trimmed", input: "  10.99  ", want: "10.99", wantErr: nil},
		{name: "scientific notation", input: "1e2", want: "100", wantErr: nil},

		// Error cases
		{name: "empty string", input: "", wantErr: require.Error},
		{name: "malformed", input: "abc", wantErr: require.Error},
		{name: "negative integer", input: "-1", wantErr: require.Error},
		{name: "negative decimal", input: "-1.00", wantErr: require.Error},
		{name: "more than 2 decimal places", input: "1.999", wantErr: require.Error},
		{name: "multiple dots", input: "1.2.3", wantErr: require.Error},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewPrice(tt.input)
			if tt.wantErr != nil {
				tt.wantErr(t, err)
				require.ErrorIs(t, err, ErrInvalidPriceFormat)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, p.String())
		})
	}
}

func TestMustNewPrice(t *testing.T) {
	t.Run("valid input no panic", func(t *testing.T) {
		require.NotPanics(t, func() {
			p := MustNewPrice("10.99")
			require.Equal(t, "10.99", p.String())
		})
	})

	t.Run("invalid input panics", func(t *testing.T) {
		require.Panics(t, func() {
			MustNewPrice("invalid")
		})
	})
}

func TestFromCents(t *testing.T) {
	tests := []struct {
		cents int64
		want  string
	}{
		{cents: 1099, want: "10.99"},
		{cents: 0, want: "0"},
		{cents: 100, want: "1"},
		{cents: 150, want: "1.5"},
		{cents: 1, want: "0.01"},
		{cents: 99999999, want: "999999.99"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("cents=%d", tt.cents), func(t *testing.T) {
			p := FromCents(tt.cents)
			require.Equal(t, tt.want, p.String())
		})
	}
}

func TestPrice_Cents(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{input: "10.99", want: 1099},
		{input: "0", want: 0},
		{input: "1.5", want: 150},
		{input: "1", want: 100},
		{input: "0.01", want: 1},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p := MustNewPrice(tt.input)
			require.Equal(t, tt.want, p.Cents())
		})
	}
}

func TestPrice_StripeRoundTrip(t *testing.T) {
	cases := []string{"0", "0.01", "0.99", "1.5", "10.99", "999999.99"}
	for _, s := range cases {
		p := MustNewPrice(s)
		cents := p.Cents()
		restored := FromCents(cents)
		require.True(t, p.Equal(restored), "round-trip failed for %s", s)
	}
}

func TestPrice_Arithmetic(t *testing.T) {
	p10 := MustNewPrice("10.00")
	p55 := MustNewPrice("5.50")
	p15 := MustNewPrice("1.50")

	require.Equal(t, "15.5", p10.Add(p55).String())
	require.Equal(t, "4.5", p10.Sub(p55).String())

	// Mul with decimal.Decimal
	result := p10.Mul(decimal.NewFromFloat(0.10))
	require.Equal(t, "1", result.String())

	// p - p = zero
	diff := p10.Sub(p10)
	require.True(t, diff.IsZero())
	require.Equal(t, "0", diff.String())

	// 10.00 + 1.50 = 11.50
	require.Equal(t, "11.5", p10.Add(p15).String())
}

func TestPrice_ValidationHelpers(t *testing.T) {
	t.Run("IsZero", func(t *testing.T) {
		require.True(t, MustNewPrice("0").IsZero())
		require.True(t, MustNewPrice("0.00").IsZero())
		require.False(t, MustNewPrice("0.01").IsZero())
		require.False(t, MustNewPrice("1").IsZero())
	})

	t.Run("IsPositive", func(t *testing.T) {
		require.True(t, MustNewPrice("0.01").IsPositive())
		require.True(t, MustNewPrice("1").IsPositive())
		require.False(t, MustNewPrice("0").IsPositive())
	})

	t.Run("GreaterThan", func(t *testing.T) {
		require.True(t, MustNewPrice("10.00").GreaterThan(MustNewPrice("9.99")))
		require.False(t, MustNewPrice("9.99").GreaterThan(MustNewPrice("10.00")))
		require.False(t, MustNewPrice("10.00").GreaterThan(MustNewPrice("10.00")))
	})

	t.Run("LessThan", func(t *testing.T) {
		require.True(t, MustNewPrice("9.99").LessThan(MustNewPrice("10.00")))
		require.False(t, MustNewPrice("10.00").LessThan(MustNewPrice("9.99")))
		require.False(t, MustNewPrice("10.00").LessThan(MustNewPrice("10.00")))
	})

	t.Run("Equal", func(t *testing.T) {
		require.True(t, MustNewPrice("10.00").Equal(MustNewPrice("10.00")))
		require.False(t, MustNewPrice("10.00").Equal(MustNewPrice("10.01")))
		// 1.50 and 1.5 are equal
		require.True(t, MustNewPrice("1.50").Equal(MustNewPrice("1.5")))
	})
}

func TestPrice_String_GoString(t *testing.T) {
	p := MustNewPrice("10.99")
	require.Equal(t, "10.99", p.String())

	// GoString roundtrips via MustNewPrice.
	gs := p.GoString()
	require.Equal(t, `customtypes.MustNewPrice("10.99")`, gs)
	restored := MustNewPrice("10.99")
	require.True(t, p.Equal(restored))
}

func TestPrice_MarshalJSON(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "10.99", want: `"10.99"`},
		{input: "0", want: `"0"`},
		{input: "1.5", want: `"1.5"`},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p := MustNewPrice(tt.input)
			data, err := p.MarshalJSON()
			require.NoError(t, err)
			require.Equal(t, tt.want, string(data))
		})
	}
}

func TestPrice_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr require.ErrorAssertionFunc
	}{
		{name: "quoted string", input: `"10.99"`, want: "10.99", wantErr: nil},
		{name: "quoted zero", input: `"0"`, want: "0", wantErr: nil},
		{name: "bare number", input: `10.99`, want: "10.99", wantErr: nil},
		{name: "bare zero", input: `0`, want: "0", wantErr: nil},
		{name: "bare float 1dp", input: `1.5`, want: "1.5", wantErr: nil},
		{name: "null", input: `null`, want: "0", wantErr: nil},
		{name: "invalid quoted", input: `"abc"`, wantErr: require.Error},
		{name: "invalid bare", input: `abc`, wantErr: require.Error},
		{name: "too many decimals quoted", input: `"1.999"`, wantErr: require.Error},
		{name: "negative quoted", input: `"-1"`, wantErr: require.Error},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Price
			err := p.UnmarshalJSON([]byte(tt.input))
			if tt.wantErr != nil {
				tt.wantErr(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, p.String())
		})
	}
}

func TestPrice_Scan(t *testing.T) {
	tests := []struct {
		name    string
		src     any
		want    string
		wantErr require.ErrorAssertionFunc
	}{
		{name: "nil", src: nil, want: "0", wantErr: nil},
		{name: "valid bytes", src: []byte("10.99"), want: "10.99", wantErr: nil},
		{name: "zero bytes", src: []byte("0"), want: "0", wantErr: nil},
		{name: "wrong type string", src: "10.99", wantErr: require.Error},
		{name: "wrong type int", src: 1099, wantErr: require.Error},
		{name: "invalid bytes", src: []byte("!!!"), wantErr: require.Error},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Price
			err := p.Scan(tt.src)
			if tt.wantErr != nil {
				tt.wantErr(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, p.String())
		})
	}
}

func TestPrice_Value(t *testing.T) {
	p := MustNewPrice("10.99")
	val, err := p.Value()
	require.NoError(t, err)

	b, ok := val.([]byte)
	require.True(t, ok, "Value() should return []byte")
	require.Equal(t, "10.99", string(b))
}

func TestPrice_DB_RoundTrip(t *testing.T) {
	cases := []string{"0", "0.01", "10.99", "999999.99"}
	for _, s := range cases {
		p := MustNewPrice(s)
		val, err := p.Value()
		require.NoError(t, err)

		var restored Price
		err = restored.Scan(val)
		require.NoError(t, err)
		require.True(t, p.Equal(restored), "DB round-trip failed for %s", s)
	}
}

func TestPrice_JSON_RoundTrip(t *testing.T) {
	cases := []string{"0", "0.01", "10.99", "999999.99", "1.5"}
	for _, s := range cases {
		p := MustNewPrice(s)
		data, err := json.Marshal(p)
		require.NoError(t, err)

		var restored Price
		err = json.Unmarshal(data, &restored)
		require.NoError(t, err)
		require.True(t, p.Equal(restored), "JSON round-trip failed for %s", s)
	}
}

func TestPrice_ImplementsInterfaces(t *testing.T) {
	// Compile-time assertions: if these types no longer implement the
	// interfaces, the program will not compile.
	var _ driver.Valuer = Price{}
	var _ interface{ Scan(any) error } = (*Price)(nil)
	var _ json.Marshaler   = Price{}
	var _ json.Unmarshaler = (*Price)(nil)
}
