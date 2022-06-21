package schema

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type Region struct {
	ID                string       `json:"id,omitempty" optional:"true" description:"region id, auto generated"`
	Name              string       `json:"name" validate:"required,hostname_rfc1123" description:"region name, format with hostname rfc1123"`
	CreationTimestamp metav1.Time  `json:"creationTimestamp,omitempty" optional:"true" description:"creation ts, auto generated"`
	UpdateTimestamp   *metav1.Time `json:"updateTimestamp,omitempty" optional:"true" description:"update ts, auto generated"`
	// TODO: add some extra field
}
