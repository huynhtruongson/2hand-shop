package utils

import (
	"crypto/rand"
	"errors"
	"math/big"
)

var ErrInvalidOTPLength = errors.New("otp length must be greater than zero")

func GenerateOTP(length int) (string, error) {
	if length <= 0 {
		return "", ErrInvalidOTPLength
	}

	const digits = "0123456789"
	otp := make([]byte, length)
	max := big.NewInt(int64(len(digits)))

	for i := range otp {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		otp[i] = digits[n.Int64()]
	}

	return string(otp), nil
}
