# Pulumi

### Overview

<https://www.pulumi.com/>

This diagram illustrates the structure and major components of Pulumi.

![Pulumi](https://www.pulumi.com/images/docs/pulumi-programming-model-diagram.svg)

Pulumi programs, written in general-purpose programming languages, describe how your cloud infrastructure should be composed. To declare new infrastructure in your program, you allocate resource objects whose properties correspond to the desired state of your infrastructure. These properties are also used between resources to handle any necessary dependencies and can be exported outside of the stack, if needed.

Programs reside in a project, which is a directory that contains source code for the program and metadata on how to run the program. After writing your program, you run the Pulumi CLI command pulumi up from within your project directory. This command creates an isolated and configurable instance of your program, known as a stack. Stacks are similar to different deployment environments that you use when testing and rolling out application updates. For instance, you can have distinct development, staging, and production stacks that you create and test against.

> source https://www.pulumi.com/docs/intro/concepts/

### Getting started

1. Install pulumi cli https://www.pulumi.com/docs/get-started/install/
2. Signup a free account https://app.pulumi.com/
3. `pulumi login` https://www.pulumi.com/docs/reference/cli/pulumi_login/
4. Create a new pulumi stack `pulumi stack init`
   - Provide stack name like `dev` or `your name`
   - Switch between `stacks` using `pulumi stack select`
5. This command to set the gcp project `pulumi config set gcp:project naveen-ossf-malware-analysis`
6. This command to set the zone gcp:zone: `pulumi config set us-central1-c`
7. This command to set config data for

```
   pulumi config set --path 'data.BucketName' naveen-analysis
   pulumi config set --path 'data.ClusterName' analysis
   pulumi config set --path 'data.MachineType' n1-standard-1
```

These settings are loaded into this struct

```
type data struct {
    CreateCluster: bool // should create k8s cluster
    CreateRegistry: bool // should create docker registry
    CreateBucket: bool // should create bucket
	ClusterName   string  // name of the k8s cluster
	NumberOfNodes int    // number of nodes that are going to be created
	MachineType   string // k8s node machine type
	BucketName    string  // package feeds analysis bucket name
}
```

The `config` is separated the code.

8. `pulumi preview` - Pulumi preview command previews the change that is happen on the `env` based on the code changes. Here is an example of the output.

```
❯ pulumi preview
Previewing update (naveen)

View Live: https://app.pulumi.com/naveensrinivasan/analyzefeeds/naveen/previews/538387dc-fa43-4eab-9846-ad73f7601fde

     Type                       Name                 Plan
 +   pulumi:pulumi:Stack        analyzefeeds-naveen  create
 +   ├─ gcp:projects:Service    container            create
 +   ├─ gcp:projects:Service    registry             create
 +   ├─ gcp:storage:Bucket      naveen-analysis      create
 +   ├─ gcp:container:Cluster   analysis             create
 +   └─ gcp:container:Registry  containerregistry    create

Resources:
    + 6 to create
```

9. `pulumi up` - Pulumi up will update the env with the desired state. There is confirmation to `apply` the changes.

```
❯ pulumi up
Previewing update (dev)

View Live: https://app.pulumi.com/naveensrinivasan/analyzefeeds/naveen/previews/8a8e7d23-6951-45f4-9588-be3b88b84a7e

     Type                       Name                 Plan
 +   pulumi:pulumi:Stack        analyzefeeds-naveen  create
 +   ├─ gcp:projects:Service    container            create
 +   ├─ gcp:storage:Bucket      naveen-analysis      create
 +   ├─ gcp:projects:Service    registry             create
 +   ├─ gcp:container:Cluster   analysis             create
 +   └─ gcp:container:Registry  containerregistry    create

Resources:
    + 6 to create

Do you want to perform this update?  [Use arrows to move, enter to select, type to filter]
  yes
> no
  details
```

10. Output of the run

```
View Live: https://app.pulumi.com/naveensrinivasan/analyzefeeds/naveen/updates/8

     Type                       Name                 Status
 +   pulumi:pulumi:Stack        analyzefeeds-naveen  creating...
 +   ├─ gcp:projects:Service    container            created
 +   ├─ gcp:projects:Service    registry             crea
     Type                       Name                 Sta
 +   pulumi:pulumi:Stack        analyzefeeds-naveen  cre
 +   ├─ gcp:projects:Service    container
 +   ├─ gcp:projects:Service    registry
 +   ├─ gcp:storage:Bucket      naveen-analysis
 +   ├─ gcp:container:Cluster   analysis
 +   └─ gcp:container:Registry  containerregistry

Outputs:
    bucket    : "gs://naveen-analysis"
    kubeconfig: "[secret]"
    registry  : "urn:pulumi:naveen::analyzefeeds::gcp:container/registry:Registry::containerregistry"
```

Outputs:
bucket : "gs://naveen-analysis"
kubeconfig: "[secret]"
registry : "urn:pulumi:naveen::analyzefeeds::gcp:container/registry:Registry::containerregistry"

11. Stack outputs - https://www.pulumi.com/docs/intro/concepts/stack/#import-and-export-a-stack-deployment

```
Outputs:
    bucket    : "gs://naveen-analysis"
    kubeconfig: "[secret]"
    registry  : "urn:pulumi:naveen::analyzefeeds::gcp:container/registry:Registry::containerregistry"
```

12. `kubeconfig` is a secret and not displayed as default output.

    `kubeconfig` can be exported `pulumi stack output kubeconfig --show-secrets > ~/.kube/config`

13. `pulumi destroy` with destroy all the resources created by `pulumi` . `destroy` asks for confirmation before destrorying the stack.

#### GitHub Actions

##### pulumi-pr

The pulumi-pr action would require these secrets and env variable to be set. This action would execute `pulumi preview` and comment the results as a `github` comment.

##### Secrets

GOOGLE_CREDENTIALS: ${{ secrets.GOOGLE_CREDENTIALS }} - <https://www.pulumi.com/docs/intro/cloud-providers/gcp/service-account/>
PULUMI_ACCESS_TOKEN: ${{ secrets.PULUMI_ACCESS_TOKEN }} - <https://www.pulumi.com/docs/intro/console/accounts-and-organizations/accounts/#access-tokens>

#### Env

PULUMI_ROOT: ./infrastructure
IS_PR_WORKFLOW: true
PULUMI_STACK_NAME: dev

#### pulumi

The difference between `pulumi-pr` and this action is that , this action would apply changes to the actual infrastructure. This action would run only `push` to `master`.

This action also requires these secrets and env being set

GOOGLE_CREDENTIALS: ${{ secrets.GOOGLE_CREDENTIALS }} 

PULUMI_ACCESS_TOKEN: ${{ secrets.PULUMI_ACCESS_TOKEN }} 

PULUMI_ROOT: /infrastructure 

PULUMI_STACK_NAME: dev 
