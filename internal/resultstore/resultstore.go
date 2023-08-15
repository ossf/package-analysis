package resultstore

import (
	"archive/zip"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"os"
	"path"
	"time"

	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/gcsblob"
	_ "gocloud.dev/blob/s3blob"

	"github.com/ossf/package-analysis/internal/log"
	"github.com/ossf/package-analysis/internal/pkgmanager"
	"github.com/ossf/package-analysis/internal/utils"
)

type ResultStore struct {
	bucket        *url.URL
	keyPrefix     string
	constructPath bool
}

type (
	Option interface{ set(*ResultStore) }
	option func(*ResultStore) // option implements Option.
)

func (o option) set(sb *ResultStore) { o(sb) }

// ConstructPath will cause Save() to append a suffix to the base path
// based on Pkg.EcosystemName() and Pkg.Name().
func ConstructPath() Option {
	return option(func(rs *ResultStore) { rs.constructPath = true })
}

// New creates a new ResultStore instance with the given bucket URL and options.
// If the bucket URL is invalid, a nil pointer is returned.
func New(bucket string, options ...Option) *ResultStore {
	bucketURL, err := url.Parse(bucket)
	if err != nil {
		return nil
	}

	var pathPrefix string

	if bucketURL.Scheme == "file" {
		// https://github.com/google/go-cloud/issues/3294
		params := bucketURL.Query()
		params.Set("no_tmp_dir", "true")
		bucketURL.RawQuery = params.Encode()
	} else {
		// for non-file schemes, we need to separate the path because OpenBucket
		// doesn't support turning the path into a key prefix
		pathPrefix = bucketURL.Path
		bucketURL.Path = ""
		bucketURL.RawPath = ""
	}

	rs := &ResultStore{
		bucket:    bucketURL,
		keyPrefix: pathPrefix,
	}

	for _, o := range options {
		o.set(rs)
	}
	return rs
}

func (rs *ResultStore) String() string {
	// label when bucket path is constructed from package name
	if rs.constructPath {
		return rs.bucket.JoinPath(rs.keyPrefix, "<dynamic path>").String()
	}

	return rs.bucket.JoinPath(rs.keyPrefix).String()
}

// generateKey creates an identifier key to store an object with.
// If p is non-nil and the ResultStore was constructed with
// the ConstructPath() option, then the base key will be prefixed
// with the ecosystem and name of the given package (in that order).
// Otherwise, the basename is returned.
func (rs *ResultStore) generateKey(p Pkg, baseKey string) string {
	if p != nil && rs.constructPath {
		return path.Join(rs.keyPrefix, p.EcosystemName(), p.Name(), baseKey)
	}
	return path.Join(rs.keyPrefix, baseKey)
}

func (rs *ResultStore) openBucket(ctx context.Context) (*blob.Bucket, error) {
	return blob.OpenBucket(ctx, rs.bucket.String())
}

func (rs *ResultStore) SaveTempFilesToZip(ctx context.Context, p Pkg, zipName string, tempFileNames []string) error {
	bkt, err := rs.openBucket(ctx)
	if err != nil {
		return err
	}
	defer bkt.Close()

	uploadPath := rs.generateKey(p, zipName+".zip")
	log.Info("Uploading results", "bucket", rs.bucket.String(), "path", uploadPath)

	bucketWriter, err := bkt.NewWriter(ctx, uploadPath, nil)
	if err != nil {
		return err
	}
	defer bucketWriter.Close()

	zipWriter := zip.NewWriter(bucketWriter)
	defer zipWriter.Close()

	for _, fileName := range tempFileNames {
		file, err := utils.OpenTempFile(fileName)
		if err != nil {
			return err
		}

		w, err := zipWriter.Create(fileName + ".json")
		if err != nil {
			return err
		}

		if _, err := io.Copy(w, file); err != nil {
			return err
		}

		if err = file.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (rs *ResultStore) SaveAnalyzedPackage(ctx context.Context, manager *pkgmanager.PkgManager, pkg Pkg) error {
	archivePath, err := manager.DownloadArchive(pkg.Name(), pkg.Version(), "")
	if err != nil {
		return err
	}

	defer func() {
		if err := os.Remove(archivePath); err != nil {
			log.Error("could not clean up downloaded archive", "error", err)
		}
	}()

	hash, err := utils.SHA256Hash(archivePath)
	if err != nil {
		return err
	}

	bkt, err := rs.openBucket(ctx)
	if err != nil {
		return err
	}
	defer bkt.Close()

	uploadPath := rs.generateKey(pkg, pkg.Version()+"-"+hash)
	log.Info("Uploading analyzed package", "bucket", rs.bucket.String(), "path", uploadPath)

	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	w, err := bkt.NewWriter(ctx, uploadPath, nil)
	if err != nil {
		return err
	}

	_, writeErr := io.Copy(w, f)
	closeErr := w.Close()

	if writeErr != nil {
		// TODO golang 1.20: use errors.Join(writeErr, closeErr)
		return writeErr
	}
	if closeErr != nil {
		return closeErr
	}

	return nil
}

// SaveWithFilename saves results to the bucket with the given filename
func (rs *ResultStore) SaveWithFilename(ctx context.Context, p Pkg, filename string, analysis any) error {
	if filename == "" {
		return errors.New("filename cannot be empty")
	}

	result := &result{
		Package: pkg{
			Name:      p.Name(),
			Ecosystem: p.EcosystemName(),
			Version:   p.Version(),
		},
		CreatedTimestamp: time.Now().UTC().Unix(),
		Analysis:         analysis,
	}

	b, err := json.Marshal(result)
	if err != nil {
		return err
	}

	bkt, err := rs.openBucket(ctx)
	if err != nil {
		return err
	}
	defer bkt.Close()

	uploadPath := rs.generateKey(p, filename)
	log.Info("Uploading results", "bucket", rs.bucket.String(), "path", uploadPath)

	w, err := bkt.NewWriter(ctx, uploadPath, nil)
	if err != nil {
		return err
	}
	if _, err := w.Write(b); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}

	return nil
}

// Save marshals the given analysis results to JSON and writes them to
// the bucket with a default filename (key). If p is non-nil and has a
// version specified, the default filename is <version>.json.
// Otherwise, it is "results.json".
func (rs *ResultStore) Save(ctx context.Context, p Pkg, analysis interface{}) error {
	filename := "results.json"
	if p != nil && p.Version() != "" {
		filename = p.Version() + ".json"
	}

	return rs.SaveWithFilename(ctx, p, filename, analysis)
}
