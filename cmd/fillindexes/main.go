package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"

	"github.com/ossf/package-analysis/internal/analysis"
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
	flag.Parse()
	if *bucket == "" || *docstorePath == "" {
		flag.Usage()
		return
	}

	ctx := context.Background()
	bkt, err := blob.OpenBucket(ctx, *bucket)
	if err != nil {
		log.Fatal(err)
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
					log.Fatal(err)
				}

				var results analysis.AnalysisResult
				err = json.Unmarshal(data, &results)
				if err != nil {
					log.Fatal(err)
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
			log.Fatal(err)
		}

		if !strings.HasSuffix(obj.Key, ".json") {
			continue
		}

		queue <- obj.Key
	}
	close(queue)
	wg.Wait()
}
