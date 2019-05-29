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
	"github.com/pwittrock/kubebuilder-workshop/api/v1alpha1"
	"github.com/pwittrock/kubebuilder-workshop/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// MongoDBReconciler reconciles a MongoDB object
type MongoDBReconciler struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

// +kubebuilder:rbac:groups=databases.example.com,resources=mongodbs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=databases.example.com,resources=mongodbs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete

func (r *MongoDBReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("mongodb", req.NamespacedName)

	// Fetch the MongoDB instance
	mongo := &v1alpha1.MongoDB{}
	if err := r.Get(ctx, req.NamespacedName, mongo); err != nil {
		log.Error(err, "unable to fetch MongoDB")
		if apierrs.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Generate Service
	service := &corev1.Service{
		ObjectMeta: ctrl.ObjectMeta{
			Name:      req.Name + "-mongodb-service",
			Namespace: req.Namespace,
		},
	}
	_, err := ctrl.CreateOrUpdate(ctx, r.Client, service, func() error {
		util.SetServiceFields(service, mongo)
		return controllerutil.SetControllerReference(mongo, service, r.Scheme)
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	// Generate StatefulSet
	ss := &appsv1.StatefulSet{
		ObjectMeta: ctrl.ObjectMeta{
			Name:      req.Name + "-mongodb-statefulset",
			Namespace: req.Namespace,
		},
	}
	_, err = ctrl.CreateOrUpdate(ctx, r.Client, ss, func() error {
		util.SetStatefulSetFields(ss, service, mongo, mongo.Spec.Replicas, mongo.Spec.Storage)
		return controllerutil.SetControllerReference(mongo, ss, r.Scheme)

	})
	if err != nil {
		return ctrl.Result{}, err
	}

	// Update Status
	ssNN := req.NamespacedName
	ssNN.Name = ss.Name
	if err := r.Get(ctx, ssNN, ss); err != nil {
		log.Error(err, "unable to fetch StatefulSet", "namespaceName", ssNN)
		return ctrl.Result{}, err
	}
	mongo.Status.StatefulSetStatus = ss.Status

	serviceNN := req.NamespacedName
	serviceNN.Name = service.Name
	if err := r.Get(ctx, serviceNN, service); err != nil {
		log.Error(err, "unable to fetch Service", "namespaceName", serviceNN)
		return ctrl.Result{}, err
	}
	mongo.Status.ServiceStatus = service.Status

	err = r.Status().Update(ctx, mongo)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *MongoDBReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.MongoDB{}).
		Owns(&appsv1.StatefulSet{}). // Generates StatefulSets
		Owns(&corev1.Service{}).     // Generates Services
		Complete(r)
}
