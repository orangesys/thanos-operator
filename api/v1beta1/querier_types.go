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

// QuerierSpec defines the desired state of Querier
type QuerierSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Standard object’s metadata. More info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md
	// Metadata Labels and Annotations gets propagated to the prometheus pods.
	PodMetadata *metav1.ObjectMeta `json:"podMetadata,omitempty"`
	// ServiceMonitors to be selected for target discovery.
	ServiceMonitorSelector *metav1.LabelSelector `json:"serviceMonitorSelector,omitempty"`
	// Namespaces to be selected for ServiceMonitor discovery. If nil, only
	// check own namespace.
	ServiceMonitorNamespaceSelector *metav1.LabelSelector `json:"serviceMonitorNamespaceSelector,omitempty"`

	// Define resources requests and limits for single Pods.
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Number of instances to deploy for a Prometheus deployment.
	Replicas *int32 `json:"replicas,omitempty"`

	// replicaLabel set query replica-label
	ReplicaLabel string `json:"replicaLabel,omitempty"`

	// storeDNS is storage gateway
	StoreDNS string `json:"storeDNS,omitempty"`

	// Image if specified has precedence over baseImage, tag and sha
	// combinations. Specifying the version is still necessary to ensure the
	// Prometheus Operator knows what version of Prometheus is being
	// configured.
	Image *string `json:"image,omitempty"`

	// Log level for Prometheus to be configured with.
	LogLevel string `json:"logLevel,omitempty"`
}

// QuerierStatus defines the observed state of Querier
type QuerierStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	DeploymentStatus appsv1.DeploymentStatus `json:"deploymentStatus,omitempty"`

	// serviceStatus contains the status of the Service managed by thanos reciver
	ServiceStatus corev1.ServiceStatus `json:"serviceStatus,omitempty"`

	// Total number of non-terminated pods targeted by this Prometheus deployment
	// that have the desired version spec.
	UpdatedReplicas int32 `json:"updatedReplicas"`
	// Total number of available pods (ready for at least minReadySeconds)
	// targeted by this Prometheus deployment.
	AvailableReplicas int32 `json:"availableReplicas"`
	// Total number of unavailable pods targeted by this Prometheus deployment.
	UnavailableReplicas int32 `json:"unavailableReplicas"`
}

// +kubebuilder:object:root=true

// Querier is the Schema for the queriers API
type Querier struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   QuerierSpec   `json:"spec,omitempty"`
	Status QuerierStatus `json:"status,omitempty"`
}

// // Thanos is the Schema for the thanos API
// type Thanos struct {
// 	Querier  Querier
// 	Receiver Receiver
// }

// +kubebuilder:object:root=true

// QuerierList contains a list of Querier
type QuerierList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Querier `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Querier{}, &QuerierList{})
}
