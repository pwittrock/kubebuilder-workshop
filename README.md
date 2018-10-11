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

## Steps

1. Define a MongoDB *Resource* (API)
1. Implement the MongoDB Controller (API Implementation)
1. Install the Resource and start the Controller
1. Create a new MongoDB instance

## Prerequisites

See [kubebuilder-workshop-prereqs](https://github.com/pwittrock/kubebuilder-workshop-prereqs)

## Building a MongoDB API

- Init a Go project
- Create the Resource
- Create the Controller
  - Watch
  - Reconcile

### Create a new Go project for your API

Create a new go project

- `mkdir -p $HOME/kubebuilder-workshop/src/github.com/my-org/my-project`
- `export GOPATH=$HOME/kubebuilder-workshop`
- `cd $HOME/kubebuilder-workshop/src/github.com/my-org/my-project`

Initialize the project

- `kubebuilder init --domain k8s.io --license apache2 --owner "My Org"`
  - enter `y` to have it run dep for you
  - read on while you wait for `dep` to download the go library dependencies (takes ~3-5 minutes)

### Scaffold the boilerplate for the MongoDB Resource and Controller

Have kubebuilder create the boilerplate for a new Resource type and Controller

- `kubebuilder create api --group databases --version v1alpha1 --kind MongoDB`
  - enter `y` to have it create boilerplate for the Resource
  - enter `y` to have it create boilerplate for the Controller
  
This will also build the project and run the tests to make sure the resource and controller are hooked up correctly.
  
### Define your Schema for the MongoDB Resource

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

The MongoDB Controller will create a StatefulSet and Service that run MongoDB.

- Add Watch Statements for StatefulSets and Services
- Generate the desired StatefuleSet and Service
- Generate the StatefulSet and Service and compare to what is live
- Create or Update the StatefulSet and Service

### Update Watch

- Edit `pkg/controller/mongodb/mongodb_controller.go`

Update the `add` function to Watch the Resources you will generate from the Controller (Service + StatefulSet)

- (No-Op) Watch MongoDB (EnqueueRequestForObject) - this was done for you
- Add Watch Services - and map to the Owning MongoDB instance (EnqueueRequestForOwner) - you need to add this
- Add Watch StatefulSets - and map to the Owning MongoDB instance (EnqueueRequestForOwner) - you need to add this
- Delete Watch Deployments

### Generate Service and StatefulSet objects from the MongoDB instance

**Copy the helper functions from [this sample code](https://github.com/pwittrock/kubebuilder-workshop/blob/master/pkg/controller/mongodb/helpers.go)**
to generate the objects for you.  These will create go objects that you can use to Create or Update the Kubernetes
Resources.  Revisit these later to add more fields and customization.

### Update Reconcile

- Edit `pkg/controller/mongodb/mongodb_controller.go`

Update the `Reconcile` function to create/update the StatefulSet and Service objects

- Generate a Service using the copied function
- Change the boilerplate code to Create a Service if it doesn't exist, or Update it if one does
  - **Warning**: For Services, be careful to only update the *Selector* and *Ports* so as not to overwrite ClusterIP.
- Generate a StatefuleSet using the copied function
- Change the boilerplate code to Create a StatefulSet if it doesn't exist, or Update it if one does
  - **Note:** For StatefulSet you *can* update the full Spec if you want

- Optional: update the RBAC rules to give perms for StatefulSets and Services (needed for if running as a container in a cluster)
  - `// +kubebuilder:rbac:groups=apps,resources=statefulesets,verbs=get;list;watch;create;update;patch;delete`
  - `// +kubebuilder:rbac:groups=,resources=services,verbs=get;list;watch;create;update;patch;delete`
  - `// +kubebuilder:rbac:groups=databases.k8s.io,resources=mongodbs,verbs=get;list;watch;create;update;patch;delete`

## Test Your API

### Install the Resource into the Cluster

- `make install` # install the CRDs

### Run the Controller locally

- `make run` # run the controller as a local process

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

### Look at the Resources in the cluster

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
- look for garbage collected resources (they should be gone)
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
