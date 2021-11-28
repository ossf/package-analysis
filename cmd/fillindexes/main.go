package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/ossf/package-analysis/internal/analysis"
	"github.com/ossf/package-analysis/internal/log"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/gcsblob"
)

var (
	bucket       = flag.String("bucket", "", "bucket")
	docstorePath = flag.String("docstore", "", "docstore path to write to")
)

const (
	numWorkers = 128
)

func main() {
	log.Initalize(false)
	flag.Parse()
	if *bucket == "" || *docstorePath == "" {
		flag.Usage()
		return
	}

	ctx := context.Background()
	bkt, err := blob.OpenBucket(ctx, *bucket)
	if err != nil {
		log.Fatal("Failed to open bucket",
			"bucket", *bucket,
			"error", err)
	}
	defer bkt.Close()

	queue := make(chan string, numWorkers)
	var wg sync.WaitGroup
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			for key := range queue {
				fmt.Println(key)
				data, err := bkt.ReadAll(ctx, key)
				if err != nil {
					log.Fatal("Failed to read bucket",
						"key", key,
						"error", err)
				}

				var results analysis.AnalysisResult
				err = json.Unmarshal(data, &results)
				if err != nil {
					log.Fatal("Failed to parse JSON data",
						"error", err)
				}

				analysis.WriteResultsToDocstore(ctx, *docstorePath, &results)
			}
			wg.Done()
		}()
	}

	it := bkt.List(nil)
	for {
		obj, err := it.Next(ctx)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("Failed to get next bucket entry",
				"error", err)
		}

		if !strings.HasSuffix(obj.Key, ".json") {
			continue
		}

		queue <- obj.Key
	}
	close(queue)
	wg.Wait()
}
