package main

import (
	"context"
	"flag"
	"net/url"

	"github.com/ossf/package-analysis/internal/analysis"
	"github.com/ossf/package-analysis/internal/log"
	"github.com/ossf/package-analysis/internal/pkgecosystem"
	"github.com/ossf/package-analysis/internal/sandbox"
)

var (
	pkg       = flag.String("package", "", "live package name")
	localPkg  = flag.String("local", "", "local package path")
	ecosystem = flag.String("ecosystem", "", "ecosystem (npm, pypi, or rubygems)")
	version   = flag.String("version", "", "version")
	upload    = flag.String("upload", "", "bucket path for uploading results")
	noPull    = flag.Bool("nopull", false, "disables pulling down sandbox images")
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
	log.Initalize(false)
	flag.Parse()
	if *ecosystem == "" {
		flag.Usage()
		return
	}

	manager, ok := pkgecosystem.SupportedPkgManagers[*ecosystem]
	if !ok {
		log.Panic("Unsupported pkg manager",
			"ecosystem", ecosystem)
	}

	var pkgName string
	live := true
	if *pkg != "" {
		pkgName = *pkg
		if *version == "" {
			*version = manager.GetLatest(pkgName)
		}
	} else if *localPkg != "" {
		pkgName = *localPkg
		if *version != "" {
			log.Panic("Unable to specify version for local packages")
		}
		live = false
	} else {
		flag.Usage()
		return
	}

	log.Info("Got request",
		"ecosystem", *ecosystem,
		"name", pkgName,
		"version", *version,
		"live", live)

	command := manager.CommandFmt(pkgName, *version)

	// Prepare the sandbox to use ensuring we respect the -"nopull" option.
	sbOpts := make([]sandbox.Option, 0)
	if *noPull {
		sbOpts = append(sbOpts, sandbox.NoPull())
	}
	sb := sandbox.New(manager.Image, sbOpts...)

	var result *analysis.AnalysisResult
	if live {
		result = analysis.RunLive(*ecosystem, pkgName, *version, sb, command)
	} else {
		result = analysis.RunLocal(*ecosystem, pkgName, *version, sb, command)
	}

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
