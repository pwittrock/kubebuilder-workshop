This is not an officially supported Google product

# Kubebuilder Workshop

The Kubebuilder Workshop is aimed at providing hands-on experience creating Kubernetes APIs using kubebuilder.
By the end of the workshop, attendees will have created a Kubernetes native API for running MongoDB instances.

Once the API is installed into a Kubernetes cluster, cluster users should be able to create new MongoDB instances
similar to the one in [this blog post](https://kubernetes.io/blog/2017/01/running-mongodb-on-kubernetes-with-statefulsets/)
by specifying the MongoDB Resource in a file and running `kubectl apply -f`.

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

The MongoDB Resource will manage (Create / Update) a Kubernetes **StatefulSet** and **Service**.

**Note**: This repo contains a full solution to the workshop exercise if you get stuck.

## **Prerequisites**

**Important:** Complete these steps first.

**Note**: If you have been provided a pre-provisioned GCP project development account see the
[preprovisioned](https://github.com/pwittrock/kubebuilder-workshop-prereqs/preprovisioned/README.md) instructions.

[kubebuilder-workshop-prereqs](https://github.com/pwittrock/kubebuilder-workshop-prereqs)

## Overview

Following is an overview of the steps required to implement the MongoDB API.

1. Create the MongoDB Resource and Controller stubs.
1. Change the MongoDB Resource `MongoDBSpec` struct stub with a Schema.
1. Change the MongoDB Controller `add` function stub to Watch StatefulSets, Services, and MongoDBs.
1. Change the MongoDB controller `Reconcile` function stub to create / update StatefulSets and Services that
   run a MongoDB instance.

## Create the MongoDB Resource and Controller Stubs

**Note:** This will also build the project and run the tests to make sure the resource and controller are hooked up
correctly.

- `kubebuilder create api --group databases --version v1alpha1 --kind MongoDB`
  - enter `y` to have it create the stub for the Resource
  - enter `y` to have it create the stub for the Controller
  
### Step 1: Add a Schema to the MongoDB Resource stub

Change the MongoDB API Schema *Spec* in `pkg/apis/databases/v1alpha1/mongodb_types.go`.

Start with 2 optional fields:

- `replicas` (int32)
- `storage` (string)

**Note:** Simply copy the following Spec, and optionally revisit later to add more fields.

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

Documentation (Read later):

- [Resource Definition](https://book.kubebuilder.io/basics/simple_resource.html)
- [PodSpec Example](https://github.com/kubernetes/api/blob/master/core/v1/types.go#L2715)


### Step 2: Change Watches for the MongoDB Controller `add` function stub

Update the `add` function in `pkg/controller/mongodb/mongodb_controller.go` to Watch the Resources the
Controller will be managing.

- *No-Op* - Watch MongoDB (EnqueueRequestForObject) - this was scaffolded for you
- *Remove* - Watch Deployments - you aren't managing Deployments so remove this
- *Add* - Watch Services - and map to the Owning MongoDB instance (EnqueueRequestForOwner) - you are managing Services so add add this
- *Add* - Watch StatefulSets - and map to the Owning MongoDB instance (EnqueueRequestForOwner) - you are managing StatefulSets so add this

**Package Hints:**

- the `StatefulSet` struct is in package `appsv1 "k8s.io/api/apps/v1"`
- the `Service` struct is in package `corev1 "k8s.io/api/core/v1"`

See the following for documentation on Watches:

- [Simple Watch](https://book.kubebuilder.io/basics/simple_controller.html#adding-a-controller-to-the-manager)
- [Advanced Watch](https://book.kubebuilder.io/beyond_basics/controller_watches.html)

### Step 3: Change logic in the MongoDB Controller `Reconcile` function stub

Update the `Reconcile` function to Create / Update the StatefulSet and Service objects to run MongoDB in
`pkg/controller/mongodb/mongodb_controller.go`.

The generated Reconcile stub manages (creates or updates) a static Deployment.  Instead of managing a Deployment,
we want to manage a StatefulSet and a Service using `util` package to create the structs.

Steps:

- Change the code that Creates / Updates a Deployment to Create / Update a Service using the `GenerateService` and `CopyServiceFields` functions
- Copy the code to also Create / Update a StatefulSet using the `GenerateStatefulSet` and `CopyStatefulSetFields` functions
- Run `make` (expect tests to fail because they have not been updated)

**Object Generation Hints:**

- Make sure you have the cloned or copied the provided
  "[github.com/pwittrock/kubebuilder-workshop-prereqs/pkg/util](https://github.com/pwittrock/kubebuilder-workshop-prereqs/blob/master/pkg/util/util.go)" functions
- Use the functions under `pkg/util` to provide StatefulSet and Service struct instances for a MongoDB instance
  - `GenerateStatefulSet(mongo metav1.Object, replicas *int32, storage *string) *appsv1.StatefulSet`
  - `GenerateService(mongo metav1.Object) *corev1.Service`
- Use the functions under `github.com/pwittrock/kubebuilder-workshop-prereqs/pkg/util` to copy fields from generated StatefulSet and Service struct instances
  to the read instances (e.g. so you don't clobber the Service.Spec.ClusterIP field).
  - `CopyStatefulSetFields(from, to *appsv1.StatefulSet) bool` // Returns true if update required
  - `CopyServiceFields(from, to *corev1.Service) bool` // Returns true if update requireds

**Note:** This will cause the tests to start failing because you changed the Reconcile behavior.  Don't worry
about this for now.

Documentation:

- [Reconcile](https://book.kubebuilder.io/basics/simple_controller.html#implementing-controller-reconcile)

#### RBAC

- **Optional:** for running in cluster only - update the RBAC rules defined as comments on the `Reconcile` function
  to give read / write access for StatefulSets and Services (required when running as a container in a cluster).
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
- requires *[updating the RBAC rules](#rbac)*
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
