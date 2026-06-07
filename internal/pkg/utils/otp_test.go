package utils

import (
	"errors"
	"testing"
	"unicode"

	"github.com/stretchr/testify/require"
)

func TestGenerateOTP(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{name: "four digits", length: 4},
		{name: "six digits", length: 6},
		{name: "eight digits", length: 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			otp, err := GenerateOTP(tt.length)

			require.NoError(t, err)
			require.Len(t, otp, tt.length)
			for _, r := range otp {
				require.True(t, unicode.IsDigit(r))
			}
		})
	}
}

func TestGenerateOTP_InvalidLength(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{name: "zero", length: 0},
		{name: "negative", length: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			otp, err := GenerateOTP(tt.length)

			require.Empty(t, otp)
			require.True(t, errors.Is(err, ErrInvalidOTPLength))
		})
	}
}
