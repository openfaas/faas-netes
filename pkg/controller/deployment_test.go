package controller

import (
	"testing"

	faasv1 "github.com/openfaas/faas-netes/pkg/apis/openfaas/v1"
	"github.com/openfaas/faas-netes/pkg/k8s"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_GracePeriodFromWriteTimeout(t *testing.T) {

	scenarios := []struct {
		name        string
		wantSeconds int64
		envs        map[string]string
	}{
		{"grace period is the default", 32, map[string]string{}},
		{"grace period is set from write_timeout", 62, map[string]string{"write_timeout": "60s"}},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {

			want := int64(s.wantSeconds)
			function := &faasv1.Function{
				ObjectMeta: metav1.ObjectMeta{
					Name: "alpine",
				},
				Spec: faasv1.FunctionSpec{
					Name:                   "alpine",
					Image:                  "ghcr.io/openfaas/alpine:latest",
					Annotations:            &map[string]string{},
					ReadOnlyRootFilesystem: true,
					Environment:            &s.envs},
			}

			factory := NewFunctionFactory(fake.NewSimpleClientset(),
				k8s.DeploymentConfig{
					HTTPProbe:      false,
					SetNonRootUser: true,
					LivenessProbe: &k8s.ProbeConfig{
						PeriodSeconds:       1,
						TimeoutSeconds:      3,
						InitialDelaySeconds: 0,
					},
					ReadinessProbe: &k8s.ProbeConfig{
						PeriodSeconds:       1,
						TimeoutSeconds:      3,
						InitialDelaySeconds: 0,
					},
				})

			secrets := map[string]*corev1.Secret{}

			deployment := newDeployment(function, nil, secrets, factory)
			got := deployment.Spec.Template.Spec.TerminationGracePeriodSeconds
			if got == nil {
				t.Errorf("TerminationGracePeriodSeconds not set, but want %d", want)
				t.Fail()
				return
			}

			if want != *got {
				t.Errorf("TerminationGracePeriodSeconds want %d, but got %d", want, got)
				t.Fail()
			}
		})
	}
}

func Test_newDeployment(t *testing.T) {
	function := &faasv1.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kubesec",
		},
		Spec: faasv1.FunctionSpec{
			Name:  "kubesec",
			Image: "docker.io/kubesec/kubesec",
			Annotations: &map[string]string{
				"com.openfaas.serviceaccount":           "kubesec",
				"com.openfaas.health.http.initialDelay": "2m",
				"com.openfaas.health.http.path":         "/healthz",
			},
			ReadOnlyRootFilesystem: true,
		},
	}

	factory := NewFunctionFactory(fake.NewSimpleClientset(),
		k8s.DeploymentConfig{
			HTTPProbe:      false,
			SetNonRootUser: true,
			LivenessProbe: &k8s.ProbeConfig{
				PeriodSeconds:       1,
				TimeoutSeconds:      3,
				InitialDelaySeconds: 0,
			},
			ReadinessProbe: &k8s.ProbeConfig{
				PeriodSeconds:       1,
				TimeoutSeconds:      3,
				InitialDelaySeconds: 0,
			},
		})

	secrets := map[string]*corev1.Secret{}

	deployment := newDeployment(function, nil, secrets, factory)

	if deployment.Spec.Template.Spec.ServiceAccountName != "kubesec" {
		t.Errorf("ServiceAccountName should be %s", "kubesec")
		t.Fail()
	}

	if deployment.Spec.Template.Spec.Containers[0].ReadinessProbe.HTTPGet.Path != "/healthz" {
		t.Errorf("Readiness probe should have HTTPGet handler set to %s", "/healthz")
		t.Fail()
	}

	if deployment.Spec.Template.Spec.Containers[0].LivenessProbe.InitialDelaySeconds != 120 {
		t.Errorf("Liveness probe should have initial delay seconds set to %s", "2m")
		t.Fail()
	}

	if !*(deployment.Spec.Template.Spec.Containers[0].SecurityContext.ReadOnlyRootFilesystem) {
		t.Errorf("ReadOnlyRootFilesystem should be true")
		t.Fail()
	}

	if *(deployment.Spec.Template.Spec.Containers[0].SecurityContext.RunAsUser) != k8s.SecurityContextUserID {
		t.Errorf("RunAsUser should be %v", k8s.SecurityContextUserID)
		t.Fail()
	}
}

func Test_newDeployment_PrometheusScrape_NotOverridden(t *testing.T) {
	function := &faasv1.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kubesec",
		},
		Spec: faasv1.FunctionSpec{
			Name:  "kubesec",
			Image: "docker.io/kubesec/kubesec",
			Annotations: &map[string]string{
				"prometheus.io.scrape": "true",
			},
		},
	}

	factory := NewFunctionFactory(fake.NewSimpleClientset(),
		k8s.DeploymentConfig{
			HTTPProbe:      false,
			SetNonRootUser: true,
			LivenessProbe: &k8s.ProbeConfig{
				PeriodSeconds:       1,
				TimeoutSeconds:      3,
				InitialDelaySeconds: 0,
			},
			ReadinessProbe: &k8s.ProbeConfig{
				PeriodSeconds:       1,
				TimeoutSeconds:      3,
				InitialDelaySeconds: 0,
			},
		})

	secrets := map[string]*corev1.Secret{}

	deployment := newDeployment(function, nil, secrets, factory)

	want := "true"

	if deployment.Spec.Template.Annotations["prometheus.io.scrape"] != want {
		t.Errorf("Annotation prometheus.io.scrape should be %s, was: %s", want, deployment.Spec.Template.Annotations["prometheus.io.scrape"])
	}
}
