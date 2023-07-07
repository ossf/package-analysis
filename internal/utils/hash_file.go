package utils

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

// SHA256Hash returns the SHA256 hashsum of a file.
func SHA256Hash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hash := sha256.New()
	if _, err = io.Copy(hash, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum([]byte{})), nil
}
