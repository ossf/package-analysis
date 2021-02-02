package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

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
	switch pkg.Type {
	case "pypi":
		return createPypiPod(pkg.Name, pkg.Version)
	case "npm":
		// createNpmPod(pkg.Name, pkg.Version)
	}
	return nil
}

func createPypiPod(name, version string) error {
	// We need to pass a bool pointer below.
	var token = false
	var retries int32 = 3
	var ttl int32 = 3600
	var deadline int64 = 600
	jobs := clientset.BatchV1().Jobs("default")

	job, err := jobs.Create(context.Background(), &bv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "pypi-",
			Labels: map[string]string{
				"install":         "1",
				"package_name":    name,
				"package_version": version,
				"package_type":    "pypi",
			},
		},
		Spec: bv1.JobSpec{
			ActiveDeadlineSeconds:   &deadline,
			TTLSecondsAfterFinished: &ttl,
			BackoffLimit:            &retries,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"install":         "1",
						"package_name":    name,
						"package_version": version,
						"package_type":    "pypi",
					},
				},
				Spec: v1.PodSpec{
					RestartPolicy:                v1.RestartPolicyNever,
					AutomountServiceAccountToken: &token,
					Containers: []v1.Container{
						{
							Name:    "install",
							Image:   "python:3",
							Command: []string{"/bin/bash", "-c"},
							Args: []string{
								fmt.Sprintf("mkdir -p /app && cd /app && pip3 install %s==%s", name, version),
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
