package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ConfigMapRef struct {
	Name string `json:"name"`
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

type RoleUpdaterSpec struct {
	ConfigMapRef ConfigMapRef `json:"configMapRef"`
}

type RoleUpdaterStatus struct {
	LastCheckTime    string `json:"lastCheckTime,omitempty"`
	ClusterVersion   string `json:"clusterVersion,omitempty"`
	Status           string `json:"status,omitempty"`
	Message          string `json:"message,omitempty"`
	ConfigMapPresent bool   `json:"configMapPresent"`
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
