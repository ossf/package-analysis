package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jordan-wright/ossmalware/pkg/library"
	"gocloud.dev/pubsub"
	_ "gocloud.dev/pubsub/gcppubsub"
	bv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var clientset *kubernetes.Clientset

func main() {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	subscriptionUrl := os.Getenv("OSSMALWARE_SUBSCRIPTION_URL")
	ctx := context.Background()
	sub, err := pubsub.OpenSubscription(ctx, subscriptionUrl)
	if err != nil {
		panic(err)
	}

	// Start the background cleanup job
	go cleanupJob()

	for {
		msg, err := sub.Receive(ctx)
		if err != nil {
			log.Println("error receiving message: ", err)
			continue
		}
		go func(m *pubsub.Message) {
			log.Println("handling message: ", string(m.Body))
			pkg := library.Package{}
			if err := json.Unmarshal(m.Body, &pkg); err != nil {
				log.Println("error unmarshalling json: ", err)
				return
			}
			if err := handlePkg(pkg); err != nil {
				fmt.Println("Error: ", err)
				msg.Nack()
				return
			}
			msg.Ack()
		}(msg)
	}
}

func handlePkg(pkg library.Package) error {
	// Turn it into a Pod!
	if pkg.Type == "pypi" || pkg.Type == "npm" {
		return createPod(pkg.Name, pkg.Version, pkg.Type)
	}
	log.Println("unknown package type: ", pkg.Type)
	return nil
}

func createPod(name, version, packageType string) error {
	// We need to pass a bool pointer below.
	var token = false
	var retries int32 = 3
	var ttl int32 = 3600
	var deadline int64 = 600
	jobs := clientset.BatchV1().Jobs("default")

	var image, command string
	switch packageType {
	case "npm":
		image = "node"
		command = "npm init --force && npm install %s@%s"
	case "pypi":
		image = "python:3"
		command = "pip3 install --no-deps %s==%s"
	}

	job, err := jobs.Create(context.Background(), &bv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: packageType + "-",
			Labels: map[string]string{
				"install":      "1",
				"package_type": packageType,
			},
			Annotations: map[string]string{
				"package_name":    name,
				"package_version": version,
			},
		},
		Spec: bv1.JobSpec{
			ActiveDeadlineSeconds:   &deadline,
			TTLSecondsAfterFinished: &ttl,
			BackoffLimit:            &retries,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"install":      "1",
						"package_type": packageType,
					},
					Annotations: map[string]string{
						"package_name":    name,
						"package_version": version,
					},
				},
				Spec: v1.PodSpec{
					RestartPolicy:                v1.RestartPolicyNever,
					AutomountServiceAccountToken: &token,
					Containers: []v1.Container{
						{
							Name:    "install",
							Image:   image,
							Command: []string{"/bin/bash", "-c"},
							Args: []string{
								"mkdir /app && cd /app && " + fmt.Sprintf(command, name, version),
							},
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceCPU: resource.MustParse("250m"),
								},
								Limits: v1.ResourceList{
									v1.ResourceCPU: resource.MustParse("500m"),
								},
							},
						},
					},
				},
			},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	fmt.Println("Created job: ", job.Name)
	return nil
}

// k8s has a TTL controller, but it's alpha.
func cleanupJob() {
	jc := clientset.BatchV1().Jobs("default")

	ctx := context.Background()
	for {
		time.Sleep(time.Minute)
		// Delete everything completed
		jobs, err := jc.List(ctx, metav1.ListOptions{})
		if err != nil {
			log.Printf("error listing jobs: %s", err)
			continue
		}

		succeededOldest := time.Now().Add(-1 * time.Hour)
		failedOldest := time.Now().Add(-24 * time.Hour)
		for _, j := range jobs.Items {
			var oldest time.Time
			if j.Status.Succeeded == 1 {
				oldest = succeededOldest
			} else {
				oldest = failedOldest
			}
			if j.Status.StartTime.Time.Before(oldest) {
				log.Printf("Deleting job %s with start time %s", j.ObjectMeta.Name, j.Status.StartTime)
				if err := jc.Delete(ctx, j.ObjectMeta.Name, metav1.DeleteOptions{}); err != nil {
					log.Printf("error deleting job: %s %s", j.ObjectMeta.Name, err)
				}
			}
		}
	}
}
