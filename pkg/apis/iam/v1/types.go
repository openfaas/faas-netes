package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
