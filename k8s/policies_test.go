package k8s

import (
	"reflect"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
)

const testPolicy = `
tolerations:
- key: "key1"
  operator: "Equal"
  value: "value1"
  effect: "NoExecute"
`

// invalidPolicyYAML has a missing colon
const invalidPolicyYAML = `
tolerations:
- key: "key1"
  operator: "Equal"
  value: "value1"
  effect "NoExecute"
`

func Test_ParsePolicyNames(t *testing.T) {
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
			name: "parses policy csv string",
			annotations: map[string]string{
				PolicyAnnotationKey: "name1,name2",
			},
			expected: []string{"name1", "name2"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ParsePolicyNames(tc.annotations)
			if !reflect.DeepEqual(tc.expected, got) {
				t.Fatalf("expected %#v, got %#v", tc.expected, got)
			}
		})
	}
}

func Test_PoliciesToRemove(t *testing.T) {
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
				PolicyAnnotationKey: "name1,name2",
			},
		},
		{
			name: "missing policy annotation on request and existing, returns nil list",
			requested: map[string]string{
				"something.else": "name1,name2",
			},
			existing: map[string]string{
				"one.more": "here",
			},
		},
		{
			name: "matching policy annotation on request and existing, returns nil list",
			requested: map[string]string{
				PolicyAnnotationKey: "name1,name2",
			},
			existing: map[string]string{
				PolicyAnnotationKey: "name1,name2",
			},
		},
		{
			name: "overlapping annotations on request and existing, returns extras from existing",
			requested: map[string]string{
				PolicyAnnotationKey: "name2",
			},
			existing: map[string]string{
				PolicyAnnotationKey: "name1,name2",
			},
			expected: []string{"name1"},
		},
		{
			name: "empty annotation on request and non-empty existing, returns all existing names",
			existing: map[string]string{
				PolicyAnnotationKey: "name1,name2",
			},
			expected: []string{"name1", "name2"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := PoliciesToRemove(tc.requested, tc.existing)
			if !reflect.DeepEqual(tc.expected, got) {
				t.Fatalf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}

func Test_TolerationsPolicy_Apply(t *testing.T) {
	expectedTolerations := []corev1.Toleration{
		{
			Key:      "foo",
			Value:    "fooValue",
			Operator: apiv1.TolerationOpEqual,
		},
	}
	p := Policy{Tolerations: expectedTolerations}

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

	basicDeployment = p.Apply(basicDeployment)
	result := basicDeployment.Spec.Template.Spec.Tolerations
	if !reflect.DeepEqual(expectedTolerations, result) {
		t.Fatalf("expected %v, got %v", expectedTolerations, result)
	}
}

func Test_RunTimeClassPolicy_Apply(t *testing.T) {
	expectedClass := "fastRunTime"
	p := Policy{RuntimeClassName: &expectedClass}

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

	basicDeployment = p.Apply(basicDeployment)
	result := basicDeployment.Spec.Template.Spec.RuntimeClassName
	if result == nil {
		t.Fatalf("expected %s, got nil", expectedClass)
	}
	if expectedClass != *result {
		t.Fatalf("expected %s, got %v", expectedClass, *result)
	}
}

func Test_RunTimeClaasPolicy_Remove(t *testing.T) {
	t.Run("remove matching runtime class ", func(t *testing.T) {
		expectedClass := "fastRunTime"
		p := Policy{RuntimeClassName: &expectedClass}

		basicDeployment := &appsv1.Deployment{
			Spec: appsv1.DeploymentSpec{
				Template: apiv1.PodTemplateSpec{
					Spec: apiv1.PodSpec{
						RuntimeClassName: &expectedClass,
						Containers: []apiv1.Container{
							{Name: "testfunc", Image: "alpine:latest"},
						},
					},
				},
			},
		}

		basicDeployment = p.Remove(basicDeployment)
		result := basicDeployment.Spec.Template.Spec.RuntimeClassName
		if result != nil {
			t.Fatalf("expected nil, got %s", *result)
		}
	})

	t.Run("leaves runtime class that does not match", func(t *testing.T) {
		expectedClass := "fastRunTime"
		policyClass := "slowRunTime"
		p := Policy{RuntimeClassName: &policyClass}

		basicDeployment := &appsv1.Deployment{
			Spec: appsv1.DeploymentSpec{
				Template: apiv1.PodTemplateSpec{
					Spec: apiv1.PodSpec{
						RuntimeClassName: &expectedClass,
						Containers: []apiv1.Container{
							{Name: "testfunc", Image: "alpine:latest"},
						},
					},
				},
			},
		}

		basicDeployment = p.Remove(basicDeployment)
		result := basicDeployment.Spec.Template.Spec.RuntimeClassName
		if !equalStrings(result, &expectedClass) {
			t.Fatalf("expected %s, got %v", expectedClass, result)
		}
	})
}

func Test_TolerationsPolicy_Remove(t *testing.T) {
	tolerations := []corev1.Toleration{
		{
			Key:      "foo",
			Value:    "fooValue",
			Operator: apiv1.TolerationOpEqual,
		},
	}
	nonPolicyToleration := corev1.Toleration{
		Key:      "second-key",
		Value:    "anotherValue",
		Operator: apiv1.TolerationOpEqual,
	}

	p := Policy{Tolerations: tolerations}

	basicDeployment := &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{Name: "testfunc", Image: "alpine:latest"},
					},
					Tolerations: append(tolerations, nonPolicyToleration),
				},
			},
		},
	}

	basicDeployment = p.Remove(basicDeployment)

	got := basicDeployment.Spec.Template.Spec.Tolerations
	expected := []corev1.Toleration{nonPolicyToleration}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("expected %v, got %v", expected, got)
	}
}

