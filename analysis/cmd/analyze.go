package main

import (
	"context"
	"flag"
	"log"
	"regexp"
	"strings"

	"github.com/ossf/package-analysis/analysis"
)

var (
	pkg      = flag.String("package", "", "ecosystem/package")
	version  = flag.String("version", "", "version")
	upload   = flag.String("upload", "", "bucket path for uploading results")
	docstore = flag.String("docstore", "", "docstore path")
)

func parseBucketPath(path string) (string, string) {
	pattern := regexp.MustCompile(`(.*?://[^/]+)/(.*)`)
	match := pattern.FindStringSubmatch(path)
	if match == nil {
		log.Panic("Failed to parse bucket path: %s", path)
	}

	return match[1], match[2]
}

func main() {
	flag.Parse()
	if *pkg == "" {
		flag.Usage()
		return
	}
	pkgParts := strings.SplitN(*pkg, "/", 2)
	if len(pkgParts) != 2 {
		log.Panicf("Invalid package format: %s", *pkg)
	}

	ecosystem, name := pkgParts[0], pkgParts[1]

	manager, ok := analysis.SupportedPkgManagers[ecosystem]
	if !ok {
		log.Panicf("Unsupported pkg manager %s", manager)
	}

	if *version == "" {
		*version = manager.GetLatest(name)
	}

	command := manager.CommandFmt(name, *version)
	result := analysis.Run(ecosystem, name, *version, manager.Image, command)

	ctx := context.Background()
	if *upload != "" {
		bucket, path := parseBucketPath(*upload)
		err := analysis.UploadResults(ctx, bucket, path, result)
		if err != nil {
			log.Fatalf("Failed to upload results to blobstore: %v", err)
		}
	}

	if *docstore != "" {
		err := analysis.WriteResultsToDocstore(ctx, *docstore, result)
		if err != nil {
			log.Fatalf("Failed to write results to docstore: %v", err)
		}
	}
}
