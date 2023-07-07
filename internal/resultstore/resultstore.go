package resultstore

import (
	"archive/zip"
	"context"
	"encoding/json"
	"errors"
	"io"
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
	bucket        string
	basePath      string
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

func (rs *ResultStore) String() string {
	s := rs.bucket + "/" + rs.basePath
	if rs.constructPath {
		s += "+"
	}
	return s
}

func (rs *ResultStore) generatePath(p Pkg) string {
	uploadPath := rs.basePath
	if rs.constructPath {
		uploadPath = path.Join(uploadPath, p.EcosystemName(), p.Name())
	}
	return uploadPath
}

func (rs *ResultStore) SaveTempFilesToZip(ctx context.Context, p Pkg, fileName string, tempFileNames []string) error {
	bkt, err := blob.OpenBucket(ctx, rs.bucket)
	if err != nil {
		return err
	}
	defer bkt.Close()

	uploadPath := path.Join(rs.generatePath(p), fileName+".zip")
	log.Info("Uploading results",
		"bucket", rs.bucket,
		"path", uploadPath)

	w, err := bkt.NewWriter(ctx, uploadPath, nil)
	if err != nil {
		return err
	}
	defer w.Close()

	zipWriter := zip.NewWriter(w)
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

	bkt, err := blob.OpenBucket(ctx, rs.bucket)
	if err != nil {
		return err
	}
	defer bkt.Close()

	uploadPath := rs.generatePath(pkg) + "-" + hash
	log.Info("Uploading analyzed package", "bucket", rs.bucket, "path", uploadPath)

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

	bkt, err := blob.OpenBucket(ctx, rs.bucket)
	if err != nil {
		return err
	}
	defer bkt.Close()

	uploadPath := path.Join(rs.generatePath(p), filename)
	log.Info("Uploading results",
		"bucket", rs.bucket,
		"path", uploadPath)

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

// MakeFilename returns the default filename to use for saving analysis results,
// using an optional label.
// If the package has a version, the default filename is
// "<label>-<version>.json" if label is nonempty, or <version>.json otherwise.
// If the package does not have a version specified, the default filename is
// "<label>.json" if label is nonempty, or "results.json" if not.
func MakeFilename(p Pkg, label string) string {
	prefix := "results"
	version := p.Version()

	if version != "" && label != "" {
		prefix = label + "-" + version
	} else if version != "" {
		prefix = version
	} else if label != "" {
		prefix = label
	}
	return prefix + ".json"

}

// Save saves the results with the default filename
func (rs *ResultStore) Save(ctx context.Context, p Pkg, analysis interface{}) error {
	filename := MakeFilename(p, "")
	return rs.SaveWithFilename(ctx, p, filename, analysis)
}
