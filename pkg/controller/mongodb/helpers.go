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
	databasesv1alpha1 "github.com/pwittrock/kubebuilder-workshop/pkg/apis/databases/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// GenerateStatefuleSet returns a new appsv1.StatefulSet pointer generated for the MongoDB instance
func GenerateStatefuleSet(instance *databasesv1alpha1.MongoDB) *appsv1.StatefulSet {
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

// GenerateService returns a new corev1.Service pointer generated for the MongoDB instance
func GenerateService(instance *databasesv1alpha1.MongoDB) *corev1.Service {
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
