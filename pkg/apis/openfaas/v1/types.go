package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:printcolumn:name="Image",type=string,JSONPath=`.spec.image`

// Function describes an OpenFaaS function
type Function struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec FunctionSpec `json:"spec"`
}

// FunctionSpec is the spec for a Function resource
type FunctionSpec struct {
	Name string `json:"name"`

	Image string `json:"image"`
	// +optional
	Handler string `json:"handler,omitempty"`
	// +optional
	Annotations *map[string]string `json:"annotations,omitempty"`
	// +optional
	Labels *map[string]string `json:"labels,omitempty"`
	// +optional
	Environment *map[string]string `json:"environment,omitempty"`
	// +optional
	Constraints []string `json:"constraints,omitempty"`
	// +optional
	Secrets []string `json:"secrets,omitempty"`
	// +optional
	Limits *FunctionResources `json:"limits,omitempty"`
	// +optional
	Requests *FunctionResources `json:"requests,omitempty"`
	// +optional
	ReadOnlyRootFilesystem bool `json:"readOnlyRootFilesystem"`
}

// FunctionResources is used to set CPU and memory limits and requests
type FunctionResources struct {
	Memory string `json:"memory,omitempty"`
	CPU    string `json:"cpu,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FunctionList is a list of Function resources
type FunctionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Function `json:"items"`
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Profile and ProfileSpec are used to customise the Pod template for
// functions
type Profile struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ProfileSpec `json:"spec"`
}

// ProfileSpec is an openfaas api extension that can be predefined and applied
// to functions by annotating them with `com.openfaas/profile: name1,name2`
type ProfileSpec struct {
	// If specified, the function's pod tolerations.
	//
	// merged into the Pod Tolerations
	//
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// RuntimeClassName refers to a RuntimeClass object in the node.k8s.io group, which should be used
	// to run this pod.  If no RuntimeClass resource matches the named class, the pod will not be run.
	// If unset or empty, the "legacy" RuntimeClass will be used, which is an implicit class with an
	// empty definition that uses the default runtime handler.
	// More info: https://git.k8s.io/enhancements/keps/sig-node/runtime-class.md
	// This is a beta feature as of Kubernetes v1.14.
	//
	// copied to the Pod RunTimeClass, this will replace any existing value or previously
	// applied Profile.
	//
	// +optional
	RuntimeClassName *string `json:"runtimeClassName,omitempty"`

	// SecurityContext holds pod-level security attributes and common container settings.
	// Optional: Defaults to empty.  See type description for default values of each field.
	//
	// each non-nil value will be merged into the function's PodSecurityContext, the value will
	// replace any existing value or previously applied Profile
	//
	// +optional
	PodSecurityContext *corev1.PodSecurityContext `json:"podSecurityContext,omitempty"`

	// If specified, the pod's scheduling constraints
	//
	// copied to the Pod Affinity, this will replace any existing value or previously
	// applied Profile. We use a replacement strategy because it is not clear that merging
	// affinities will actually produce a meaning Affinity definition, it would likely result in
	// an impossible to satisfy constraint
	//
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// TopologySpreadConstraints describes how a group of pods ought to spread across topology
	// domains. The Kubernetes will schedule pods in a way which abides by the constraints.
	//
	// https://kubernetes.io/docs/concepts/scheduling-eviction/topology-spread-constraints/
	// +optional
	TopologySpreadConstraints []corev1.TopologySpreadConstraint `json:"topologySpreadConstraints,omitempty"`

	// DNSPolicy determines how DNS resolution is handled for Pods
	//
	// copied to the Pod DNSPolicy, this will replace any existing value or previously
	// applied Profile.
	//
	// https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/#pod-s-dns-policy
	// +optional
	DNSPolicy corev1.DNSPolicy `json:"dnsPolicy,omitempty"`

	// DNSConfig allows customizing DNS resolution for Pods. See type description for default values
	// of each field.
	//
	// each non-nil value will be merged into the function's pods DNSConfig, the value will
	// replace any existing value or previously applied Profile
	//
	// https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/#pod-dns-config
	// +optional
	DNSConfig *corev1.PodDNSConfig `json:"dnsConfig,omitempty"`

	// Resources allows customizing resource requests and limits for the function container.
	//
	// Resource requests and limits keys are merged with the function container resources.
	// This will replace any existing value or previously applied Profile for that key.
	//
	// +optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ProfileList is a list of Profiles
type ProfileList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Profile `json:"items"`
}
