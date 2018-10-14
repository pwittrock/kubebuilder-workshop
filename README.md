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

## **Prerequisites**

**Important:** Do these first

See [kubebuilder-workshop-prereqs](https://github.com/pwittrock/kubebuilder-workshop-prereqs)

## Steps

1. Define a MongoDB *Resource* (API Definition)
1. Implement the MongoDB Controller (API Implementation)
  - Update Watch for the Resources the Controller creates / updates
  - Update Reconcile to create / update the MongoDB StatefulSet and Service
1. Install the Resource into a cluster and start the Controller locally
1. Test the API by creating a MongoDB instance

### Scaffold the boilerplate for the MongoDB Resource and Controller

Scaffold the boilerplate for a new MongoDB Resource type and Controller

**Note:** This will also build the project and run the tests to make sure the resource and controller are hooked up
correctly.

- `kubebuilder create api --group databases --version v1alpha1 --kind MongoDB`
  - enter `y` to have it create boilerplate for the Resource
  - enter `y` to have it create boilerplate for the Controller
  
### Update the scaffolded the Schema with the MongoDB Resource API Definition

Define the MongoDB API Schema *Spec*for in `pkg/apis/databases/v1alpha/mongodb_types.go`.

**Note:** Copy the following Spec, optionally revisit later to add more fields.

```go
type MongoDBSpec struct {
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// +optional
	Storage *string `json:"storage,omitempty"`
}
```

Spec contains 2 optional fields:

- `replicas` (int32)
- `storage` (string)

**Note:** Fields have been made optional by:

- setting `// +optional`
- making them pointers with `*`
- adding the `omitempty` struct tag

Further reading: [PodSpec Example](https://github.com/kubernetes/api/blob/master/core/v1/types.go#L2715)

### Implement the MongoDB Controller

The MongoDB Controller should manage a StatefulSet and Service for running MongoDB.

Steps:

- Add Watch Statements for StatefulSets and Services (instructions below)
- Generate the desired StatefuleSet and Service (instructions below)
- Generate the StatefulSet and Service and compare to what is live (instructions below)
- Create or Update the StatefulSet and Service (instructions below)

### Update `add` with Watches

Update the `add` function to Watch the Resources you will be creating / updating in
`pkg/controller/mongodb/mongodb_controller.go`.

- (No-Op) Watch MongoDB (EnqueueRequestForObject) - this was scaffolded for you
- Add Watch Services - and map to the Owning MongoDB instance (EnqueueRequestForOwner) - you need to add this
- Add Watch StatefulSets - and map to the Owning MongoDB instance (EnqueueRequestForOwner) - you need to add this
- Delete Watch Deployments - you aren't managing Deployments

### Update `Reconcile` with object creation

Update the `Reconcile` function to Create / Update the StatefulSet and Service objects to run MongoDB in
`pkg/controller/mongodb/mongodb_controller.go`.

- Use the provided `GenerateService` function to get a Service struct instance
- Use the struct to either Create or Update a Service to run MongoDB
  - **Warning**: For Services, be careful to only update the *Selector* and *Ports* so as not to overwrite ClusterIP.
- Use the provided `GenerateStatefuleSet` function to get a StatefulSet struct instance
- Use the struct to either Create or Update a StatefulSet to run MongoDB
  - **Note:** For StatefulSet you *can* update the full Spec if you want

- **Optional:** for when running in cluster - update the RBAC rules to give perms for StatefulSets and
  Services (needed for if running as a container in a cluster)
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

- requires [kustomize](https://github.com/kubernetes-sigs/kustomize)
- `IMG=foo make docker-build` && `IMG=foo make docker-push`
- `kustomize build config/default > mongodb_api.yaml`
- `kubectl apply -f mongodb_api.yaml`
- Get logs from the Controller using `kubectl logs`

### Add Simple Schema Validation

- [Validation tags docs](https://book.kubebuilder.io/basics/simple_resource.html)

### Publish Events from the Controller code

- [Event docs](https://book.kubebuilder.io/beyond_basics/creating_events.html)

### Add Operational Logic to the Reconcile

Add logic to the Reconcile to handle lifecycle events.

### Define and add a list of Conditions to the Status, then set them from the Controller

- Example: [Node Conditions](https://kubernetes.io/docs/concepts/architecture/nodes/#condition)

### Setup Defaulting and complex Validation

- [Webhook docs](https://book.kubebuilder.io/beyond_basics/what_is_a_webhook.html)
