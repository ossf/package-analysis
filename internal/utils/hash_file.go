package utils

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

// HashFile returns the SHA256 hashsum of a file.
// If prependHashType is true, the string "sha256:" is prepended
func HashFile(path string, prependHashType bool) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hash := sha256.New()
	if _, err = io.Copy(hash, f); err != nil {
		return "", err
	}

	digest := fmt.Sprintf("%x", hash.Sum([]byte{}))
	if prependHashType {
		digest = "sha256:" + digest
	}
	return digest, nil
}
