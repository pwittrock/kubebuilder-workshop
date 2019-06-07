/*
Copyright 2019 The Kubernetes authors.

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

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	databasesv1alpha1 "github.com/pwittrock/kubebuilder-workshop/api/v1alpha1"
	"github.com/pwittrock/kubebuilder-workshop/util"
)

// MongoDBReconciler reconciles a MongoDB object
type MongoDBReconciler struct {
	Scheme *runtime.Scheme
	client.Client
	Log logr.Logger
}

// +kubebuilder:rbac:groups=databases.example.com,resources=mongodbs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=databases.example.com,resources=mongodbs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete

func (r *MongoDBReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("mongodb", req.NamespacedName)

	mongo := &databasesv1alpha1.MongoDB{}
	err := r.Get(ctx, req.NamespacedName, mongo)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Create the Service for mongodb
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name + "mongodb",
			Namespace: req.Namespace,
		},
	}
	_, err = ctrl.CreateOrUpdate(ctx, r.Client, service, func() error {
		util.SetServiceFields(service, mongo)
		ctrl.SetControllerReference(mongo, service, r.Scheme)
		return nil
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	// Create the StatefulSet for mongodb
	stateful := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name + "mongodb",
			Namespace: req.Namespace,
		},
	}
	_, err = ctrl.CreateOrUpdate(ctx, r.Client, stateful, func() error {
		util.SetStatefulSetFields(stateful, service, mongo, mongo.Spec.Replicas, mongo.Spec.Storage)
		ctrl.SetControllerReference(mongo, stateful, r.Scheme)
		return nil
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	// err = r.Get(ctx, types.NamespacedName{Namespace: stateful.Namespace, Name: stateful.Name}, stateful)
	// if err != nil {
	// 	return ctrl.Result{}, err
	// }
	// err = r.Get(ctx, types.NamespacedName{Namespace: service.Namespace, Name: service.Name}, service)
	// if err != nil {
	// 	return ctrl.Result{}, err
	// }

	// Update MongoStatus
	mongo.Status.StatefulSetStatus = stateful.Status
	mongo.Status.ServiceStatus = service.Status
	mongo.Status.ClusterIP = service.Spec.ClusterIP
	err = r.Status().Update(ctx, mongo)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *MongoDBReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&databasesv1alpha1.MongoDB{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
