/*
Copyright 2018 The Kubernetes authors.

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
	databasesv1alpha1 "github.com/pwittrock/kubebuilder-workshop/pkg/apis/databases/v1alpha1"
	"github.com/pwittrock/kubebuilder-workshop/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"log"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

/**
 * This is the completed workshop solution.
 */

// Add creates a new MongoDB Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
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

	// Watch for changes to MongoDB
	err = c.Watch(&source.Kind{Type: &databasesv1alpha1.MongoDB{}}, &handler.EnqueueRequestForObject{})
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

	err = c.Watch(&source.Kind{Type: &appsv1.StatefulSet{}}, &handler.EnqueueRequestForOwner{
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

// Reconcile reads that state of the cluster for a MongoDB object and makes changes based on the state read
// and what is in the MongoDB.Spec
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
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

	// Define the desired Service object
	service := util.GenerateService(instance)
	if err := controllerutil.SetControllerReference(instance, service, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if the Service already exists
	foundService := &corev1.Service{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, foundService)
	if err != nil && errors.IsNotFound(err) {
		log.Printf("Creating Service %s/%s\n", service.Namespace, service.Name)
		err = r.Create(context.TODO(), service)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Update the foundService object and write the result back if there are any changes
	if util.CopyServiceFields(service, foundService) {
		log.Printf("Updating Service %s/%s\n", service.Namespace, service.Name)
		err = r.Update(context.TODO(), foundService)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	// Define the desired StatefulSet object
	stateful := util.GenerateStatefulSet(instance, instance.Spec.Replicas, instance.Spec.Storage)
	if err := controllerutil.SetControllerReference(instance, stateful, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if the StatefulSet already exists
	foundStateful := &appsv1.StatefulSet{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: stateful.Name, Namespace: stateful.Namespace}, foundStateful)
	if err != nil && errors.IsNotFound(err) {
		log.Printf("Creating StatefulSet %s/%s\n", stateful.Namespace, stateful.Name)
		err = r.Create(context.TODO(), stateful)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Update the foundStateful object and write the result back if there are any changes
	if util.CopyStatefulSetFields(stateful, foundStateful) {
		log.Printf("Updating StatefulSet %s/%s\n", stateful.Namespace, stateful.Name)
		err = r.Update(context.TODO(), foundStateful)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}
