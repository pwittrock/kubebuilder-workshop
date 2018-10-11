/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package mongodb

import (
	"context"
	"log"
	"reflect"

	"fmt"
	databasesv1alpha1 "github.com/pwittrock/kubebuilder-workshop/pkg/apis/databases/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new MongoDB Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
// USER ACTION REQUIRED: update cmd/manager/main.go to call this databases.Add(mgr) to install this Controller
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileMongoDB{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
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

var _ reconcile.Reconciler = &ReconcileMongoDB{}

// ReconcileMongoDB reconciles a MongoDB object
type ReconcileMongoDB struct {
	client.Client
	scheme *runtime.Scheme
}

// CreateOrUpdate with either create / update an object, or if no changes are needed it will do nothing.
//
// It takes a function to determine if an update is required based on the expected generated object
// and the live read object.
//
// If the and update is needed, the function should update the *read* object with the changes from the
// *generated* object and return *false*.
func (r *ReconcileMongoDB) CreateOrUpdate(
	generated runtime.Object, copy func(generated, read runtime.Object) bool) (reconcile.Result, error) {

	meta, ok := generated.(metav1.Object)
	if !ok {
		// This should never happen
		return reconcile.Result{}, fmt.Errorf("invalid object type %T does not implement metav1.Object", generated)
	}

	read := generated.DeepCopyObject()
	err := r.Get(context.TODO(), types.NamespacedName{Name: meta.GetName(), Namespace: meta.GetNamespace()}, read)
	if err != nil && errors.IsNotFound(err) {
		log.Printf("Creating %T %s/%s\n", generated, meta.GetNamespace(), meta.GetName())
		err = r.Create(context.TODO(), generated)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Update the found object and write the result back if there are any changes
	if !copy(generated, read) {
		log.Printf("Updating %T %s/%s\n", generated, meta.GetNamespace(), meta.GetName())
		err = r.Update(context.TODO(), read)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

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
