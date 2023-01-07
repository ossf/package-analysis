package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

func GetSHA256Hash(content string) string {
	hash := sha256.New()
	hash.Write([]byte(content))
	return hex.EncodeToString(hash.Sum(nil))
}
