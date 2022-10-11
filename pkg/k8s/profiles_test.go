package k8s

import (
	"reflect"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
)

const testProfile = `
podSecurityContext:
    runAsUser: 1000
    runAsGroup: 3000
    fsGroup: 2000
tolerations:
- key: "key1"
  operator: "Equal"
  value: "value1"
  effect: "NoExecute"
`

// invalidProfileYAML has a missing colon
const invalidProfileYAML = `
tolerations:
- key: "key1"
  operator: "Equal"
  value: "value1"
  effect "NoExecute"
`

func Test_ParseProfileNames(t *testing.T) {
	cases := []struct {
		name        string
		annotations map[string]string
		expected    []string
	}{
		{
			name: "empty annotations returns nil list",
		},
		{
			name: "if annotation is missing, returns nil list",
			annotations: map[string]string{
				"something.else": "foo",
			},
		},
		{
			name: "parses profile csv string",
			annotations: map[string]string{
				ProfileAnnotationKey: "name1,name2",
			},
			expected: []string{"name1", "name2"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ParseProfileNames(tc.annotations)
			if !reflect.DeepEqual(tc.expected, got) {
				t.Fatalf("expected %#v, got %#v", tc.expected, got)
			}
		})
	}
}

func Test_ProfilesToRemove(t *testing.T) {
	cases := []struct {
		name      string
		requested map[string]string
		existing  map[string]string
		expected  []string
	}{
		{
			name: "empty annotations returns nil list",
		},
		{
			name: "requested non-empty and existing is empty, returns nil list",
			requested: map[string]string{
				ProfileAnnotationKey: "name1,name2",
			},
		},
		{
			name: "missing profile annotation on request and existing, returns nil list",
			requested: map[string]string{
				"something.else": "name1,name2",
			},
			existing: map[string]string{
				"one.more": "here",
			},
		},
		{
			name: "matching profile annotation on request and existing, returns nil list",
			requested: map[string]string{
				ProfileAnnotationKey: "name1,name2",
			},
			existing: map[string]string{
				ProfileAnnotationKey: "name1,name2",
			},
		},
		{
			name: "overlapping annotations on request and existing, returns extras from existing",
			requested: map[string]string{
				ProfileAnnotationKey: "name2",
			},
			existing: map[string]string{
				ProfileAnnotationKey: "name1,name2",
			},
			expected: []string{"name1"},
		},
		{
			name: "empty annotation on request and non-empty existing, returns all existing names",
			existing: map[string]string{
				ProfileAnnotationKey: "name1,name2",
			},
			expected: []string{"name1", "name2"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ProfilesToRemove(tc.requested, tc.existing)
			if !reflect.DeepEqual(tc.expected, got) {
				t.Fatalf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}

func Test_TolerationsProfile_Apply(t *testing.T) {
	expectedTolerations := []corev1.Toleration{
		{
			Key:      "foo",
			Value:    "fooValue",
			Operator: apiv1.TolerationOpEqual,
		},
	}
	p := Profile{Tolerations: expectedTolerations}

	basicDeployment := &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{Name: "testfunc", Image: "alpine:latest"},
					},
				},
			},
		},
	}

	factory := mockFactory()
	factory.ApplyProfile(p, basicDeployment)
	result := basicDeployment.Spec.Template.Spec.Tolerations
	if !reflect.DeepEqual(expectedTolerations, result) {
		t.Fatalf("expected %v, got %v", expectedTolerations, result)
	}
}

func Test_RunAsNonRootProfile_Apply(t *testing.T) {
	expectedRoot := true
	truev := true

	p := Profile{PodSecurityContext: &corev1.PodSecurityContext{RunAsNonRoot: &truev}}

	basicDeployment := &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{Name: "testfunc", Image: "alpine:latest"},
					},
					SecurityContext: &apiv1.PodSecurityContext{
						RunAsNonRoot: &truev,
					},
				},
			},
		},
	}

	factory := mockFactory()
	factory.ApplyProfile(p, basicDeployment)
	result := basicDeployment.Spec.Template.Spec.SecurityContext.RunAsNonRoot
	if result == nil {
		t.Fatalf("expected %v, got nil", expectedRoot)
	}
	if expectedRoot != *result {
		t.Fatalf("expected %v, got %v", expectedRoot, *result)
	}
}

func Test_TolerationsProfile_Remove(t *testing.T) {
	tolerations := []corev1.Toleration{
		{
			Key:      "foo",
			Value:    "fooValue",
			Operator: apiv1.TolerationOpEqual,
		},
	}
	nonProfileToleration := corev1.Toleration{
		Key:      "second-key",
		Value:    "anotherValue",
		Operator: apiv1.TolerationOpEqual,
	}

	p := Profile{Tolerations: tolerations}

	basicDeployment := &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{Name: "testfunc", Image: "alpine:latest"},
					},
					Tolerations: append(tolerations, nonProfileToleration),
				},
			},
		},
	}

	factory := mockFactory()
	factory.RemoveProfile(p, basicDeployment)

	got := basicDeployment.Spec.Template.Spec.Tolerations
	expected := []corev1.Toleration{nonProfileToleration}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("expected %v, got %v", expected, got)
	}
}

func Test_PodSecurityProfile_Apply(t *testing.T) {
	expectedProfile := apiv1.PodSecurityContext{
		RunAsUser:  intp(1001),
		RunAsGroup: intp(2002),
	}
	p := Profile{PodSecurityContext: &expectedProfile}

	basicDeployment := &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{Name: "testfunc", Image: "alpine:latest"},
					},
				},
			},
		},
	}

	factory := mockFactory()
	factory.ApplyProfile(p, basicDeployment)
	result := basicDeployment.Spec.Template.Spec.SecurityContext
	if !reflect.DeepEqual(&expectedProfile, result) {
		t.Fatalf("expected %+v\n got %+v", &expectedProfile, result)
	}
}

func Test_PodSecurityProfile_Remove(t *testing.T) {
	p := Profile{PodSecurityContext: &apiv1.PodSecurityContext{
		RunAsUser:  intp(1001),
		RunAsGroup: intp(2002),
	}}

	runAsNonRoot := true
	expectedProfile := &apiv1.PodSecurityContext{RunAsNonRoot: &runAsNonRoot}
	basicDeployment := &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					SecurityContext: &apiv1.PodSecurityContext{
						RunAsUser:    intp(1001),
						RunAsGroup:   intp(2002),
						RunAsNonRoot: &runAsNonRoot,
					},
					Containers: []apiv1.Container{
						{Name: "testfunc", Image: "alpine:latest"},
					},
				},
			},
		},
	}

	factory := mockFactory()
	factory.RemoveProfile(p, basicDeployment)
	result := basicDeployment.Spec.Template.Spec.SecurityContext
	if !reflect.DeepEqual(expectedProfile, result) {
		t.Fatalf("expected %+v\n got %+v", &expectedProfile, result)
	}
}

func intp(v int64) *int64 {
	return &v
}
