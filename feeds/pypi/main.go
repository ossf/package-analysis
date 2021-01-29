package main

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jordan-wright/ossmalware/pkg/library"
	"github.com/jordan-wright/ossmalware/pkg/processor"

	"gocloud.dev/pubsub"
	_ "gocloud.dev/pubsub/gcppubsub"
)

const (
	delta   = 5 * time.Minute
	baseURL = "https://pypi.org/rss/updates.xml"
)

type Response struct {
	Packages []*Package `xml:"channel>item"`
}

type Package struct {
	Title        string      `xml:"title"`
	ModifiedDate rfc1123Time `xml:"pubDate"`
	Link         string      `xml:"link"`
}

type rfc1123Time struct {
	time.Time
}

func (p *Package) Name() string {
	// The XML Feed has a "Title" element that contains the package and version in it.
	return strings.Split(p.Title, " ")[0]
}

func (p *Package) Version() string {
	// The XML Feed has a "Title" element that contains the package and version in it.
	return strings.Split(p.Title, " ")[1]
}

func (t *rfc1123Time) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var marshaledTime string
	err := d.DecodeElement(&marshaledTime, &start)
	if err != nil {
		return err
	}
	decodedTime, err := time.Parse(time.RFC1123, marshaledTime)
	if err != nil {
		return err
	}
	*t = rfc1123Time{decodedTime}
	return nil
}

func fetchPackages() ([]*Package, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get(baseURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	rssResponse := &Response{}
	err = xml.NewDecoder(resp.Body).Decode(rssResponse)
	if err != nil {
		return nil, err
	}
	return rssResponse.Packages, nil
}

// Poll receives a message from Cloud Pub/Sub. Ideally, this will be from a
// Cloud Scheduler trigger every `delta`.
func Poll(w http.ResponseWriter, r *http.Request) {
	topicURL := os.Getenv("OSSMALWARE_TOPIC_URL")
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
		log.Println("Processing:", pkg.Name(), pkg.Version())
		if pkg.ModifiedDate.Before(cutoff) {
			continue
		}
		processed++
		msg := library.Package{
			Name:    pkg.Name(),
			Version: pkg.Version(),
			Type:    processor.TypePyPI,
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
