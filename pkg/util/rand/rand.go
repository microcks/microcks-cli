package rand

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// String generates, from the set of capital and lowercase letters, a cryptographically-secure pseudo-random string of a given length.
func String(n int) (string, error) {
	return StringFromCharset(n, letterBytes)
}

// StringFromCharset generates, from a given charset, a cryptographically-secure pseudo-random string of a given length.
func StringFromCharset(n int, charset string) (string, error) {
	if n<0{
		return "", fmt.Errorf("n must be greater than 0")
	}
	if n > 0 && len(charset) == 0 {
		return "", fmt.Errorf("charset must not be empty when n > 0")
	}
	b := make([]byte, n)
	maxIdx := big.NewInt(int64(len(charset)))
	for i := 0; i < n; i++ {
		randIdx, err := rand.Int(rand.Reader, maxIdx)
		if err != nil {
			return "", fmt.Errorf("failed to generate random string: %w", err)
		}
		randIdxInt := int(randIdx.Int64())
		b[i] = charset[randIdxInt]
	}
	return string(b), nil
}
