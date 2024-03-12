package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// Hash creates a SHA256 hash from the passed string
// and key and returns the result in hex format.
func Hash(value, key []byte) (string, error) {
	h := hmac.New(sha256.New, key)
	_, err := h.Write(value)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
