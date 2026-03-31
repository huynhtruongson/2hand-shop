package utils

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
