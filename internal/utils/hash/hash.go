package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func Hash(value, key []byte) (string, error) {
	h := hmac.New(sha256.New, key)
	_, err := h.Write(value)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
