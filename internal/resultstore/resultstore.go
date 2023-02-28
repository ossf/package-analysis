package resultstore

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"time"

	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/gcsblob"
	_ "gocloud.dev/blob/s3blob"

	"github.com/ossf/package-analysis/internal/log"
)

const writeBufferFolder = "file-write-contents"

type ResultStore struct {
	bucket        string
	basePath      string
	constructPath bool
}

type (
	Option interface{ set(*ResultStore) }
	option func(*ResultStore) // option implements Option.
)

func (o option) set(sb *ResultStore) { o(sb) }

// ConstructPath will cause Save() to generate the path based on Pkg.EcosystemName()
// and Pkg.Name().
func ConstructPath() Option {
	return option(func(rs *ResultStore) { rs.constructPath = true })
}

// BasePath sets the base path used while saving files to storage.
func BasePath(base string) Option {
	return option(func(rs *ResultStore) { rs.basePath = base })
}

func New(bucket string, options ...Option) *ResultStore {
	rs := &ResultStore{
		bucket: bucket,
	}
	for _, o := range options {
		o.set(rs)
	}
	return rs
}

func (rs *ResultStore) generatePath(p Pkg) string {
	path := rs.basePath
	if rs.constructPath {
		path = filepath.Join(path, p.EcosystemName(), p.Name())
	}
	return path
}

func (rs *ResultStore) OpenAndWriteToBucket(ctx context.Context, contents []byte, path, filename string) error {
	bkt, err := blob.OpenBucket(ctx, rs.bucket)
	if err != nil {
		return err
	}
	defer bkt.Close()

	uploadPath := filepath.Join(path, filename)
	log.Info("Uploading results",
		"bucket", rs.bucket,
		"path", uploadPath)

	w, err := bkt.NewWriter(ctx, uploadPath, nil)
	if err != nil {
		return err
	}
	if _, err := w.Write(contents); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}

	return nil
}

func (rs *ResultStore) SaveWriteBuffer(ctx context.Context, p Pkg, fileName string, writeBuffer []byte) error {
	path := filepath.Join(rs.generatePath(p), writeBufferFolder)
	return rs.OpenAndWriteToBucket(ctx, writeBuffer, path, fileName+".json")
}

func (rs *ResultStore) SaveWriteBufferZip(ctx context.Context, p Pkg, fileName string, writeBufferZip *os.File) error {
	path := filepath.Join(rs.generatePath(p), writeBufferFolder)
	bkt, err := blob.OpenBucket(ctx, rs.bucket)
	if err != nil {
		return err
	}
	defer bkt.Close()

	uploadPath := filepath.Join(path, fileName+".zip")
	log.Info("Uploading results",
		"bucket", rs.bucket,
		"path", uploadPath)

	w, err := bkt.NewWriter(ctx, uploadPath, nil)
	if err != nil {
		return err
	}
	//zipWriter := zip.NewWriter(w)
	//archive, err := zip.OpenReader(writeBufferZip.Name())
	if err != nil {
		return err
	}
	//defer archive.Close()
	zipFile, err := os.Open(writeBufferZip.Name())
	io.Copy(w, zipFile)
	//for _, f := range archive.File {
	//	fileInArchive, err := f.Open()
	//	if err != nil {
	//		return err
	//	}
	//	if _, err := io.Copy(w, fileInArchive); err != nil {
	//		return err
	//	}
	//}
	//// Only 1 file, but should fix this. mabe pass in the zip reader with multiple files read in.
	//zipWriter.Copy(archive.File[0])
	//if err := zipWriter.Close(); err != nil {
	//	return err
	//}
	return nil
}

func (rs *ResultStore) Save(ctx context.Context, p Pkg, analysis interface{}) error {
	version := p.Version()
	result := &result{
		Package: pkg{
			Name:      p.Name(),
			Ecosystem: p.EcosystemName(),
			Version:   version,
		},
		CreatedTimestamp: time.Now().UTC().Unix(),
		Analysis:         analysis,
	}

	filename := "results.json"
	if version != "" {
		filename = version + ".json"
	}

	b, err := json.Marshal(result)
	if err != nil {
		return err
	}

	return rs.OpenAndWriteToBucket(ctx, b, rs.generatePath(p), filename)
}
