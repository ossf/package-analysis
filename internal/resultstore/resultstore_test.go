package resultstore

import (
	"context"
	"fmt"
	"testing"
)

func TestFileBucket(t *testing.T) {
	tmpDir := t.TempDir()

	testBucketURL := "file://" + tmpDir
	fmt.Println(testBucketURL)

	testKeys := []string{
		"test1.txt",
		"testdir/test2.txt",
	}

	ctx := context.Background()

	rs := New(testBucketURL)
	if rs == nil {
		t.Errorf("failed to open create resultstore with URL %s (invalid url)", testBucketURL)
	}

	bucket, err := rs.openBucket(ctx)
	if err != nil {
		t.Errorf("failed to open bucket: %v", err)
	}

	for _, key := range testKeys {
		t.Run(key, func(t *testing.T) {
			writer, err := bucket.NewWriter(ctx, key, nil)
			if err != nil {
				t.Errorf("failed to create writer: %v", err)
			}

			if _, err := writer.Write([]byte("test bytes")); err != nil {
				t.Errorf("failed to write to file: %v", err)
			}

			if err := writer.Close(); err != nil {
				t.Errorf("failed to close writer: %v", err)
			}
		})
	}

	if err := bucket.Close(); err != nil {
		t.Errorf("failed to close bucket: %v", err)
	}
}
