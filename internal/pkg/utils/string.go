package utils

import "time"

// strconv.Itoa is not used to avoid importing strconv package just for this function
func Itoa(n int) string {
	if n == 0 {
		return "0"
	}
	const digits = "0123456789"
	var buf [20]byte // fixed-size buffer (big enough for int64 max: 9223372036854775807)
	i := len(buf)
	for x := n; x > 0; x /= 10 {
		i--
		buf[i] = digits[x%10] // append least-significant digit
	}
	return string(buf[i:]) // slice from first filled position
}

func GenerateRefNumber(prefix string) string {
	const refChars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	ts := time.Now().UTC().UnixNano()
	b := make([]byte, 8)
	for i := 0; i < 8; i++ {
		b[i] = refChars[ts%36]
		ts /= 36
	}
	return prefix + "-" + string(b)
}
