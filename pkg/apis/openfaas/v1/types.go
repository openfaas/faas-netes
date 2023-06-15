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
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ProfileList is a list of Profiles
type ProfileList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Profile `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PolicyList is a list of Policy resources
type PolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Policy `json:"items"`
}

type ConditionMap map[string]map[string][]string

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Policy is used to define a policy for a function
// +kubebuilder:printcolumn:name="Statement",type=string,JSONPath=`.spec.statement`
type Policy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec PolicySpec `json:"spec"`
}

// PolicySpec is the spec for a Policy resource

type PolicySpec struct {
	Statement []PolicyStatement `json:"statement"`
}

type PolicyStatement struct {
	// SID is the unique identifier for the policy
	SID string `json:"sid"`

	// Effect is the effect of the policy - only Allow is supported
	Effect string `json:"effect"`

	// Action is a set of actions that the policy applies to i.e. Function:Read
	Action []string `json:"action"`

	// Resource is a set of resources that the policy applies to - only namespaces are supported at
	// present
	Resource []string `json:"resource"`

	// +optional
	// Condition is a set of conditions that the policy applies to
	// {
	// 	"StringLike": {
	// 		"jwt:https://my-identity-provider.com#sub-id": [
	// 			"1234567890",
	// 			"0987654321"
	// 		],
	// 	}
	// }
	Condition *ConditionMap `json:"condition,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RoleList is a list of Role resources
type RoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Role `json:"items"`
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Role is used to define a role for a function
// +kubebuilder:printcolumn:name="Principal",type=string,JSONPath=`.spec.principal`
// +kubebuilder:printcolumn:name="Condition",type=string,JSONPath=`.spec.condition`
// +kubebuilder:printcolumn:name="Policy",type=string,JSONPath=`.spec.policy`
type Role struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec RoleSpec `json:"spec"`
}

// RoleSpec maps a number of principals or attributes within a JWT to
// a set of policies.
type RoleSpec struct {
	// +optional
	// Policy is a list of named policies which apply to this role
	Policy []string `json:"policy"`

	// +optional
	// Principal is the principal that the role applies to i.e.
	// {
	// 		"jwt:sub":["repo:alexellis/minty:ref:refs/heads/master"]
	// }
	Principal map[string][]string `json:"principal"`

	// +optional
	// Condition is a set of conditions that can be used instead of a principal
	// to match against claims within a JWT
	// {
	// 	"StringLike": {
	// 		"jwt:https://my-identity-provider.com#sub-id": [
	// 			"1234567890",
	// 			"0987654321"
	// 		],
	// 	}
	// }
	Condition *ConditionMap `json:"condition,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// JwtIssuerList
type JwtIssuerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []JwtIssuer `json:"items"`
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// JwtIssuer is used to define a JWT issuer for a function
// +kubebuilder:printcolumn:name="Issuer",type=string,JSONPath=`.spec.iss`
// +kubebuilder:printcolumn:name="Audience",type=string,JSONPath=`.spec.aud`
// +kubebuilder:printcolumn:name="Expiry",type=string,JSONPath=`.spec.tokenExpiry`
type JwtIssuer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec JwtIssuerSpec `json:"spec"`
}

// JwtIssuerSpec is the spec for a JwtIssuer resource
type JwtIssuerSpec struct {
	// Issuer is the issuer of the JWT
	Issuer string `json:"iss"`

	// +optional
	// IssuerInternal provides an alternative URL to use to download the public key
	// for this issuer. It's useful for the system issuer.
	IssuerInternal string `json:"issInternal,omitempty"`

	// Audience is the intended audience of the JWT, at times, like with Auth0 this is the
	// client ID of the app, and not our validating server
	Audience []string `json:"aud"`

	// +optional
	TokenExpiry string `json:"tokenExpiry,omitempty"`
}
