package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jordan-wright/ossmalware/pkg/library"

	"gocloud.dev/pubsub"
	_ "gocloud.dev/pubsub/gcppubsub"
)

const (
	url   = "https://crates.io/api/v1/summary"
	delta = 5 * time.Minute
)

var topicURL string

type crates struct {
	JustUpdated []struct {
		ID               string    `json:"id"`
		Name             string    `json:"name"`
		UpdatedAt        time.Time `json:"updated_at"`
		NewestVersion    string    `json:"newest_version"`
		MaxStableVersion string    `json:"max_stable_version"`
		Repository       string    `json:"repository"`
	} `json:"just_updated"`
}

// Package stores the information from crates.io updates
type Package struct {
	Title        string
	ModifiedDate time.Time
	Link         string
	Version      string
}

// Gets crates.io packages
func fetchCratePackages() ([]Package, error) {
	result := []Package{}
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	v := &crates{}
	err = json.NewDecoder(resp.Body).Decode(v)
	if err != nil {
		return nil, err
	}

	cutoff := time.Now().UTC().Add(-delta)
	for _, item := range v.JustUpdated {
		if item.UpdatedAt.Before(cutoff) {
			continue
		}
		result = append(result, Package{
			Title:        item.ID,
			ModifiedDate: item.UpdatedAt,
			Link:         item.Repository,
			Version:      item.NewestVersion,
		})
	}

	return result, nil
}

// Poll receives a message from Cloud Pub/Sub. Ideally, this will be from a
// Cloud Scheduler trigger every `delta`.
func Poll(w http.ResponseWriter, r *http.Request) {
	topic, err := pubsub.OpenTopic(context.TODO(), topicURL)
	if err != nil {
		panic(err)
	}

	packages, err := fetchCratePackages()
	if err != nil {
		log.Printf("error fetching packages: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	processed := 0
	for _, pkg := range packages {
		processed++
		msg := library.Package{
			Name:    pkg.Title,
			Version: pkg.Version,
			Type:    "crates",
		}
		b, err := json.Marshal(msg)
		if err != nil {
			log.Printf("error marshaling message: %#v", msg)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := topic.Send(context.TODO(), &pubsub.Message{Body: b}); err != nil {
			log.Printf("error sending package to upstream topic %s: %v", topicURL, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	//nolint:errcheck
	log.Printf("Processed %d packages", processed)
	//nolint:errcheck
	w.Write([]byte("OK"))
}

func main() {
	log.Print("polling crates for packages")
	var status bool
	if topicURL, status = os.LookupEnv("OSSMALWARE_TOPIC_URL"); !status {
		log.Fatal("OSSMALWARE_TOPIC_URL env variable is empty.")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("defaulting to port %s", port)
	}
	http.HandleFunc("/", Poll)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		log.Fatal(err)
	}
}
