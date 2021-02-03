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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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
var clientset *kubernetes.Clientset

func main() {
	ctx := context.Background()

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	bucket, err = blob.OpenBucket(ctx, os.Getenv("OSSF_MALWARE_ANALYSIS_RESULTS"))
	if err != nil {
		log.Panic(err)
	}
	defer bucket.Close()

	http.HandleFunc("/falco", falcoHandler)
	http.ListenAndServe(":8080", nil)
}

func setTimer(pod string) {
	if podTimers[pod] != nil {
		podTimers[pod].Stop()
	}
	podTimers[pod] = time.AfterFunc(time.Minute, finalizePod(pod))
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
		log.Println("skipping upload for pod:", pod)
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
	setTimer(pod)
	fmt.Println(output.Rule, output.OutputFields)
}

type data struct {
	Files []string
}

func finalizePod(podName string) func() {
	return func() {
		log.Println("Fetching info for pod: ", podName)
		ctx := context.Background()
		// Fetch the pod once.
		pod, err := clientset.CoreV1().Pods("default").Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			log.Println("fetching pod: ", err)
			return
		}

		mutex.Lock()
		defer mutex.Unlock()
		// ips := podIps[pod]
		files := podFiles[podName]

		d := data{}

		for f, _ := range files {
			d.Files = append(d.Files, f)
		}

		b, err := json.Marshal(d)
		if err != nil {
			log.Print(err)
			return
		}

		path := filepath.Join(
			pod.ObjectMeta.Labels["package_type"],
			pod.ObjectMeta.Annotations["package_name"],
			pod.ObjectMeta.Annotations["package_version"],
			"results.json")

		log.Printf("Uploading files and ips for %s to %s\n", podName, path)
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
		delete(podFiles, podName)
		delete(podTimers, podName)
	}
}
