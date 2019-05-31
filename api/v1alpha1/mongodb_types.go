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

package v1alpha1

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MongoDBSpec defines the desired state of MongoDB
type MongoDBSpec struct {

	// replicas is the number of MongoDB StatefulSet replicas to run
	// +kubebuilder:validation:Minimum=1
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// storage is the volume size for each MongoDB replica
	// +optional
	Storage *string `json:"storage,omitempty"`
}

// MongoDBStatus defines the observed state of MongoDB
type MongoDBStatus struct {
	ServiceStatus corev1.ServiceStatus `json:"serviceStatus"`

	StatefulSetStatus appsv1.StatefulSetStatus `json:"statefulSetStatus"`

	ClusterIP string `json:"clusterIP,omitempty"`
}

// +kubebuilder:printcolumn:name="storage",type="string",JSONPath=".spec.storage",format="byte"
// +kubebuilder:printcolumn:name="replicas",type="string",JSONPath=".spec.replicas",format="byte"
// +kubebuilder:printcolumn:name="ready replicas",type="string",JSONPath=".status.statefulSetStatus.readyReplicas",format="byte"
// +kubebuilder:printcolumn:name="current replicas",type="string",JSONPath=".status.statefulSetStatus.currentReplicas",format="byte"
// +kubebuilder:printcolumn:name="cluster-ip",type="string",JSONPath=".status.clusterIP",format="byte"
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.statefulSetStatus.replicas
// +kubebuilder:object:root=true

// MongoDB is the Schema for the mongodbs API
type MongoDB struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MongoDBSpec   `json:"spec,omitempty"`
	Status MongoDBStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MongoDBList contains a list of MongoDB
type MongoDBList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MongoDB `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MongoDB{}, &MongoDBList{})
}

// // +kubebuilder:webhook:failurePolicy=fail,groups=databases.example.com,resources=mongodbs,verbs=create;update,versions=v1alpha1,name=mongodbs.example.com,path=/validate-example-com-v1alpha1-mongodbs,mutating=true
//
// // ValidateCreate implements webhookutil.validator so a webhook will be registered for the type
// func (c *MongoDB) ValidateCreate() error {
// 	if _, err := resource.ParseQuantity(*c.Spec.Storage); err != nil {
// 		return fmt.Errorf(".spec.stroage by a parseable quantity")
// 	}
// 	return nil
// }
//
// // ValidateUpdate implements webhookutil.validator so a webhook will be registered for the type
// func (c *MongoDB) ValidateUpdate(old runtime.Object) error {
// 	if _, err := resource.ParseQuantity(*c.Spec.Storage); err != nil {
// 		return fmt.Errorf(".spec.stroage by a parseable quantity")
// 	}
//
// 	return nil
// }
//
// var _ webhook.Defaulter = &MongoDB{}
//
// // Default implements webhookutil.defaulter so a webhook will be registered for the type
// func (c *MongoDB) Default() {
// 	if c.Spec.Storage == nil || *c.Spec.Storage == "" {
// 		s := "100Gi"
// 		c.Spec.Storage = &s
// 	}
//
// 	if c.Spec.Replicas == nil {
// 		r := int32(1)
// 		c.Spec.Replicas = &r
// 	}
// }
