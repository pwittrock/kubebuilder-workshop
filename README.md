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

To make them optional:

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

- quickly skim the [blogpost on running MongoDB as a StatefulSet on Kubernetes](https://kubernetes.io/blog/2017/01/running-mongodb-on-kubernetes-with-statefulsets/)
  for background information - don't use the actual StatefulSet and Service, we will do something different.
- edit `pkg/controller/mongodb/mongodb_controller.go`
- remove the Deployment creation code
- replace with StatefulSet and Service creation code
- *do not* copy the labels from the blogpost directly, use something based off the instance name
- optional: update labels and selectors to be more correct
- update RBAC annotations on the Reconcile function with StatefuleSet and Service

### Update the Watch config

- Watch MongoDB (generated for you)
- Watch Services - and map to the Owning MongoDB instance (because we will generate them)
- Watch StatefulSets - and map to the Owning MongoDB instance (because we will generate them)
- (Delete Watch for Deployments because we don't generate them)

Try to do it on your own first - but it should look something like this:

```go
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("mongodb-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &databasesv1alpha1.MongoDB{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &appsv1.StatefulSet{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &databasesv1alpha1.MongoDB{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &databasesv1alpha1.MongoDB{},
	})
	if err != nil {
		return err
	}

	return nil
}
```

### Generate a StatefulSet from the MongoDB instance

Copy this helper function to save time instead of writing it yourself.

This function creates a new appsv1.StatefulSet Datastructure with the Name+Namespace, Labels, Selector,
PodTemplate, ServiceName and Replicas set.


```go
func GetStatefuleSet(instance *databasesv1alpha1.MongoDB) *appsv1.StatefulSet {
    gracePeriodTerm := int64(10)

    // TODO: Default and Validate these with Webhooks
    if instance.Spec.Replicas == nil {
        r := int32(1)
        instance.Spec.Replicas = &r
    }
    if instance.Spec.Storage == nil {
        s := "100Gi"
        instance.Spec.Storage = &s
    }
    if instance.Labels == nil {
        instance.Labels = map[string]string{}
    }

    labels := map[string]string{}
    for k, v := range instance.Labels {
        labels[k] = v
    }
    labels["mongodb-statefuleset"] = instance.Name

    rl := corev1.ResourceList{}
    rl["storage"] = resource.MustParse(*instance.Spec.Storage)

    pvc := corev1.PersistentVolumeClaim{
        ObjectMeta: metav1.ObjectMeta{Name: "mongo-persistent-storage"},
        Spec: corev1.PersistentVolumeClaimSpec{
            AccessModes: []corev1.PersistentVolumeAccessMode{"ReadWriteOnce"},
            Resources: corev1.ResourceRequirements{
                Requests: rl,
            },
        },
    }

    stateful := &appsv1.StatefulSet{
        ObjectMeta: metav1.ObjectMeta{
            Name:      instance.Name + "-mongodb-statefulset",
            Namespace: instance.Namespace,
            Labels:    labels,
        },
        Spec: appsv1.StatefulSetSpec{
            Selector: &metav1.LabelSelector{
                MatchLabels: map[string]string{"statefulset": instance.Name + "-mongodb-statefulset"},
            },
            ServiceName: "mongo",
            Replicas:    instance.Spec.Replicas,
            Template: corev1.PodTemplateSpec{
                ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"statefulset": instance.Name + "-mongodb-statefulset"}},

                Spec: corev1.PodSpec{
                    TerminationGracePeriodSeconds: &gracePeriodTerm,
                    Containers: []corev1.Container{
                        {
                            Name:         "mongo",
                            Image:        "mongo",
                            Command:      []string{"mongod", "--replSet", "rs0", "--smallfiles", "--noprealloc", "--bind_ip_all"},
                            Ports:        []corev1.ContainerPort{{ContainerPort: 27017}},
                            VolumeMounts: []corev1.VolumeMount{{Name: "mongo-persistent-storage", MountPath: "/data/db"}},
                        },
                        {
                            Name:  "mongo-sidecar",
                            Image: "cvallance/mongo-k8s-sidecar",
                            Env:   []corev1.EnvVar{{Name: "MONGO_SIDECAR_POD_LABELS", Value: "role=mongo,environment=test"}},
                        },
                    },
                },
            },
            VolumeClaimTemplates: []corev1.PersistentVolumeClaim{pvc},
        },
    }
    return stateful
}
```

### Generate a Service from the MongoDB instance

Copy this helper function to save time instead of writing it yourself.

This function creates a new corev1.Service Datastructure with the Name+Namespace, Labels, Selector and Port set.

```go
// GetService returns the desired generated Service for the MongoDB instance
func GetService(instance *databasesv1alpha1.MongoDB) *corev1.Service {
	// TODO: Default and Validate these with Webhooks
	if instance.Labels == nil {
		instance.Labels = map[string]string{}
	}
	labels := map[string]string{}
	for k, v := range instance.Labels {
		labels[k] = v
	}
	labels["mongodb-service"] = instance.Name

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-mongodb-service",
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{Port: 27017, TargetPort: intstr.IntOrString{IntVal: 27017, Type: intstr.Int}},
			},
			Selector: map[string]string{"statefulset": instance.Name + "-mongodb-statefulset"},
		},
	}
	return service
}
```

### Create or Update the generated Objects

Copy this helper function to reduce boilerplate.  It takes a function to determine if the generated
object matches the live copy of the generated object, and copy the spec to the live object if they do not match.

```go
func (r *ReconcileMongoDB) CreateOrUpdate(
	object runtime.Object, copy func(o1, o2 runtime.Object) bool) (reconcile.Result, error) {

	meta, ok := object.(metav1.Object)
	if !ok {
		// This should never happen
		return reconcile.Result{}, fmt.Errorf("invalid object type %T does not implement metav1.Object", object)
	}

	lookup := object.DeepCopyObject()
	err := r.Get(context.TODO(), types.NamespacedName{Name: meta.GetName(), Namespace: meta.GetNamespace()}, lookup)
	if err != nil && errors.IsNotFound(err) {
		log.Printf("Creating %T %s/%s\n", object, meta.GetNamespace(), meta.GetName())
		err = r.Create(context.TODO(), object)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Update the found object and write the result back if there are any changes
	if !copy(object, lookup) {
		log.Printf("Updating %T %s/%s\n", object, meta.GetNamespace(), meta.GetName())
		err = r.Update(context.TODO(), lookup)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}
```


### Controller Code

- Update the RBAC rules to give perms for StatefulSets and Services
- Generate a Service
- Use the Helper CreateOrUpdate to update the Service
- Generate a StatefuleSet
- Use the Helper CreateOrUpdate to update the StatefulSet

```go
// +kubebuilder:rbac:groups=apps,resources=statefulesets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=databases.k8s.io,resources=mongodbs,verbs=get;list;watch;create;update;patch;delete
func (r *ReconcileMongoDB) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the MongoDB instance
	instance := &databasesv1alpha1.MongoDB{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Create or Update the Service
	service := GetService(instance)
	if err := controllerutil.SetControllerReference(instance, service, r.scheme); err != nil {
		return reconcile.Result{}, err
	}
	if r, err := r.CreateOrUpdate(service, func(generated, fetched runtime.Object) bool {
		generatedS := generated.(*corev1.Service)
		fetchedS := fetched.(*corev1.Service)
		if reflect.DeepEqual(generatedS.Spec.Selector, fetchedS.Spec.Selector) &&
			reflect.DeepEqual(generatedS.Spec.Ports, fetchedS.Spec.Ports) {
			// Don't update
			return true
		}
		fetchedS.Spec.Selector = generatedS.Spec.Selector
		fetchedS.Spec.Ports = generatedS.Spec.Ports
		// Update
		return false
	}); err != nil {
		log.Printf("failed to Create or Update Service %s/%s %v", service.Namespace, service.Name, err)
		return r, err
	}

	// Create or Update the StatefulSet
	stateful := GetStatefuleSet(instance)
	if err := controllerutil.SetControllerReference(instance, stateful, r.scheme); err != nil {
		return reconcile.Result{}, err
	}
	if r, err := r.CreateOrUpdate(stateful, func(generated, fetched runtime.Object) bool {
		generatedS := generated.(*appsv1.StatefulSet)
		fetchedS := fetched.(*appsv1.StatefulSet)
		if reflect.DeepEqual(generatedS.Spec, fetchedS.Spec) {
			// Don't update
			return true
		}
		// TODO: Compare only fields we own
		fetchedS.Spec = generatedS.Spec
		// Update
		return false
	}); err != nil {
		log.Printf("failed to Create or Update StatefulSet %s/%s %v", stateful.Namespace, stateful.Name, err)
		return r, err
	}

	return reconcile.Result{}, nil
}
```

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
