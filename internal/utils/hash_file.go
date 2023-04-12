package utils

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

func HashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	hash := sha256.New()
	_, err = io.Copy(hash, f)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("sha256:%x", hash.Sum([]byte{})), nil
}
