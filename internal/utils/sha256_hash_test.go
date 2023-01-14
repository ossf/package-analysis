package utils

import (
	"testing"
)

// Test multiple hash queries to make sure the hashes are computed as expected.
func TestMultipleHashQueries(t *testing.T) {
	firstTestString := "First test string"
	secondTestString := "Second test string"
	firstExpectedHash := "bdb0b089f806e5fc8e71198c965d351c5ff28ff6ea6e5d8fff31bf55c7267b25"
	secondExpectedHash := "51fa0cdbe10324f18ac4d123eb29a402be2f3b9c5623a12ea013ec7e5fbb655e"
	firstHash := GetSHA256Hash(firstTestString)
	if firstHash != firstExpectedHash {
		t.Errorf(`Expected = %v, Actual  = %v`, firstExpectedHash, firstHash)
	}
	secondHash := GetSHA256Hash(secondTestString)
	if secondHash != secondExpectedHash {
		t.Errorf(`Expected = %v, Actual  = %v`, secondExpectedHash, secondHash)
	}
}
