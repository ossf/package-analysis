package resultstore

import (
	"context"
	"encoding/json"
	"path/filepath"
	"time"

	"github.com/ossf/package-analysis/internal/log"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/gcsblob"
	_ "gocloud.dev/blob/s3blob"
)

type ResultStore struct {
	bucket        string
	basePath      string
	constructPath bool
}

type Option interface{ set(*ResultStore) }
type option func(*ResultStore)       // option implements Option.
func (o option) set(sb *ResultStore) { o(sb) }

// ConstructPath will cause Save() to generate the path based on Pkg.Ecosystem()
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
		path = filepath.Join(path, p.Ecosystem(), p.Name())
	}
	return path
}

func (rs *ResultStore) Save(ctx context.Context, p Pkg, analysis interface{}) error {
	version := p.Version()
	result := &result{
		Package: pkg{
			Name:      p.Name(),
			Ecosystem: p.Ecosystem(),
			Version:   version,
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

	filename := "results.json"
	if version != "" {
		filename = version + ".json"
	}

	path := rs.generatePath(p)
	uploadPath := filepath.Join(path, filename)
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
