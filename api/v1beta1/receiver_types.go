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

	// Standard object’s metadata. More info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md
	// Metadata Labels and Annotations gets propagated to the prometheus pods.
	PodMetadata *metav1.ObjectMeta `json:"podMetadata,omitempty"`
	// ServiceMonitors to be selected for target discovery.
	ServiceMonitorSelector *metav1.LabelSelector `json:"serviceMonitorSelector,omitempty"`
	// Namespaces to be selected for ServiceMonitor discovery. If nil, only
	// check own namespace.
	ServiceMonitorNamespaceSelector *metav1.LabelSelector `json:"serviceMonitorNamespaceSelector,omitempty"`

	// Number of instances to deploy for a Prometheus deployment.
	Replicas *int32 `json:"replicas,omitempty"`

	// Version of Prometheus to be deployed.
	Version string `json:"version,omitempty"`
	// Tag of Prometheus container image to be deployed. Defaults to the value of `version`.
	// Version is ignored if Tag is set.
	Tag string `json:"tag,omitempty"`

	// If specified, the pod's scheduling constraints.
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// Image if specified has precedence over baseImage, tag and sha
	// combinations. Specifying the version is still necessary to ensure the
	// Prometheus Operator knows what version of Prometheus is being
	// configured.
	Image *string `json:"image,omitempty"`
	// Base image to use for a Prometheus deployment.

	BaseImage string `json:"baseImage,omitempty"`

	// Define resources requests and limits for single Pods.
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Time duration Prometheus shall retain data for. Default is '24h',
	// and must match the regular expression `[0-9]+(ms|s|m|h|d|w|y)` (milliseconds seconds minutes hours days weeks years).
	Retention string `json:"retention,omitempty"`

	// The recieve prefix storage with tsdb
	ReceivePrefix string `json:"receivePrefix,omitempty"`

	// Log level for Prometheus to be configured with.
	LogLevel string `json:"logLevel,omitempty"`

	// the receiver labels to set with receiver config
	ReceiveLables string `json:"receiveLabels,omitempty"`

	// object storage type GCS OR S3
	ObjectStorageType string `json:"objstoreType,omitempty"`

	// secret name is gcs iam secret name
	SecretName string `json:"secretName,omitempty"`

	// object storage bucket name need set object storage type
	BucketName string `json:"bucketName,omitempty"`

	// The labels to add to any time series or alerts when communicating with
	// external systems (federation, remote storage, Alertmanager).
	ExternalLabels map[string]string `json:"externalLabels,omitempty"`

	// Storage spec to specify how storage shall be used.
	Storage string `json:"storage,omitempty"`

	// Define which Nodes the Pods are scheduled on.
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Secrets is a list of Secrets in the same namespace as the Prometheus
	// object, which shall be mounted into the Prometheus Pods.
	// The Secrets are mounted into /etc/prometheus/secrets/<secret-name>.
	Secrets []string `json:"secrets,omitempty"`

	// containers is entirely outside the scope of what the maintainers will support and by doing
	// so, you accept that this behaviour may break at any time without notice.
	Containers []corev1.Container `json:"containers,omitempty"`

	// ObjectStorageConfig configures object storage in Thanos.
	ObjectStorageConfig *corev1.SecretKeySelector `json:"objectStorageConfig,omitempty"`
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

	// Total number of non-terminated pods targeted by this Prometheus deployment
	// that have the desired version spec.
	UpdatedReplicas int32 `json:"updatedReplicas"`
	// Total number of available pods (ready for at least minReadySeconds)
	// targeted by this Prometheus deployment.
	AvailableReplicas int32 `json:"availableReplicas"`
	// Total number of unavailable pods targeted by this Prometheus deployment.
	UnavailableReplicas int32 `json:"unavailableReplicas"`
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
