This is not an officially supported Google product

# Kubebuilder Workshop

The Kubebuilder Workshop is aimed at providing hands-on experience building Kubernetes APIs using kubebuilder.
By the end of the workshop, participants will have built a Kubernetes native API for running MongoDB instances.

Once the API is installed into a Kubernetes cluster, users should be able to create new MongoDB instance similar
to the one in [this blog post](https://kubernetes.io/blog/2017/01/running-mongodb-on-kubernetes-with-statefulsets/) by
specifying a MongoDB file and running `kubectl apply -f` on it.

The MongoDB API will manage (Create / Update) a Kubernetes **StatefulSet** and **Service** that runs a MongoDB instance.

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

**Note**: This repo contains a full solution for the workshop exercise if you get stuck

## **Prerequisites**

**Important:** Do these steps before moving on

[kubebuilder-workshop-prereqs](https://github.com/pwittrock/kubebuilder-workshop-prereqs)

## Overview

Following is an overview of the steps required to implement the MongoDB API.

1. Create the scaffolded boilerplate for the MongoDB Resource and Controller
1. Update the MongoDB Resource scaffold with a Schema
1. Update the MongoDB Controller `add` stub to Watch StatefulSets, Services, and MongoDBs
1. Update the MongoDB controller `Reconcile` stub to create / update StatefulSets and Services

## Scaffold the boilerplate for the MongoDB Resource and Controller

Scaffold the boilerplate for a new MongoDB Resource type and Controller

**Note:** This will also build the project and run the tests to make sure the resource and controller are hooked up
correctly.

- `kubebuilder create api --group databases --version v1alpha1 --kind MongoDB`
  - enter `y` to have it create boilerplate for the Resource
  - enter `y` to have it create boilerplate for the Controller
  
## Update the MongoDB Resource scaffold with a Schema

Define the MongoDB API Schema *Spec*for in `pkg/apis/databases/v1alpha/mongodb_types.go`.

Spec contains 2 optional fields:

- `replicas` (int32)
- `storage` (string)

**Note:** Copy the following Spec, optionally revisit later to add more fields.

```go
type MongoDBSpec struct {
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// +optional
	Storage *string `json:"storage,omitempty"`
}
```

**Note:** Fields have been made optional by:

- setting `// +optional`
- making them pointers with `*`
- adding the `omitempty` struct tag

Documentation:

- [Resource Definition](https://book.kubebuilder.io/basics/simple_resource.html)
- [PodSpec Example](https://github.com/kubernetes/api/blob/master/core/v1/types.go#L2715)

## Implement the MongoDB Controller

The MongoDB Controller should manage a StatefulSet and Service for running MongoDB.

### Update `add` Watches

Update the `add` function to Watch the Resources you will be creating / updating in
`pkg/controller/mongodb/mongodb_controller.go`.

- *No-Op* - Watch MongoDB (EnqueueRequestForObject) - this was scaffolded for you
- *Remove* - Watch Deployments - you aren't managing Deployments so remove this
- *Add* - Watch Services - and map to the Owning MongoDB instance (EnqueueRequestForOwner) - you are managing Services so add add this
- *Add* - Watch StatefulSets - and map to the Owning MongoDB instance (EnqueueRequestForOwner) - you are managing StatefulSets so add this

**Package Hints:**

- for the StatefulSet struct use package - `appsv1 "k8s.io/api/apps/v1"`
- for the Services struct use package - `corev1 "k8s.io/api/core/v1"`

Documentation:

- [Simple Watch](https://book.kubebuilder.io/basics/simple_controller.html#adding-a-controller-to-the-manager)
- [Advanced Watch](https://book.kubebuilder.io/beyond_basics/controller_watches.html)

### Update `Reconcile` with object creation

Update the `Reconcile` function to Create / Update the StatefulSet and Service objects to run MongoDB in
`pkg/controller/mongodb/mongodb_controller.go`.

- Create a Service for running MongoDB, or Update the existing one
- Create a StatefulSet for running MongoDB, or Update the existing one

**Object Generation Hints:**

- Make sure you have the cloned or copied the provided
  [pkg/util](https://github.com/pwittrock/kubebuilder-workshop-prereqs/blob/master/pkg/util/util.go) functions
- Use the functions under `pkg/util` to provide StatefulSet and Service struct instances for a MongoDB instance
  - `GenerateStatefuleSet(mongo metav1.Object, replicas *int32, storage *string) *appsv1.StatefulSet`
  - `GenerateService(mongo metav1.Object) *corev1.Service`
- Use the functions under `pkg/util` to copy fields from generated StatefulSet and Service struct instances
  to the live copies that have been read (e.g. so you don't clobber the Service.Spec.ClusterIP field).
  - `CopyStatefulSetFields(from, to *appsv1.StatefulSet) bool`
  - `CopyServiceFields(from, to *corev1.Service) bool`

**Note:** This will cause the tests to start failing because you changed the Reconcile behavior.  Don't worry
about this for now.

Documentation:

- [Reconcile](https://book.kubebuilder.io/basics/simple_controller.html#implementing-controller-reconcile)

- **Optional:** for when running in cluster - update the RBAC rules to give perms for StatefulSets and
  Services (needed for if running as a container in a cluster)
  - `// +kubebuilder:rbac:groups=apps,resources=statefulesets,verbs=get;list;watch;create;update;patch;delete`
  - `// +kubebuilder:rbac:groups=,resources=services,verbs=get;list;watch;create;update;patch;delete`
  - `// +kubebuilder:rbac:groups=databases.k8s.io,resources=mongodbs,verbs=get;list;watch;create;update;patch;delete`

## Try your API in a Kubernetes Cluster

Now that you have finished implementing the MongoDB API, lets try it out in a Kubernetes cluster.

### Install the Resource into the Cluster

- `make install` # install the CRDs

### Run the Controller locally

- `make run` # run the controller as a local process

### Edit the sample MongoDB file

Edit `config/samples/databases_v1alpha1_mongodb.yaml`

```yaml
apiVersion: databases.k8s.io/v1alpha1
kind: MongoDB
metadata:
  name: mongo-instance
spec:
  replicas: 1
  storage: 100Gi
```

- create the mongodb instance
  - `kubectl apply -f config/samples/databases_v1alpha1_mongodb.yaml`
  - observe output from Controller

### Check out the Resources in the cluster

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
- recreate the MongoDB instance
  - `kubectl apply -f config/samples/databases_v1alpha1_mongodb.yaml`

### Connect to the running MongoDB instance from within the cluster using a Pod

- `kubectl run mongo-test -t -i --rm --image mongo bash`
- `mongo <cluster ip address of mongodb service>:27017`

## Experiment some more

- Try deleting the StatefulSet - what happens when you look for it?
- Try deleting the Service - what happens when you look for it?
- Try adding fields to control new things such as the Port

## Bonus Objectives

If you finish early, or want to continue working on your API after the workshop, try these exercises.

### Build your Controller into a container and host it in the cluster

- requires [kustomize](https://github.com/kubernetes-sigs/kustomize)
- `IMG=foo make docker-build` && `IMG=foo make docker-push`
- `kustomize build config/default > mongodb_api.yaml`
- `kubectl apply -f mongodb_api.yaml`
- Get logs from the Controller using `kubectl logs`

### Add Simple Schema Validation for field values

- Add validation tags to the struct fields as annotation comments
    - [Validation tags docs](https://book.kubebuilder.io/basics/simple_resource.html)

### Publish Events and update Status from `Reconcile`

- [Event docs](https://book.kubebuilder.io/beyond_basics/creating_events.html)
- Use `kubectl describe` to view the events
- Add Status fields
  - Define list of Conditions (e.g. [Node Conditions](https://kubernetes.io/docs/concepts/architecture/nodes/#condition))

### Configure the `scale` endpoint

- Use kustomize to patch the generated crd with the `scale` endpoint
  - [docs](https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/#scale-subresource)

### Add more to the API

Allow further customization of what gets generated by adding more fields to the Resource type

- ports
- bind-ips
- [side-car options](https://github.com/cvallance/mongo-k8s-sidecar)
  - namespace (KUBE_NAMESPACE)
  - setup username / password from a secret (MONGODB_USERNAME / MONGODB_PASSWORD)
  - ssl (MONGO_SSL_ENABLED)
  - stable ips (KUBERNETES_MONGO_SERVICE_NAME)

### Update the tests to make them pass

Update the tests to check the Controller logic you added
  - verify creation of StatefulSet and Service
  - verify update of StatefulSet and Service

### Add more Operational Logic to the Reconcile

Add logic to the Reconcile to handle MongoDB lifecycle events such as upgrades / downgrades.

### Setup Defaulting and complex Validation

- [Webhook docs](https://book.kubebuilder.io/beyond_basics/what_is_a_webhook.html)
