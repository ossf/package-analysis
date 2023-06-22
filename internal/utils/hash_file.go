package utils

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
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

// RenameWithHash renames the file at the given path to the SHA256 digest
// of its contents, with an optional prefix and suffix (which may be empty).
func RenameWithHash(path, prefix, suffix string) (string, error) {
	hashsum, hashErr := HashFile(path, false)
	if hashErr != nil {
		return "", hashErr
	}

	dir := filepath.Dir(path)
	newPath := filepath.Join(dir, prefix+hashsum+suffix)

	if err := os.Rename(path, newPath); err != nil {
		return "", err
	}
	return newPath, nil
}
