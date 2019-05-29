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
	// +kubebuilder:validation:Minimum=1
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// +optional
	Storage *string `json:"storage,omitempty"`
}

// MongoDBStatus defines the observed state of MongoDB
type MongoDBStatus struct {
	// statefulSetStatus contains the status of the StatefulSet managed by MongoDB
	StatefulSetStatus appsv1.StatefulSetStatus `json:"statefulSetStatus,omitempty"`

	// serviceStatus contains the status of the Service managed by MongoDB
	ServiceStatus corev1.ServiceStatus `json:"serviceStatus,omitempty"`
}

// +kubebuilder:printcolumn:name="storage",type="string",JSONPath=".spec.storage",format="byte"
// +kubebuilder:printcolumn:name="replicas",type="integer",JSONPath=".spec.replicas",format="int32"
// +kubebuilder:printcolumn:name="ready replicas",type="integer",JSONPath=".status.statefulSetStatus.readyReplicas",format="int32"
// +kubebuilder:printcolumn:name="current replicas",type="integer",JSONPath=".status.statefulSetStatus.currentReplicas",format="int32"
// +kubebuilder:object:root=true
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.statefulSetStatus.replicas
// +kubebuilder:subresource:status

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
