package helpers

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func EncodeBase62(n uint64) string {
	if n == 0 {
		return "0"
	}

	buf := make([]byte, 0, 11)
	for n > 0 {
		buf = append(buf, base62Chars[n%62])
		n /= 62
	}
	for i, j := 0, len(buf) - 1 ; i < j ; i, j = i + 1 , j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}

	return string(buf)
}
