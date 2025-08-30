package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ConfigMapRef struct {

	// Name is the name of the configMap that contains the script to be run
	// when the clusterVersion changes.
	Name string `json:"name"`

	// +optional
	// Namespace is the name of the configMap that contains the script to be run when the clusterVersion changes.
	// If left empty, it will default to the namespace of the RoleUpdater.
	Namespace string `json:"namespace,omitempty"`
}

// Spec is a list of the desired state of the RoleUpdater. It has a single field, configMapRef, which is a reference to a configMap
// in the cluster that contains the script that will be run every single clusterVersion change.
type RoleUpdaterSpec struct {

	// ConfigMapRef is a field that contains the name and the namespace of the configMap that contains the script to be run
	// when the clusterVersion changes.
	ConfigMapRef ConfigMapRef `json:"configMapRef"`

	// ForceRun is an optional boolean field that, when set to true, forces the operator to run the script inside the configMap
	// specified in .spec.configMapRef during the next reconciliation loop, regardless of whether the clusterVersion changed or not.
	ForceRun bool `json:"forceRun,omitempty"`
}

// Status defines the observed state of RoleUpdater.
type RoleUpdaterStatus struct {

	// The last time the operator checked the clusterVersion resource.
	LastCheckTime string `json:"lastCheckTime,omitempty"`

	// The current clusterVersion fetched from the ClusterVersion resource.
	ClusterVersion string `json:"clusterVersion,omitempty"`

	// A short string value indicating the status of the RoleUpdater.
	Status string `json:"status,omitempty"`

	// An informative message indicating the status of the RoleUpdater.
	Message string `json:"message,omitempty"`

	// A boolean value that indicates whether the configMap specified in spec.configMapRef is present in the cluster.
	ConfigMapPresent bool `json:"configMapPresent"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

type RoleUpdater struct {
	metav1.TypeMeta `json:",inline"`

	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// +required
	Spec RoleUpdaterSpec `json:"spec"`

	// +optional
	Status RoleUpdaterStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

type RoleUpdaterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RoleUpdater `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RoleUpdater{}, &RoleUpdaterList{})
}
