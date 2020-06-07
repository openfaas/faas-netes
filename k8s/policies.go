package k8s

import (
	"reflect"
	"strings"

	"gopkg.in/yaml.v2"

	typedCorev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const PolicyAnnotationKey = "com.openfaas.policies"

// PolicyClient defines the interface for CRUD operations on policies
// and applying faaas-netes policies to function Deployments.
type PolicyClient interface {
	Get(namespace string, names ...string) ([]Policy, error)
}

// Policy defined kubernetest specific api extensions that can be predefined and applied
// to functions by annotating them with `com.openfaas/policy: name1,name2`
type Policy struct {
	// If specified, the function's pod tolerations.
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
}

// Apply adds or mutates the configuration of the Deployment with the values defined
// in the Policy. Policies are not merged, if two policies are applied, the last Policy will
// override preceding Policies with overlapping configurations.
func (p Policy) Apply(deployment *appsv1.Deployment) *appsv1.Deployment {
	if len(p.Tolerations) > 0 {
		deployment.Spec.Template.Spec.Tolerations = p.Tolerations
	}
	return deployment
}

// Remove is the inverse of Apply, removing the mutations that the Policy would have applied
func (p Policy) Remove(deployment *appsv1.Deployment) *appsv1.Deployment {
	if reflect.DeepEqual(deployment.Spec.Template.Spec.Tolerations, p.Tolerations) {
		deployment.Spec.Template.Spec.Tolerations = nil
	}
	return deployment
}

type policyClient struct {
	kube typedCorev1.ConfigMapsGetter
}

func NewConfigMapPolicyClient(kube kubernetes.Interface) PolicyClient {
	return &policyClient{kube: kube.CoreV1()}
}

// Get returns the named policies, if found, from the namespace
func (c policyClient) Get(namespace string, names ...string) ([]Policy, error) {
	var resp []Policy
	for _, name := range names {
		cm, err := c.kube.ConfigMaps(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		policy := Policy{}
		payload := []byte(cm.Data["policy"])
		err = yaml.Unmarshal(payload, &policy)
		if err != nil {
			return nil, err
		}
		resp = append(resp, policy)
	}
	return resp, nil
}

// ParsePolicyNames parsed the Policy annotation and returns the policy names it contains
func ParsePolicyNames(annotations map[string]string) (values []string) {
	if len(annotations) == 0 {
		return values
	}

	v := annotations[PolicyAnnotationKey]
	if v == "" {
		return values
	}
	values = strings.Split(v, ",")
	for idx, v := range values {
		values[idx] = strings.TrimSpace(v)
	}
	return values
}

// PoliciesToRemove parse the requested and existing annotations to determine which
// policies should be removed
func PoliciesToRemove(requested, existing map[string]string) []string {

	requestedPolicies := map[string]struct{}{}
	for _, value := range ParsePolicyNames(requested) {
		requestedPolicies[value] = struct{}{}
	}

	if len(requestedPolicies) == 0 {
		return ParsePolicyNames(existing)
	}

	var toRemove []string
	for _, name := range ParsePolicyNames(existing) {
		_, ok := requestedPolicies[name]
		if !ok {
			toRemove = append(toRemove, name)
		}
	}

	return toRemove
}
