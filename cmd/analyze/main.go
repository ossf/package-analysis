package main

import (
	"context"
	"flag"
	"net/url"
	"os"

	"github.com/ossf/package-analysis/internal/analysis"
	"github.com/ossf/package-analysis/internal/log"
	"github.com/ossf/package-analysis/internal/pkgecosystem"
	"github.com/ossf/package-analysis/internal/sandbox"
)

var (
	pkg       = flag.String("package", "", "package name")
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

	manager, ok := pkgecosystem.SupportedPkgManagers[*ecosystem]
	if !ok {
		log.Panic("Unsupported pkg manager",
			log.Label("ecosystem", *ecosystem))
	}

	if *pkg == "" {
		flag.Usage()
		return
	}

	live := true
	if *localPkg == "" {
		if *version == "" {
			*version = manager.GetLatest(*pkg)
		}
	} else {
		live = false
		if *version != "" {
			log.Panic("Unable to specify version for local packages")
		}
	}

	log.Info("Got request",
		log.Label("ecosystem", *ecosystem),
		log.Label("name", *pkg),
		log.Label("version", *version),
		"live", live)

	args := manager.Args("all", *pkg, *version, *localPkg)

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
	if !live {
		sbOpts = append(sbOpts, sandbox.Volume(*localPkg, *localPkg))
	}

	sb := sandbox.New(manager.Image, sbOpts...)
	result := analysis.Run(*ecosystem, *pkg, *version, sb, args)

	ctx := context.Background()
	if *upload != "" {
		bucket, path := parseBucketPath(*upload)
		err := analysis.UploadResults(ctx, bucket, path, result)
		if err != nil {
			log.Fatal("Failed to upload results to blobstore",
				"error", err)
		}
	}
}