func Test_ConfigMapPolicyParsing(t *testing.T) {
	allowSpotConfig := corev1.ConfigMap{}
	allowSpotConfig.Name = "allowSpot"
	allowSpotConfig.Namespace = "functions"
	allowSpotConfig.Data = map[string]string{"policy": testPolicy}

	allowSpot := Policy{
		Tolerations: []apiv1.Toleration{
			{
				Key:               "key1",
				Value:             "value1",
				Operator:          apiv1.TolerationOpEqual,
				Effect:            apiv1.TaintEffectNoExecute,
				TolerationSeconds: nil,
			},
		},
	}

	invalidConfig := corev1.ConfigMap{}
	invalidConfig.Name = "allowSpot"
	invalidConfig.Namespace = "functions"
	invalidConfig.Data = map[string]string{"policy": invalidPolicyYAML}

	cases := []struct {
		name       string
		namespace  string
		policyName string
		configmap  corev1.ConfigMap
		expected   []Policy
		err        string
	}{
		{
			name:       "unknown policy returns error",
			namespace:  "functions",
			policyName: "unknown",
			err:        `configmaps "unknown" not found`,
		},
		{
			name:       "yaml policy parsed correctly",
			namespace:  "functions",
			policyName: "allowSpot",
			configmap:  allowSpotConfig,
			expected:   []Policy{allowSpot},
		},
		{
			name:       "yaml parsing errors are returned",
			namespace:  "functions",
			policyName: "allowSpot",
			configmap:  invalidConfig,
			err:        `yaml: line 7: could not find expected ':'`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			kube := fake.NewSimpleClientset(&tc.configmap)
			client := NewConfigMapPolicyClient(kube)
			got, err := client.Get(tc.namespace, tc.policyName)
			if tc.err != "" {
				if err == nil {
					t.Fatalf("expected error %s, got nil", tc.err)
				}

				if tc.err != err.Error() {
					t.Fatalf("expected error %s, got %s", tc.err, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error %s", err.Error())
			}

			if !reflect.DeepEqual(tc.expected, got) {
				t.Fatalf("expected %#v, got %#v", tc.expected, got)
			}

		})
	}
}
