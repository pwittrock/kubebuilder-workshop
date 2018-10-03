# Kubebuilder Workshop

## Setup a dev environment

- Get a Kubernetes Cluster
  - Install [minikube](https://github.com/kubernetes/minikube)
  - Start a minikube cluster

- Install Go 1.10+
  - [go](https://golang.org/)

- Install dep
  - [dep](https://github.com/golang/dep)

- Install kubebuilder
  - [kubebuilder](https://book.kubebuilder.io/getting_started/installation_and_setup.html)
  
- Install kustomize
  - [kustomize](https://github.com/kubernetes-sigs/kustomize)

## Additional Resources

- [http://book.kubebuilder.io](http://book.kubebuilder.io)

## Create a new Go Project

Example:

- `mkdir -p $HOME/kubebuilder-workshop/src/github.com/my-org/my-project`
- `export GOPATH=$HOME/kubebuilder-workshop`
- `cd $HOME/kubebuilder-workshop/src/github.com/my-org/my-project`


## Initialize your Project

- `kubebuilder init --domain mydomain.io --license apache2 --owner "My Org"`
  - have it use dep to install the go dependencies

## Define an empty MongoDB API

- `kubebuilder create api --group databases --version v1alpha1 --kind MongoDB`
  - have it create both the Resource and the Controller
  
## Define the Schema for the MongoDB Resource

- `nano -w pkg/apis/databases/v1alpha/mongodb_types.go`

## Define the Implementation of the MongoDB Controller

- [Blogpost on running MongoDB as a StatefulSet on Kubernetes](https://kubernetes.io/blog/2017/01/running-mongodb-on-kubernetes-with-statefulsets/)
- `nano -w pkg/controller/mongodb/mongodb_controller.go`
- Change from Deployment to StatefulSet similar to one in Blogpost

## Install the Resource into your cluster, and run the Controller locally

- `make install` # install the CRDs
- `make run` # run the controller

## Edit a sample MongoDB instance and create it

- `nano -w config/samples/databases_v1alpha1_mongodb.yaml`
- `kubectl apply -f config/samples/databases_v1alpha1_mongodb.yaml`
- Observe output from Controller

## Bonus Objectives

- Build your API into a container and publish it
  - `IMG=foo make docker-build` && `IMG=foo make docker-push`
  - `kustomize build config/default > mongodb_api.yaml`
  - `kubectl apply -f mongodb_api.yaml`
  - Get logs from the Controller using `kubectl logs`
- Add Simple Schema Validation
  - [Validation tags docs](https://book.kubebuilder.io/basics/simple_resource.html)
- Publish Events from the Controller code
  - [Event docs](https://book.kubebuilder.io/beyond_basics/creating_events.html)
- Define and add a list of Conditions to the Status, then set them from the Controller
  - Example: [Node Conditions](https://kubernetes.io/docs/concepts/architecture/nodes/#condition)
- Setup Defaulting and complex Validation
  - [Webhook docs](https://book.kubebuilder.io/beyond_basics/what_is_a_webhook.html)
- Adopt objects if the parent is deleted / recreated without Garbage Collection