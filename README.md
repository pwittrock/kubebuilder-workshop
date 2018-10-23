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

- If you have been given a workshop GCP account with an already provisioned dev machine
  - See [preprovisioned](https://github.com/pwittrock/kubebuilder-workshop-prereqs/blob/master/preprovisioned/README.md#instructions-for-using-a-provisioned-workshop-development-machine)
    instructions.

- If you are setting up a local development environment (e.g. laptop)
  - See [kubebuilder-workshop-prereqs](https://github.com/pwittrock/kubebuilder-workshop-prereqs) instructions.

- If you are using GCP Deployment Manager to setup a cloud dev machine
  - See [DM](https://github.com/pwittrock/kubebuilder-workshop-prereqs/blob/master/dm-setup/README.md#instructions-for-getting-a-development-machine)
    instructions.

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

#### Break Glass
 
If you get stuck - see the completed solution for this step [here](https://github.com/pwittrock/kubebuilder-workshop/blob/master/pkg/apis/databases/v1alpha1/mongodb_types.go#L28)

##### Instructions

Change the MongoDB API Schema (e.g. *MongoDBSpec*) in `pkg/apis/databases/v1alpha1/mongodb_types.go`.

Start with 2 optional fields:

- `replicas` (int32)
- `storage` (string)

**Note:** Simply update the stubbed MongoDBSpec with the following code.

```go
type MongoDBSpec struct {
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// +optional
	Storage *string `json:"storage,omitempty"`
}
```

##### Optional Fields

Optional fields are defined by:

- setting `// +optional`
- making them pointers with `*`
- adding the `omitempty` struct tag

##### Documentation (Read later)

- [Resource Definition](https://book.kubebuilder.io/basics/simple_resource.html)
- [PodSpec Example](https://github.com/kubernetes/api/blob/master/core/v1/types.go#L2715)

### Step 2: Change Watches for the MongoDB Controller `add` function stub

#### Break Glass

If you get stuck - see the completed solution for this step [here](https://github.com/pwittrock/kubebuilder-workshop/blob/master/pkg/controller/mongodb/mongodb_controller.go#L67)

#### Instructions

Update the `add` function in `pkg/controller/mongodb/mongodb_controller.go` to Watch the Resources the
Controller will be managing.

The generated `add` stub watches Deployments *owned* by the Controller.  Instead we want to watch StatefulSets
and Services *owned* by the Controller.  Modify / copy the `Watch` configuration for Deployment.

1. *Add* - Watch Services - and map to the Owning MongoDB instance (using `EnqueueRequestForOwner`)
1. *Add* - Watch StatefulSets - and map to the Owning MongoDB instance (using `EnqueueRequestForOwner`)
1. *Remove* - Watch Deployments - you aren't managing Deployments so remove this
1. *No-Op* - Watch MongoDB (EnqueueRequestForObject) - this was stubbed for you

##### Package Hints

- the `StatefulSet` struct is in package `appsv1 "k8s.io/api/apps/v1"`
- the `Service` struct is in package `corev1 "k8s.io/api/core/v1"`

See the following for documentation on Watches:

- [Simple Watch](https://book.kubebuilder.io/basics/simple_controller.html#adding-a-controller-to-the-manager)
- [Advanced Watch](https://book.kubebuilder.io/beyond_basics/controller_watches.html)

### Step 3: Change logic in the MongoDB Controller `Reconcile` function stub

#### Break Glass

If you get stuck - see the completed solution for this step [here](https://github.com/pwittrock/kubebuilder-workshop/blob/master/pkg/controller/mongodb/mongodb_controller.go#L114)

**Important:** The break-glass link has a different `util` import than you will use
(kubebuilder-workshop instead of kubebuilder-workshop-prereqs).

#### Instructions

Update the `Reconcile` function to Create / Update the StatefulSet and Service objects to run MongoDB in
`pkg/controller/mongodb/mongodb_controller.go`.

The generated Reconcile stub manages (creates or updates) a static Deployment.  Instead of managing a Deployment,
we want to manage a StatefulSet and a Service using `util` package to create the structs.

##### Overview

1. Compute (generate) the desired struct instance of the Service / StatefulSet
1. Check if the Service / StatefulSet already exists by trying to read it
1. Either
  - Create the Service / StatefulSet if it doesn't exist
  - Compare the desired (generated) Service / StatefulSet to the read instance
  - If they do not match - copy the fields to the read instance and update using the read instance

##### Steps

1. Update the *import* statement at the top of the file by adding `"github.com/pwittrock/kubebuilder-workshop-prereqs/pkg/util"`
   (if you cloned the kubebuilder-workshop-prereqs project you should able to see them under `pkg/util`)
  - **Note:** If you did not clone the kubebuilder-workshop-prereqs project, *and* you are not using a preprovisioned
    GCE development VM, you will have needed to copy [these functions](https://github.com/pwittrock/kubebuilder-workshop-prereqs/blob/master/pkg/util/util.go)
    into your project - as described in the [prereqs](https://github.com/pwittrock/kubebuilder-workshop-prereqs#create-a-new-kubebuilder-project-pick-1).
1. Change the code that Creates / Updates a Deployment to Create / Update a Service using the `GenerateService` and `CopyServiceFields` functions
  - Use `util.GenerateService(mongo metav1.Object) *corev1.Service` to create the service struct (instead of `deploy := &appsv1.Deployment{...}`)
  - Use `util.CopyServiceFields(from, to *corev1.Service) bool` to check if we need to update the object and copy the
    fields (instead of `reflect.DeepEquals` and `found.Spec = deploy.Spec`)
1. Copy the code to also Create / Update a StatefulSet using the `GenerateStatefulSet` and `CopyStatefulSetFields` functions
  - Use `util.GenerateStatefulSet(mongo metav1.Object, replicas *int32, storage *string) *appsv1.StatefulSet`
  - Use `util.CopyStatefulSetFields(from, to *appsv1.StatefulSet) bool`
1. Delete unused imports `reflect` and `metav1`
1. Run `make` (expect tests to fail because they have not been updated)

**Note:** This will cause the tests to start failing because you changed the Reconcile behavior.  Don't worry
about this for now.

##### Documentation

- [Reconcile](https://book.kubebuilder.io/basics/simple_controller.html#implementing-controller-reconcile)

##### RBAC

- **Optional:** for running in cluster only - update the RBAC rules defined as comments on the `Reconcile` function
  to give read / write access for StatefulSets and Services (required when running as a container in a cluster).
  - `// +kubebuilder:rbac:groups=apps,resources=statefulesets,verbs=get;list;watch;create;update;patch;delete`
  - `// +kubebuilder:rbac:groups=,resources=services,verbs=get;list;watch;create;update;patch;delete`
  - `// +kubebuilder:rbac:groups=databases.k8s.io,resources=mongodbs,verbs=get;list;watch;create;update;patch;delete`

## Try your API in a Kubernetes Cluster

Now that you have finished implementing the MongoDB API, lets try it out in a Kubernetes cluster.

**Remember:** The tests will have started to fail because you changed the Reconcile behavior in Step 3.
Don't worry about this for now.

### Install the Resource into the Cluster

- `make install` # install the CRDs

### Run the Controller locally

- `make run` # run the controller as a local process
  - **Note:** this will not return, you will need to open another
    shell or send it to the background for the next steps.

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
  - `kubectl get mongodbs,statefulsets,services,pods` (no spaces)
    - **note**: the containers may be creating - wait for them to come up
  - `kubectl describe pods`
  - `kubectl logs mongo-instance-mongodb-statefulset-0 mongo`

### Connect to the running MongoDB instance from within the cluster using a Pod

- `kubectl run mongo-test -t -i --rm --image mongo bash`
- `mongo <cluster ip address of mongodb service>:27017`

### Verify Garbage Collection is working

- delete the mongodb instance
  - `kubectl delete -f config/samples/databases_v1alpha1_mongodb.yaml`
- look for garbage collected resources (they should be gone)
  - `kubectl get monogodbs`
  - `kubectl get statefulsets`
  - `kubectl get services`
  - `kubectl get pods`
- recreate the MongoDB instance
  - `kubectl apply -f config/samples/databases_v1alpha1_mongodb.yaml`

## Experiment some more

- Try deleting the StatefulSet - what happens when you look for it?
- Try deleting the Service - what happens when you look for it?
- Try adding fields to control new things such as the Port

## Feedback

- [Workshop Feedback](https://goo.gl/vgNvV1)

## Bonus Objectives

If you finish early, or want to continue working on your API after the workshop, try these exercises.

### Run the Controller in the cluster

Build your Controller into a container and host it on the cluster itself.

- requires installing [kustomize](https://github.com/kubernetes-sigs/kustomize/releases)
- requires installing [docker](https://docs.docker.com/install/)
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
