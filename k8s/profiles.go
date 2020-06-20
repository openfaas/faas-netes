package k8s

import (
	"reflect"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	typedCorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const ProfileAnnotationKey = "com.openfaas.profile"

// ProfileClient defines the interface for CRUD operations on profiles
// and applying faas-netes profiles to function Deployments.
type ProfileClient interface {
	Get(namespace string, names ...string) ([]Profile, error)
}

// Profile is and openfaas api extensions that can be predefined and applied
// to functions by annotating them with `com.openfaas.profile: name1,name2`
type Profile struct {
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

	// If specified, the pod's scheduling constraints
	//
	// copied to the Pod Affinity, this will replace any existing value or previously
	// applied Profile. We use a replacement strategy because it is not clear that merging
	// affinities will actually produce a meaning Affinity definition, it would likely result in
	// an impossible to satisfy constraint
	//
	// +optional
	Affinity *corev1.Affinity

	// SecurityContext holds pod-level security attributes and common container settings.
	// Optional: Defaults to empty.  See type description for default values of each field.
	//
	// each non-nil value will be merged into the function's PodSecurityContext, the value will
	// replace any existing value or previously applied Profile
	//
	// +optional
	PodSecurityContext *corev1.PodSecurityContext `json:"podSecurityContext,omitempty"`
}

// Apply adds or mutates the configuration of the Deployment with the values defined
// in the Profile. Profiles are not merged, if two profiles are applied, the last Profile will
// override preceding Profiles with overlapping configurations.
func (p Profile) Apply(deployment *appsv1.Deployment) *appsv1.Deployment {
	if len(p.Tolerations) > 0 {
		deployment.Spec.Template.Spec.Tolerations = append(deployment.Spec.Template.Spec.Tolerations, p.Tolerations...)
	}

	if p.RuntimeClassName != nil {
		deployment.Spec.Template.Spec.RuntimeClassName = p.RuntimeClassName
	}

	if p.Affinity != nil {
		// use a replacement strategy because it is not clear that merging affinities will
		// actually produce a meaning Affinity definition, it would likely result in
		// an impossible to satisfy constraint
		deployment.Spec.Template.Spec.Affinity = p.Affinity
	}

	if p.PodSecurityContext != nil {
		if deployment.Spec.Template.Spec.SecurityContext == nil {
			deployment.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{}
		}

		p.PodSecurityContext.DeepCopyInto(deployment.Spec.Template.Spec.SecurityContext)
	}

	return deployment
}

// Remove is the inverse of Apply, removing the mutations that the Profile would have applied
func (p Profile) Remove(deployment *appsv1.Deployment) *appsv1.Deployment {

	for _, profileToleration := range p.Tolerations {
		// filter the existing tolerations and then update the deployment
		// filter without allocation implementation from
		// https://github.com/golang/go/wiki/SliceTricks#filtering-without-allocating
		newTolerations := deployment.Spec.Template.Spec.Tolerations[:0]
		for _, toleration := range deployment.Spec.Template.Spec.Tolerations {
			if !reflect.DeepEqual(profileToleration, toleration) {
				newTolerations = append(newTolerations, toleration)
			}
		}
		deployment.Spec.Template.Spec.Tolerations = newTolerations
	}

	if p.RuntimeClassName != nil {
		if equalStrings(deployment.Spec.Template.Spec.RuntimeClassName, p.RuntimeClassName) {
			deployment.Spec.Template.Spec.RuntimeClassName = nil
		}
	}

	if p.Affinity != nil && reflect.DeepEqual(p.Affinity, deployment.Spec.Template.Spec.Affinity) {
		deployment.Spec.Template.Spec.Affinity = nil
	}

	if p.PodSecurityContext != nil {
		sc := deployment.Spec.Template.Spec.SecurityContext

		if reflect.DeepEqual(p.PodSecurityContext.SELinuxOptions, sc.SELinuxOptions) {
			deployment.Spec.Template.Spec.SecurityContext.SELinuxOptions = nil
		}
		if reflect.DeepEqual(p.PodSecurityContext.SELinuxOptions, sc.WindowsOptions) {
			deployment.Spec.Template.Spec.SecurityContext.WindowsOptions = nil
		}
		if p.PodSecurityContext.RunAsUser != nil {
			deployment.Spec.Template.Spec.SecurityContext.RunAsUser = nil
		}
		if p.PodSecurityContext.RunAsGroup != nil {
			deployment.Spec.Template.Spec.SecurityContext.RunAsGroup = nil
		}
		if p.PodSecurityContext.RunAsNonRoot != nil {
			deployment.Spec.Template.Spec.SecurityContext.RunAsNonRoot = nil
		}
		if p.PodSecurityContext.SupplementalGroups != nil {
			deployment.Spec.Template.Spec.SecurityContext.SupplementalGroups = nil
		}
		if p.PodSecurityContext.FSGroup != nil {
			deployment.Spec.Template.Spec.SecurityContext.FSGroup = nil
		}
		if p.PodSecurityContext.Sysctls != nil {
			deployment.Spec.Template.Spec.SecurityContext.Sysctls = nil
		}
	}

	return deployment
}

func equalStrings(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}

	if a != nil && b == nil {
		return false
	}

	if a == nil && b != nil {
		return false
	}

	// now we know both values are non-nil
	return *a == *b
}

type profileClient struct {
	kube typedCorev1.ConfigMapsGetter
}

func NewConfigMapProfileClient(kube kubernetes.Interface) ProfileClient {
	return &profileClient{kube: kube.CoreV1()}
}

// Get returns the named profiles, if found, from the namespace
func (c profileClient) Get(namespace string, names ...string) ([]Profile, error) {
	var resp []Profile
	for _, name := range names {
		cm, err := c.kube.ConfigMaps(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		profile := Profile{}

		data := strings.NewReader(cm.Data["profile"])
		err = yaml.NewYAMLOrJSONDecoder(data, 100).Decode(&profile)
		if err != nil {
			return nil, err
		}
		resp = append(resp, profile)
	}
	return resp, nil
}

// ParseProfileNames parsed the Profile annotation and returns the profile names it contains
func ParseProfileNames(annotations map[string]string) (values []string) {
	if len(annotations) == 0 {
		return values
	}

	v := annotations[ProfileAnnotationKey]
	if v == "" {
		return values
	}
	values = strings.Split(v, ",")
	for idx, v := range values {
		values[idx] = strings.TrimSpace(v)
	}
	return values
}

// ProfilesToRemove parse the requested and existing annotations to determine which
// profiles should be removed
func ProfilesToRemove(requested, existing map[string]string) []string {

	requestedProfiles := map[string]struct{}{}
	for _, value := range ParseProfileNames(requested) {
		requestedProfiles[value] = struct{}{}
	}

	if len(requestedProfiles) == 0 {
		return ParseProfileNames(existing)
	}

	var toRemove []string
	for _, name := range ParseProfileNames(existing) {
		_, ok := requestedProfiles[name]
		if !ok {
			toRemove = append(toRemove, name)
		}
	}

	return toRemove
}
