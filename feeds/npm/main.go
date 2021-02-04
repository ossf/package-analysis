package main

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jordan-wright/ossmalware/pkg/library"
	"github.com/pkg/errors"
	"gocloud.dev/pubsub"
	_ "gocloud.dev/pubsub/gcppubsub"
)

const (
	delta      = 5 * time.Minute
	timeout    = 10 * time.Second
	baseURL    = "https://registry.npmjs.org/-/rss"
	versionURL = "https://registry.npmjs.org/"
)

var topicURL string

type Response struct {
	Packages []*Package `xml:"channel>item"`
}

type Package struct {
	Title        string      `xml:"title"`
	ModifiedDate rfc1123Time `xml:"pubDate"`
	Link         string      `xml:"link"`
	Version      string
}

type PackageVersion struct {
	ID       string `json:"_id"`
	Rev      string `json:"_rev"`
	Name     string `json:"name"`
	DistTags struct {
		Latest string `json:"latest"`
	} `json:"dist-tags"`
}

type rfc1123Time struct {
	time.Time
}

func (t *rfc1123Time) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var marshaledTime string
	err := d.DecodeElement(&marshaledTime, &start)
	if err != nil {
		return errors.Wrap(err, "unable to unmarshal xml")
	}
	decodedTime, err := time.Parse(time.RFC1123, marshaledTime)
	if err != nil {
		return errors.Wrap(err, "unable to deocde time")
	}
	*t = rfc1123Time{decodedTime}
	return nil
}

func fetchPackages() ([]*Package, error) {
	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(baseURL)
	if err != nil {
		return nil, errors.Wrap(err, "unable to connect")
	}
	defer resp.Body.Close()
	rssResponse := &Response{}
	err = xml.NewDecoder(resp.Body).Decode(rssResponse)
	if err != nil {
		return nil, errors.Wrap(err, "unable to decode the xml response")
	}
	return rssResponse.Packages, nil
}

// Gets the package version from the NPM.
func fetchVersionInformation(packageName string) (string, error) {
	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(fmt.Sprintf("%s/%s", versionURL, packageName))
	if err != nil {
		return "", errors.Wrap(err, "unable to get http for version information")
	}
	defer resp.Body.Close()
	v := &PackageVersion{}
	err = json.NewDecoder(resp.Body).Decode(v)
	if err != nil {
		return "", errors.Wrap(err, "unable to decode json for the version information")
	}
	return v.DistTags.Latest, nil
}

// Poll receives a message from Cloud Pub/Sub. Ideally, this will be from a
// Cloud Scheduler trigger every `delta`.
func Poll(w http.ResponseWriter, r *http.Request) {
	topic, err := pubsub.OpenTopic(context.TODO(), topicURL)
	if err != nil {
		panic(err)
	}

	packages, err := fetchPackages()
	if err != nil {
		log.Printf("error fetching packages: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	cutoff := time.Now().UTC().Add(-delta)
	processed := 0
	for _, pkg := range packages {
		log.Println("Processing:", pkg.Title, pkg.Version)
		if pkg.ModifiedDate.Before(cutoff) {
			continue
		}
		v, err := fetchVersionInformation(pkg.Title)
		if err != nil {
			log.Printf("error in fetching version information: %#v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		processed++
		msg := library.Package{
			Name:    pkg.Title,
			Version: v,
			Type:    "npm",
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

	log.Printf("Processed %d packages", processed)
	//nolint:errcheck
	w.Write([]byte("OK"))
}

func main() {
	log.Print("polling npm for packages")
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
