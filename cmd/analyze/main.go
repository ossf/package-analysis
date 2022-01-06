package main

import (
	"context"
	"flag"
	"net/url"
	"os"

	"github.com/ossf/package-analysis/internal/analysis"
	"github.com/ossf/package-analysis/internal/log"
	"github.com/ossf/package-analysis/internal/pkgecosystem"
	"github.com/ossf/package-analysis/internal/resultstore"
	"github.com/ossf/package-analysis/internal/sandbox"
)

var (
	pkgName   = flag.String("package", "", "package name")
	localPkg  = flag.String("local", "", "local package path")
	ecosystem = flag.String("ecosystem", "", "ecosystem (npm, pypi, or rubygems)")
	version   = flag.String("version", "", "version")
	upload    = flag.String("upload", "", "bucket path for uploading results")
	noPull    = flag.Bool("nopull", false, "disables pulling down sandbox images")
	imageTag  = flag.String("image-tag", "", "set a image tag")
)

func parseBucketPath(path string) (string, string) {
	parsed, err := url.Parse(path)
	if err != nil {
		log.Panic("Failed to parse bucket path",
			"path", path)
	}

	return parsed.Scheme + "://" + parsed.Host, parsed.Path
}

func main() {
	log.Initalize(os.Getenv("LOGGER_ENV"))
	flag.Parse()
	if *ecosystem == "" {
		flag.Usage()
		return
	}

	manager := pkgecosystem.Manager(*ecosystem)
	if manager == nil {
		log.Panic("Unsupported pkg manager",
			log.Label("ecosystem", *ecosystem))
	}

	if *pkgName == "" {
		flag.Usage()
		return
	}

	log.Info("Got request",
		log.Label("ecosystem", *ecosystem),
		log.Label("name", *pkgName),
		log.Label("localPath", *localPkg),
		log.Label("version", *version))

	var pkg *pkgecosystem.Pkg
	var err error
	if *localPkg != "" {
		if *version != "" {
			log.Panic("Unable to specify version for local packages")
		}
		pkg = manager.Local(*pkgName, *version, *localPkg)
	} else if *version != "" {
		pkg = manager.Package(*pkgName, *version)
	} else {
		pkg, err = manager.Latest(*pkgName)
		if err != nil {
			log.Panic("Failed to get latest version",
				log.Label("ecosystem", *ecosystem),
				log.Label("name", *pkgName))
		}
	}

	log.Info("Got request",
		log.Label("ecosystem", *ecosystem),
		log.Label("name", *pkgName),
		log.Label("version", *version))

	// Prepare the sandbox:
	// - Always pass through the tag. An empty tag is the same as "latest".
	// - Respect the "-nopull" option.
	// - Ensure any local package is mapped through.
	sbOpts := []sandbox.Option{
		sandbox.Tag(*imageTag),
	}
	if *noPull {
		sbOpts = append(sbOpts, sandbox.NoPull())
	}
	if *localPkg != "" {
		sbOpts = append(sbOpts, sandbox.Volume(*localPkg, *localPkg))
	}

	sb := sandbox.New(manager.Image(), sbOpts...)
	if err := sb.Start(); err != nil {
		log.Panic("start failed", "error", err)
	}
	defer sb.Stop()
	results := make(map[string]*analysis.Result)
	for _, phase := range manager.DynamicPhases() {
		result := analysis.Run(sb, pkg.Command(phase))
		results[phase] = result
	}

	ctx := context.Background()
	if *upload != "" {
		bucket, path := parseBucketPath(*upload)
		err := resultstore.New(bucket, resultstore.BasePath(path)).Save(ctx, pkg, results)
		if err != nil {
			log.Fatal("Failed to upload results to blobstore",
				"error", err)
		}
	}
}
