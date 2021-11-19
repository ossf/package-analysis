package analysis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/ossf/package-analysis/internal/sandbox"
	"github.com/ossf/package-analysis/internal/strace"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/gcsblob"
	_ "gocloud.dev/blob/s3blob"

	"gocloud.dev/docstore"
	_ "gocloud.dev/docstore/gcpfirestore"
	_ "gocloud.dev/docstore/mongodocstore"
)

type fileResult struct {
	Path  string
	Read  bool
	Write bool
}

type socketResult struct {
	Address string
	Port    int
}

type commandResult struct {
	Command     []string
	Environment []string
}

type Package struct {
	Ecosystem string
	Name      string
	Version   string
}

type AnalysisResult struct {
	Package  Package
	Files    []fileResult
	Sockets  []socketResult
	Commands []commandResult
}

type DocstoreIndex struct {
	ID      string
	Package Package
	Indexes []string
}

const (
	maxIndexEntries = 10000
)

func RunLocal(ecosystem, pkgPath, version, image, command string) *AnalysisResult {
	return run(ecosystem, pkgPath, version, image, command, []string{
		"-v", fmt.Sprintf("%s:%s", pkgPath, pkgPath),
	})
}

func RunLive(ecosystem, pkgName, version, image, command string) *AnalysisResult {
	return run(ecosystem, pkgName, version, image, command, nil)
}

func run(ecosystem, pkgName, version, image, command string, args []string) *AnalysisResult {
	log.Printf("Running analysis using %s %s", command, args)

	// Init the sandbox
	sb, err := sandbox.Init(image)
	if err != nil {
		log.Panic(err)
	}

	// Run the command
	r, err := sb.Run(command, args...)

	if err != nil {
		log.Panic(err)
	}

	// Grab the log file
	l, err := r.Log()
	if err != nil {
		log.Panic(err)
	}
	defer l.Close()

	straceResult, err := strace.Parse(l)
	if err != nil {
		log.Panic(err)
	}

	result := AnalysisResult{}
	result.setData(ecosystem, pkgName, version, straceResult)
	return &result
}

func (d *AnalysisResult) setData(ecosystem, pkgName, version string, straceResult *strace.Result) {
	d.Package.Ecosystem = ecosystem
	d.Package.Name = pkgName
	d.Package.Version = version

	for _, f := range straceResult.Files() {
		d.Files = append(d.Files, fileResult{
			Path:  f.Path,
			Read:  f.Read,
			Write: f.Write,
		})
	}
	for _, s := range straceResult.Sockets() {
		d.Sockets = append(d.Sockets, socketResult{
			Address: s.Address,
			Port:    s.Port,
		})
	}
	for _, c := range straceResult.Commands() {
		d.Commands = append(d.Commands, commandResult{
			Command:     c.Command,
			Environment: c.Env,
		})
	}

}

func generateDocstoreName(pkg Package) string {
	id := fmt.Sprintf("%s:%s:%s", pkg.Ecosystem, pkg.Name, pkg.Version)
	id = strings.ReplaceAll(id, "/", "\\")
	return id
}

func generateIndexEntries(pkg Package, indexValues []string) []*DocstoreIndex {
	var entries []*DocstoreIndex
	for i := 0; i < len(indexValues); i += maxIndexEntries {
		endIdx := i + maxIndexEntries
		if endIdx > len(indexValues) {
			endIdx = len(indexValues)
		}

		entry := &DocstoreIndex{
			ID:      fmt.Sprintf("%s-%d", generateDocstoreName(pkg), i/maxIndexEntries),
			Package: pkg,
			Indexes: indexValues[i:endIdx],
		}
		entries = append(entries, entry)
	}
	return entries
}

func (r *AnalysisResult) GenerateFileIndexes() []*DocstoreIndex {
	fileParts := map[string]bool{}
	for _, f := range r.Files {
		cur := f.Path
		for cur != "/" && cur != "." {
			name := filepath.Base(cur)
			fileParts[name] = true
			cur = filepath.Dir(cur)
		}
	}

	var parts []string
	for part, _ := range fileParts {
		parts = append(parts, part)
	}

	return generateIndexEntries(r.Package, parts)
}

func (r *AnalysisResult) GenerateSocketIndexes() []*DocstoreIndex {
	var parts []string
	for _, socket := range r.Sockets {
		parts = append(parts, fmt.Sprintf("%s-%d", socket.Address, socket.Port))
		parts = append(parts, socket.Address)
	}
	return generateIndexEntries(r.Package, parts)
}

func (r *AnalysisResult) GenerateCmdIndexes() []*DocstoreIndex {
	// Index command components.
	cmdParts := map[string]bool{}
	for _, cmd := range r.Commands {
		for _, part := range cmd.Command {
			cmdParts[part] = true
		}
	}
	var parts []string
	for part, _ := range cmdParts {
		parts = append(parts, part)
	}
	return generateIndexEntries(r.Package, parts)
}

func UploadResults(ctx context.Context, bucket, path string, result *AnalysisResult) error {
	b, err := json.Marshal(result)
	if err != nil {
		return err
	}

	bkt, err := blob.OpenBucket(ctx, bucket)
	if err != nil {
		return err
	}
	defer bkt.Close()

	filename := "results.json"
	if result.Package.Version != "" {
		filename = result.Package.Version + ".json"
	}

	uploadPath := filepath.Join(path, filename)
	log.Printf("uploading to bucket=%s, path=%s", bucket, uploadPath)

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

func writeIndexes(ctx context.Context, collectionPath string, indexes []*DocstoreIndex) error {
	coll, err := docstore.OpenCollection(ctx, collectionPath)
	if err != nil {
		return err
	}
	defer coll.Close()

	actionList := coll.Actions()
	for _, index := range indexes {
		actionList.Put(index)
	}
	return actionList.Do(ctx)
}

func buildCollectionPath(prefix, name string) (string, error) {
	if strings.HasPrefix(prefix, "firestore://") {
		return prefix + name + "?name_field=ID", nil
	} else if strings.HasPrefix(prefix, "mongo://") {
		return prefix + name + "?id_field=ID", nil
	} else {
		return "", fmt.Errorf("unknown docstore collection path prefix: %v", prefix)
	}
}

func WriteResultsToDocstore(ctx context.Context, collectionPrefix string, result *AnalysisResult) error {
	files := result.GenerateFileIndexes()
	filesPath, err := buildCollectionPath(collectionPrefix, "files")
	if err != nil {
		return err
	}
	if err := writeIndexes(ctx, filesPath, files); err != nil {
		return err
	}

	sockets := result.GenerateSocketIndexes()
	socketsPath, err := buildCollectionPath(collectionPrefix, "sockets")
	if err != nil {
		return err
	}
	if err := writeIndexes(ctx, socketsPath, sockets); err != nil {
		return err
	}

	cmds := result.GenerateCmdIndexes()
	cmdsPath, err := buildCollectionPath(collectionPrefix, "commands")
	if err != nil {
		return err
	}
	return writeIndexes(ctx, cmdsPath, cmds)
}
