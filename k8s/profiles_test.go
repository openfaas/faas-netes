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
runtimeClassName: "gVisor"
podSecurityContext:
    runAsUser: 1000
    runAsGroup: 3000
    fsGroup: 2000
affinity:
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms:
        - matchExpressions:
          - key: kubernetes.io/e2e-az-name
            operator: In
            values:
            - e2e-az1
            - e2e-az2
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

func Test_AffinityPolicy_Apply(t *testing.T) {
	expectedAffinity := corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					{
						MatchFields: []corev1.NodeSelectorRequirement{
							{
								Key:      "gpu",
								Operator: apiv1.NodeSelectorOpExists,
							},
						},
					},
				},
			},
		},
	}
	p := Policy{Affinity: &expectedAffinity}

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
	result := basicDeployment.Spec.Template.Spec.Affinity
	if !reflect.DeepEqual(&expectedAffinity, result) {
		t.Fatalf("expected %+v\n got %+v", &expectedAffinity, result)
	}
}

func Test_AffinityPolicy_Remove(t *testing.T) {
	t.Run("removes matching affinity definition", func(t *testing.T) {
		affinity := corev1.Affinity{
			NodeAffinity: &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{
						{
							MatchFields: []corev1.NodeSelectorRequirement{
								{
									Key:      "gpu",
									Operator: apiv1.NodeSelectorOpExists,
								},
							},
						},
					},
				},
			},
		}
		p := Policy{Affinity: &affinity}

		basicDeployment := &appsv1.Deployment{
			Spec: appsv1.DeploymentSpec{
				Template: apiv1.PodTemplateSpec{
					Spec: apiv1.PodSpec{
						Affinity: &affinity,
						Containers: []apiv1.Container{
							{Name: "testfunc", Image: "alpine:latest"},
						},
					},
				},
			},
		}

		basicDeployment = p.Remove(basicDeployment)
		result := basicDeployment.Spec.Template.Spec.Affinity
		if result != nil {
			t.Fatalf("expected nil\n got %+v", result)
		}
	})

	t.Run("does not remove non-matching affinity definition", func(t *testing.T) {
		policyAffinity := corev1.Affinity{
			NodeAffinity: &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{
						{
							MatchFields: []corev1.NodeSelectorRequirement{
								{
									Key:      "gpu",
									Operator: apiv1.NodeSelectorOpExists,
								},
							},
						},
					},
				},
			},
		}
		p := Policy{Affinity: &policyAffinity}

		expectedAffinity := corev1.Affinity{
			NodeAffinity: &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{
						{
							MatchFields: []corev1.NodeSelectorRequirement{
								{
									Key:      "bigcpu",
									Operator: apiv1.NodeSelectorOpExists,
								},
							},
						},
					},
				},
			},
		}
		basicDeployment := &appsv1.Deployment{
			Spec: appsv1.DeploymentSpec{
				Template: apiv1.PodTemplateSpec{
					Spec: apiv1.PodSpec{
						Affinity: &expectedAffinity,
						Containers: []apiv1.Container{
							{Name: "testfunc", Image: "alpine:latest"},
						},
					},
				},
			},
		}

		basicDeployment = p.Remove(basicDeployment)
		result := basicDeployment.Spec.Template.Spec.Affinity

		// the GPU affinity policy _should not_ remove the bigcpu affinity
		if !reflect.DeepEqual(&expectedAffinity, result) {
			t.Fatalf("expected %+v\n got %+v", &expectedAffinity, result)
		}
	})
}

func Test_PodSecurityPolicy_Apply(t *testing.T) {
	expectedPolicy := apiv1.PodSecurityContext{
		RunAsUser:  intp(1001),
		RunAsGroup: intp(2002),
	}
	p := Policy{PodSecurityContext: &expectedPolicy}

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
	result := basicDeployment.Spec.Template.Spec.SecurityContext
	if !reflect.DeepEqual(&expectedPolicy, result) {
		t.Fatalf("expected %+v\n got %+v", &expectedPolicy, result)
	}
}

func Test_PodSecurityPolicy_Remove(t *testing.T) {
	p := Policy{PodSecurityContext: &apiv1.PodSecurityContext{
		RunAsUser:  intp(1001),
		RunAsGroup: intp(2002),
	}}

	runAsNonRoot := true
	expectedPolicy := &apiv1.PodSecurityContext{RunAsNonRoot: &runAsNonRoot}
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

	basicDeployment = p.Remove(basicDeployment)
	result := basicDeployment.Spec.Template.Spec.SecurityContext
	if !reflect.DeepEqual(expectedPolicy, result) {
		t.Fatalf("expected %+v\n got %+v", &expectedPolicy, result)
	}
}

func Test_ConfigMapPolicyParsing(t *testing.T) {
	validConfig := corev1.ConfigMap{}
	validConfig.Name = "allowSpot"
	validConfig.Namespace = "functions"
	validConfig.Data = map[string]string{"policy": testPolicy}

	runtime := "gVisor"
	allowSpot := Policy{
		RuntimeClassName: &runtime,
		Tolerations: []apiv1.Toleration{
			{
				Key:               "key1",
				Value:             "value1",
				Operator:          apiv1.TolerationOpEqual,
				Effect:            apiv1.TaintEffectNoExecute,
				TolerationSeconds: nil,
			},
		},
		PodSecurityContext: &apiv1.PodSecurityContext{
			RunAsUser:  intp(1000),
			RunAsGroup: intp(3000),
			FSGroup:    intp(2000),
		},
		Affinity: &apiv1.Affinity{
			NodeAffinity: &apiv1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &apiv1.NodeSelector{
					NodeSelectorTerms: []apiv1.NodeSelectorTerm{
						{
							MatchExpressions: []apiv1.NodeSelectorRequirement{
								{
									Key:      "kubernetes.io/e2e-az-name",
									Operator: apiv1.NodeSelectorOpIn,
									Values:   []string{"e2e-az1", "e2e-az2"},
								},
							},
						},
					},
				},
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
			configmap:  validConfig,
			expected:   []Policy{allowSpot},
		},
		{
			name:       "yaml parsing errors are returned",
			namespace:  "functions",
			policyName: "allowSpot",
			configmap:  invalidConfig,
			err:        `error converting YAML to JSON: yaml: line 7: could not find expected ':'`,
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
				t.Fatalf("\nwant %#v\n got %#v", tc.expected, got)
			}

		})
	}
}

func intp(v int64) *int64 {
	return &v
}
