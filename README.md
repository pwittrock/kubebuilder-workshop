# Kubebuilder Workshop

The Kubebuilder Workshop is aimed at providing hands-on experience building Kubernetes APIs using kubebuilder.
By the end of the workshop, participants will have built a Kubernetes native API for running MongoDB instances.

Once the API is installed into a Kubernetes cluster, users should be able to create new MongoDB instance similar
to the one in [this blog post](https://kubernetes.io/blog/2017/01/running-mongodb-on-kubernetes-with-statefulsets/) by
specifying a MongoDB config file and running `kubectl apply -f` on the file.

Example file:

```yaml
apiVersion: databases.k8s.io/v1alpha1
kind: MongoDB
metadata:
  name: mongo-instance
spec:
  replicas: 3
  storage: 100Gi
```

The Template to create a StatefuleSet and Service for running a MongoDB instance will be provided as Go functions.

## Goals

- Define a new Kubernetes API for running MongoDB instances
  - Define fields for users to specify
- Implement the API as a Controller
  - Creates StatefulSet and Service for the MongoDB instance
- Install the API into a Kubernetes Cluster using kubectl apply
- Test the API by creating a MongoDB instance

## Prerequisites

To get the most out of this Workshop, users should have:

- A basic understanding of [Go Programming Language](https://golang.org/)
- A basic understanding of [Kubernetes APIs](https://kubernetes.io/docs/user-journeys/users/application-developer/foundational/#section-2)

## Setup a Kubernetes API dev environment

### Install dev tools

Install the dev tools that will be used to build and publish the MongoDB API.

- Install Go 1.10+
  - [go](https://golang.org/)

- Install dep
  - [dep](https://github.com/golang/dep)

- Install kubectl
  - [link](https://kubernetes.io/docs/tasks/tools/install-kubectl/#install-kubectl)

- Install kustomize
  - [kustomize](https://github.com/kubernetes-sigs/kustomize)

- Install kubebuilder
  - [kubebuilder](https://book.kubebuilder.io/getting_started/installation_and_setup.html)  

### Create a dev cluster (Pick 1)

Setup a dev cluster.

- Google Kubernetes Engine (GKE) - Remote
  - Install [GKE](https://cloud.google.com/kubernetes-engine/)
  - Create the cluster `gcloud container clusters create "test-cluster"`
  - Fetch the credentials for the cluster `gcloud container clusters get-credentials test-cluster`
  - Test the setup `kubectl get nodes`

- Minikube - Local VM
  - Install [minikube](https://github.com/kubernetes/minikube)
  - Start a minikube cluster `minikube start`
  - Test the setup `kubectl get nodes`

## Supplementary Resources on Kubebuilder and Kubernetes APIs

Documentation on kubebuilder available here:

- [http://book.kubebuilder.io](http://book.kubebuilder.io)

## Building a MongoDB API

### Create a new Go project for your API

Create a new project for the workshop

- `mkdir -p $HOME/kubebuilder-workshop/src/github.com/my-org/my-project`
- `export GOPATH=$HOME/kubebuilder-workshop`
- `cd $HOME/kubebuilder-workshop/src/github.com/my-org/my-project`

### Initialize the project structure go library deps

- `kubebuilder init --domain k8s.io --license apache2 --owner "My Org"`
  - enter `y` to have it run dep for you
  - read on while you wait for `dep` to download the go library dependencies (takes ~3-5 minutes)

### Define an empty MongoDB API

Have kubebuilder create the boilerplate for a new Resource type and Controller

- `kubebuilder create api --group databases --version v1alpha1 --kind MongoDB`
  - enter `y` to have it create boilerplate for the Resource
  - enter `y` to have it create boilerplate for the Controller
  
This will also build the project and run the tests to make sure the resource and controller are hooked up correctly.
  
### Define the Schema for the MongoDB Resource

Define the Schema *Spec* for the MongoDB API

- edit `pkg/apis/databases/v1alpha/mongodb_types.go`

Add optional fields for users to specify when creating MongoDB instances - for example:

- `replicas` (int32)
- `storage` (string)

To make them optional do the following:

- set `// +optional`
- make them pointers with `*`
- add the `omitempty` struct tag

```go
type MongoDBSpec struct {
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// +optional
	Storage *string `json:"storage,omitempty"`
}
```

Example Spec for Kubernetes Pods:

- [PodSpec](https://github.com/kubernetes/api/blob/master/core/v1/types.go#L2715)

## Implement the MongoDB Controller

- Edit `pkg/controller/mongodb/mongodb_controller.go`
- Update the Watch statements in `add`
- Copy the Helper functions to generate StatefulSets and Services for MongoDB
- Update the Resources that get created in `Reconcile`
- Update the RBAC directives for StatefulSets and Services

### Update the Watch config

Update the `add` function to Watch the Resources you will generate from the Controller (Service + StatefulSet)

- (No-Op) Watch MongoDB (EnqueueRequestForObject) - this was done for you
- Add Watch Services - and map to the Owning MongoDB instance (EnqueueRequestForOwner) - you need to add this
- Add Watch StatefulSets - and map to the Owning MongoDB instance (EnqueueRequestForOwner) - you need to add this
- Delete Watch Deployments

### Generate Service and StatefulSet objects from the MongoDB instance

Copy the helper functions from [this sample code](https://github.com/pwittrock/kubebuilder-workshop/blob/master/pkg/controller/mongodb/helpers.go)
to generate the objects for you.  These will create go objects that you can use to Create or Update the Kubernetes
Resources.  Revisit these later to add more fields and customization.

### Controller Code

- Update the RBAC rules to give perms for StatefulSets and Services
  - `// +kubebuilder:rbac:groups=apps,resources=statefulesets,verbs=get;list;watch;create;update;patch;delete`
  - `// +kubebuilder:rbac:groups=,resources=services,verbs=get;list;watch;create;update;patch;delete`
  - `// +kubebuilder:rbac:groups=databases.k8s.io,resources=mongodbs,verbs=get;list;watch;create;update;patch;delete`
- Generate a Service using the copied function
- Change the boilerplate code to Create a Service if it doesn't exist, or Update it if one does
  - **Warning**: For Services, be careful to only update the *Selector* and *Ports* so as not to overwrite ClusterIP.
- Generate a StatefuleSet using the copied function
- Change the boilerplate code to Create a StatefulSet if it doesn't exist, or Update it if one does
  - **Note:** For StatefulSet you *can* update the full Spec if you want

## Test Your API

### Install the Resource into the Cluster

### Run the Controller locally

- `make install` # install the CRDs
- `make run` # run the controller

### Edit a sample MongoDB instance and create it

```yaml
apiVersion: databases.k8s.io/v1alpha1
kind: MongoDB
metadata:
  name: mongo-instance
spec:
  replicas: 1
  storage: 100Gi
```

- edit `config/samples/databases_v1alpha1_mongodb.yaml`
- create the mongodb instance
  - `kubectl apply -f config/samples/databases_v1alpha1_mongodb.yaml`
  - observe output from Controller

### Look at the Resources

- look at created resources
  - `kubectl get monogodbs`
  - `kubectl get statefulsets`
  - `kubectl get services`
  - `kubectl get pods`
    - **note**: the containers may be creating - wait for them to come up
  - `kubectl describe pods`
  - `kubectl logs mongo-instance-mongodb-statefulset-0 mongo`
- delete the mongodb instance
  - `kubectl delete -f config/samples/databases_v1alpha1_mongodb.yaml`
- look for garbage collected resources
  - `kubectl get monogodbs`
  - `kubectl get statefulsets`
  - `kubectl get services`
  - `kubectl get pods`

### Connect to the MongoDB instance from a Pod

- `kubectl run mongo-test -t -i --rm --image mongo bash`
- `mongo <address of service>:27017`


## Experiment

- Try deleting the statefulset - what happens when you look for it?
- Try deleting the service - what happens when you look for it?
- Try adding fields to control new things such as the Port
- Try adding a *Spec*, what useful things can you put in there?

## Bonus Objectives

### Build your API into a container and publish it

- `IMG=foo make docker-build` && `IMG=foo make docker-push`
- `kustomize build config/default > mongodb_api.yaml`
- `kubectl apply -f mongodb_api.yaml`
- Get logs from the Controller using `kubectl logs`

### Add Simple Schema Validation

- [Validation tags docs](https://book.kubebuilder.io/basics/simple_resource.html)

### Publish Events from the Controller code

- [Event docs](https://book.kubebuilder.io/beyond_basics/creating_events.html)

### Define and add a list of Conditions to the Status, then set them from the Controller

- Example: [Node Conditions](https://kubernetes.io/docs/concepts/architecture/nodes/#condition)

### Setup Defaulting and complex Validation

- [Webhook docs](https://book.kubebuilder.io/beyond_basics/what_is_a_webhook.html)
