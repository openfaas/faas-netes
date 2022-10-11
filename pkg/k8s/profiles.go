package k8s

import (
	"context"
	"reflect"
	"strings"

	v1 "github.com/openfaas/faas-netes/pkg/apis/openfaas/v1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

const ProfileAnnotationKey = "com.openfaas.profile"

// ProfileClient defines the interface for CRUD operations on profiles
// and applying faas-netes profiles to function Deployments.
type ProfileClient interface {
	Get(ctx context.Context, namespace string, names ...string) ([]Profile, error)
}

// Profile is and openfaas api extensions that can be predefined and applied
// to functions by annotating them with `com.openfaas.profile: name1,name2`
type Profile v1.ProfileSpec

// profileCRDClient implements PolicyClient using the openfaas CRD Profile
type profileCRDClient struct {
	client NamespacedProfiler
}

func (c profileCRDClient) Get(ctx context.Context, namespace string, names ...string) ([]Profile, error) {
	var resp []Profile
	for _, name := range names {
		// this is where we would consider using an informer/lister. The throughput on this
		// API. We expect will be similar to the secrets API, since we only use it during
		// function Deploy
		// Note Lister interfaces do not have context yet
		profile, err := c.client.Profiles(namespace).Get(name)
		if err != nil {
			return nil, err
		}
		resp = append(resp, Profile(profile.Spec))
	}
	return resp, nil
}

// NewProfileClient returns the ProfilerClient powered by the Profile CRD
func (f FunctionFactory) NewProfileClient() ProfileClient {
	// this is where we can replace with an informer/listener in the future
	return &profileCRDClient{client: f.Profiler}
}

// GetProfiles retrieves in the names string, names is the raw csv value in the
// function annotation
func (f FunctionFactory) GetProfiles(ctx context.Context, namespace string, annotations map[string]string) ([]Profile, error) {
	if len(annotations) == 0 {
		return nil, nil
	}

	client := f.NewProfileClient()
	profileNames := ParseProfileNames(annotations)

	return client.Get(ctx, namespace, profileNames...)
}

func (f FunctionFactory) GetProfilesToRemove(ctx context.Context, namespace string, annotations, currentAnnotations map[string]string) ([]Profile, error) {
	if len(annotations) == 0 {
		return nil, nil
	}

	toRemove := ProfilesToRemove(annotations, currentAnnotations)

	client := f.NewProfileClient()
	return client.Get(ctx, namespace, toRemove...)
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

// ApplyProfile adds or mutates the configuration of the Deployment with the values defined
// in the Profile. Profiles are not merged, if two profiles are applied, the last Profile will
// override preceding Profiles with overlapping configurations.
func (f FunctionFactory) ApplyProfile(profile Profile, deployment *appsv1.Deployment) {
	if len(profile.Tolerations) > 0 {
		deployment.Spec.Template.Spec.Tolerations = append(deployment.Spec.Template.Spec.Tolerations, profile.Tolerations...)
	}

	if profile.PodSecurityContext != nil {
		if deployment.Spec.Template.Spec.SecurityContext == nil {
			deployment.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{}
		}

		profile.PodSecurityContext.DeepCopyInto(deployment.Spec.Template.Spec.SecurityContext)
	}
}

// RemoveProfile is the inverse of Apply, removing the mutations that the Profile would have applied
func (f FunctionFactory) RemoveProfile(profile Profile, deployment *appsv1.Deployment) {
	for _, profileToleration := range profile.Tolerations {
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

	if profile.PodSecurityContext != nil {
		sc := deployment.Spec.Template.Spec.SecurityContext

		if reflect.DeepEqual(profile.PodSecurityContext.SELinuxOptions, sc.SELinuxOptions) {
			deployment.Spec.Template.Spec.SecurityContext.SELinuxOptions = nil
		}
		if reflect.DeepEqual(profile.PodSecurityContext.SELinuxOptions, sc.WindowsOptions) {
			deployment.Spec.Template.Spec.SecurityContext.WindowsOptions = nil
		}
		if profile.PodSecurityContext.RunAsUser != nil {
			deployment.Spec.Template.Spec.SecurityContext.RunAsUser = nil
		}
		if profile.PodSecurityContext.RunAsGroup != nil {
			deployment.Spec.Template.Spec.SecurityContext.RunAsGroup = nil
		}
		if profile.PodSecurityContext.RunAsNonRoot != nil {
			deployment.Spec.Template.Spec.SecurityContext.RunAsNonRoot = nil
		}
		if profile.PodSecurityContext.SupplementalGroups != nil {
			deployment.Spec.Template.Spec.SecurityContext.SupplementalGroups = nil
		}
		if profile.PodSecurityContext.FSGroup != nil {
			deployment.Spec.Template.Spec.SecurityContext.FSGroup = nil
		}
		if profile.PodSecurityContext.Sysctls != nil {
			deployment.Spec.Template.Spec.SecurityContext.Sysctls = nil
		}
	}
}
