package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gocloud.dev/blob"
	_ "gocloud.dev/blob/gcsblob"
)

var mutex sync.Mutex
var podFiles = map[string]map[string]bool{}
var podTimers = map[string]*time.Timer{}

//{"output":"14:10:23.822542303: Notice unexpected file access (command=falco --cri /run/containerd/containerd.sock -K /var/run/secrets/kubernetes.io/serviceaccount/token -k https://10.68.0.1 -pk fd=/etc/hosts user=root k8s.ns=default k8s.pod=falco-4qg8j container=a1ce096597d1 image=sha256:8e85af245293402c1219ad382aeb71c9336a525c7afe595fdbd9dcd040c9103b)","priority":"Notice","rule":"Unexpected file access","time":"2021-01-26T14:10:23.822542303Z", "output_fields": {"container.id":"a1ce096597d1","container.image":"sha256:8e85af245293402c1219ad382aeb71c9336a525c7afe595fdbd9dcd040c9103b","evt.time":1611670223822542303,"fd.name":"/etc/hosts","k8s.ns.name":"default","k8s.pod.name":"falco-4qg8j","proc.cmdline":"falco --cri /run/containerd/containerd.sock -K /var/run/secrets/kubernetes.io/serviceaccount/token -k https://10.68.0.1 -pk","user.name":"root"}}
type falcoOutput struct {
	Output       string
	Priority     string
	Rule         string
	Time         string
	OutputFields OutputFields `json:"output_fields"`
}

type OutputFields struct {
	ContainerImage string `json:"container.image"`
	Pod            string `json:"k8s.pod.name"`
	CmdLine        string `json:"proc.cmdline"`
	IP             string `json:"fd.sip"`
	Path           string `json:"fd.name"`
	Labels         string `json:"k8s.pod.labels"`
}

var bucket *blob.Bucket

func main() {
	ctx := context.Background()
	var err error
	bucket, err = blob.OpenBucket(ctx, os.Getenv("OSSF_MALWARE_ANALYSIS_RESULTS"))
	if err != nil {
		log.Panic(err)
	}
	defer bucket.Close()

	http.HandleFunc("/falco", falcoHandler)
	http.ListenAndServe(":8080", nil)
}

func setTimer(pod string, labels map[string]string) {
	if podTimers[pod] != nil {
		podTimers[pod].Stop()
	}
	podTimers[pod] = time.AfterFunc(time.Minute, finalizePod(pod, labels))
}

func falcoHandler(w http.ResponseWriter, r *http.Request) {
	// Decode the container name
	dec := json.NewDecoder(r.Body)
	defer r.Body.Close()
	var output falcoOutput
	if err := dec.Decode(&output); err != nil {
		log.Print(err)
		return
	}
	mutex.Lock()
	defer mutex.Unlock()
	pod := output.OutputFields.Pod

	// Not from one of our pods.
	if pod == "" {
		return
	}
	kvps := strings.Split(output.OutputFields.Labels, ", ")
	labels := map[string]string{}
	for _, kvp := range kvps {
		kv := strings.SplitN(kvp, ":", 2)
		if len(kv) != 2 {
			continue
		}
		labels[kv[0]] = kv[1]
	}

	// Not a pod for analysis, ignore
	if labels["install"] != "1" {
		return
	}

	switch output.Rule {
	case "Unexpected file access":
		if podFiles[pod] == nil {
			podFiles[pod] = map[string]bool{}
		}
		podFiles[pod][output.OutputFields.Path] = true
	default:
		return
	}
	setTimer(pod, labels)
	fmt.Println(output.Rule, output.OutputFields)
}

type data struct {
	Files []string
}

func finalizePod(pod string, labels map[string]string) func() {
	return func() {
		ctx := context.Background()
		mutex.Lock()
		defer mutex.Unlock()
		// ips := podIps[pod]
		files := podFiles[pod]

		d := data{}

		for f, _ := range files {
			d.Files = append(d.Files, f)
		}

		b, err := json.Marshal(d)
		if err != nil {
			log.Print(err)
			return
		}
		log.Printf("Uploading files and ips for %s\n", pod)

		path := filepath.Join(
			labels["package_type"],
			labels["package_name"],
			labels["package_version"],
			"results.json")
		w, err := bucket.NewWriter(ctx, path, nil)
		if err != nil {
			log.Print(err)
			return
		}
		if _, err := w.Write(b); err != nil {
			log.Print(err)
			return
		}
		if err := w.Close(); err != nil {
			log.Print(err)
			return
		}
		delete(podFiles, pod)
		delete(podTimers, pod)
	}
}
