/*
Copyright 2019 Gavin Zhou.

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

package v1beta1

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ReceiverSpec defines the desired state of Receiver
type ReceiverSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// +optional
	Storage *string `json:"storage,omitempty"`
}

// ReceiverStatus defines the observed state of Receiver
type ReceiverStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// deploymentstatus contains the status of the deployment managed by thanos reciver
	// DeploymentStatus appsv1.DeploymentStatus `json:"deploymentStatus,omitempty"`

	// statefulSetStatus contains the status of the StatefulSet managed by Thanos
	StatefulSetStatus appsv1.StatefulSetStatus `json:"statefulSetStatus,omitempty"`

	// serviceStatus contains the status of the Service managed by thanos reciver
	ServiceStatus corev1.ServiceStatus `json:"serviceStatus,omitempty"`
}

// +kubebuilder:printcolumn:name="storage",type="string",JSONPath=".spec.storage",format="byte"
// +kubebuilder:printcolumn:name="ready replicas",type="integer",JSONPath=".status.statefulSetStatus.readyReplicas",format="int32"
// +kubebuilder:printcolumn:name="current replicas",type="integer",JSONPath=".status.statefulSetStatus.currentReplicas",format="int32"

// +kubebuilder:object:root=true
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.statefulSetStatus.replicas
// +kubebuilder:subresource:status

// Receiver is the Schema for the receivers API
type Receiver struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ReceiverSpec   `json:"spec,omitempty"`
	Status ReceiverStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ReceiverList contains a list of Receiver
type ReceiverList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Receiver `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Receiver{}, &ReceiverList{})
}
