package utils

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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

/*
BasenameWithHash computes the SHA256 digest of the file at the given path
and produces a new filename (basename) by inserting the digest into the
current filename just before the file extension, if present.

The file extension is defined as everything after and including the first
'.' character in the basename. If there is no '.' character, the digest is
simply appended to the filename. Note, this definition is different from the
one used in filepath.Ext(); this way allows for extensions such as .tar.gz

The prefix and suffix are extra strings which are concatenated with the
digest before the result is added to the basename. They may be left blank.

If an error occurs during hashing, it is returned along with an empty path.
*/
func BasenameWithHash(path, prefix, suffix string) (string, error) {
	digest, hashErr := HashFile(path, false)
	if hashErr != nil {
		return "", hashErr
	}
	hashString := prefix + digest + suffix

	// check for extension
	basename := filepath.Base(path)
	if extIndex := strings.IndexByte(basename, '.'); extIndex >= 0 {
		return basename[0:extIndex] + hashString + basename[extIndex:], nil
	}

	// else no extension, just append
	return basename + hashString, nil
}
