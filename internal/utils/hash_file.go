package utils

import (
	"crypto/sha256"
	"fmt"
	"os"
)

func HashFile(path string) (string, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(bytes)
	return fmt.Sprintf("sha256:%x", hash), nil
}
