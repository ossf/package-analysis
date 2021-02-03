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
	delta   = 5 * time.Minute
	baseURL = "https://rubygems.org/api/v1/activity"
)

type Package struct {
	Name         string    `json:"name"`
	Version      string    `json:"version"`
	ModifiedDate time.Time `json:"version_created_at"`
}

func fetchPackages(url string) ([]*Package, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	response := []*Package{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	return response, err
}

// Poll receives a message from Cloud Pub/Sub. Ideally, this will be from a
// Cloud Scheduler trigger every `delta`.
func Poll(w http.ResponseWriter, r *http.Request) {
	topicURL := os.Getenv("OSSMALWARE_TOPIC_URL")
	topic, err := pubsub.OpenTopic(context.TODO(), topicURL)
	if err != nil {
		panic(err)
	}

	packages := make(map[string]*Package)
	newPackages, err := fetchPackages(fmt.Sprintf("%s/%s", baseURL, "latest.json"))
	if err != nil {
		log.Printf("error fetching new packages: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, pkg := range newPackages {
		packages[pkg.Name] = pkg
	}
	updatedPackages, err := fetchPackages(fmt.Sprintf("%s/%s", baseURL, "just_updated.json"))
	if err != nil {
		log.Printf("error fetching updated packages: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, pkg := range updatedPackages {
		packages[pkg.Name] = pkg
	}

	cutoff := time.Now().UTC().Add(-delta)
	processed := 0
	for _, pkg := range packages {
		log.Println("Processing:", pkg.Name, pkg.Version)
		if pkg.ModifiedDate.Before(cutoff) {
			continue
		}
		processed++
		msg := library.Package{
			Name:    pkg.Name,
			Version: pkg.Version,
			Type:    "rubygems",
		}
		b, err := json.Marshal(msg)
		if err != nil {
			log.Printf("error marshaling message: %#v", msg)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := topic.Send(context.TODO(), &pubsub.Message{
			Body: b,
		}); err != nil {
			log.Printf("error sending package to upstream topic %s: %v", topicURL, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	log.Printf("Processed %d packages", processed)
	w.Write([]byte("OK"))
}

func main() {
	log.Print("polling pypi for packages")
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
