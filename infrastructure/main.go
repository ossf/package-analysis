package main

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v4/go/gcp/container"
	"github.com/pulumi/pulumi-gcp/sdk/v4/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v4/go/gcp/storage"
	"github.com/pulumi/pulumi/sdk/v2/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v2/go/pulumi/config"
)

// config that holds data for each stack. Stacks are dev,staging, prod.
// https://www.pulumi.com/docs/intro/concepts/stack/
type data struct {
	CreateCluster  bool
	CreateRegistry bool
	CreateBucket   bool
	ClusterName    string
	NumberOfNodes  int
	MachineType    string
	BucketName     string
}

// infrastructure for analysis.
type infrastructure struct {
	Bucket     *storage.Bucket
	Registry   *container.Registry
	Kubeconfig pulumi.StringOutput
}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		var configuration data

		conf := config.New(ctx, "")
		conf.RequireObject("data", &configuration)
		infra, err := createAnalysisInfrastructure(ctx, configuration)
		if err != nil {
			return err
		}

		// https://www.pulumi.com/docs/intro/concepts/stack/#outputs
		if configuration.CreateCluster {
			ctx.Export("kubeconfig", infra.Kubeconfig)
		}
		if configuration.CreateBucket {
			ctx.Export("bucket", infra.Bucket.Url)
		}
		if configuration.CreateRegistry {
			ctx.Export("registry", infra.Registry.URN())
		}

		return nil
	})
}

// createAnalysisInfrastructure creates k8s cluster, container registry and gcs bucket.
func createAnalysisInfrastructure(ctx *pulumi.Context, configuration data) (*infrastructure, error) {
	const location = "US"
	infra := infrastructure{}

	if configuration.CreateCluster {
		if configuration.ClusterName == "" {
			return nil, errors.New("cluster name cannot be empty")
		}
		if configuration.NumberOfNodes < 1 {
			return nil, errors.New("numberofnodes cannot less than 1")
		}
		if configuration.MachineType == "" {
			return nil, errors.New("machine type cannot be empty")
		}
	}

	if configuration.CreateCluster {
		// enabling container api
		containerService, err := projects.NewService(ctx, "container", &projects.ServiceArgs{
			Service: pulumi.String("container.googleapis.com"),
		})
		if err != nil {
			return nil, errors.Wrap(err, "error in creating container Service")
		}

		engineVersions, err := container.GetEngineVersions(ctx, &container.GetEngineVersionsArgs{})
		if err != nil {
			return nil, errors.Wrap(err, "error in k8s engine Versions")
		}
		masterVersion := engineVersions.LatestMasterVersion

		// creating k8s cluster
		cluster, err := container.NewCluster(ctx, configuration.ClusterName, &container.ClusterArgs{
			Name:             pulumi.String(configuration.ClusterName),
			InitialNodeCount: pulumi.Int(configuration.NumberOfNodes),
			MinMasterVersion: pulumi.String(masterVersion),
			NodeVersion:      pulumi.String(masterVersion),
			// workload identity https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity
			WorkloadIdentityConfig: container.ClusterWorkloadIdentityConfigArgs{
				IdentityNamespace: pulumi.String(fmt.Sprintf("%s.svc.id.goog", config.Get(ctx, "gcp:project"))),
			},
			NodeConfig: &container.ClusterNodeConfigArgs{
				MachineType: pulumi.String(configuration.MachineType),
				OauthScopes: pulumi.StringArray{
					pulumi.String("https://www.googleapis.com/auth/compute"),
					pulumi.String("https://www.googleapis.com/auth/devstorage.read_only"),
					pulumi.String("https://www.googleapis.com/auth/logging.write"),
					pulumi.String("https://www.googleapis.com/auth/monitoring"),
				},
			},
		}, pulumi.DependsOn([]pulumi.Resource{containerService})) // dependency to containerService being turned on.
		if err != nil {
			return nil, errors.Wrap(err, "error in creating k8s")
		}
		infra.Kubeconfig = generateKubeconfig(cluster.Endpoint, cluster.Name, cluster.MasterAuth)
	}

	if configuration.CreateBucket {
		// creating a bucket for storing analysis results
		bucket, err := storage.NewBucket(ctx, configuration.BucketName, &storage.BucketArgs{
			Name: pulumi.String(configuration.BucketName),
		})
		if err != nil {
			return nil, errors.Wrap(err, "error in creating buket")
		}
		infra.Bucket = bucket
	}

	if configuration.CreateRegistry {
		// turning on registry service
		registryService, err := projects.NewService(ctx, "registry", &projects.ServiceArgs{
			Service: pulumi.String("containerregistry.googleapis.com"),
		})
		if err != nil {
			return nil, errors.Wrap(err, "error in creating registry service")
		}

		// creating a container registry
		registry, err := container.NewRegistry(ctx, "containerregistry",
			&container.RegistryArgs{Location: pulumi.String(location)},
			pulumi.DependsOn([]pulumi.Resource{registryService}))
		if err != nil {
			return nil, errors.Wrap(err, "error in creating registry")
		}

		infra.Registry = registry
	}

	return &infra, nil
}

// generate kubeconfig from created k8s cluster.
func generateKubeconfig(clusterEndpoint pulumi.StringOutput, clusterName pulumi.StringOutput,
	clusterMasterAuth container.ClusterMasterAuthOutput) pulumi.StringOutput {
	context := pulumi.Sprintf("demo_%s", clusterName)

	return pulumi.Sprintf(`apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: %s
    server: https://%s
  name: %s
contexts:
- context:
    cluster: %s
    user: %s
  name: %s
current-context: %s
kind: Config
preferences: {}
users:
- name: %s
  user:
    auth-provider:
      config:
        cmd-args: config config-helper --format=json
        cmd-path: gcloud
        expiry-key: '{.credential.token_expiry}'
        token-key: '{.credential.access_token}'
      name: gcp`,
		clusterMasterAuth.ClusterCaCertificate().Elem(),
		clusterEndpoint, context, context, context, context, context, context)
}
