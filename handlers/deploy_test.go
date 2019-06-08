package handlers

import (
	"github.com/openfaas/faas-netes/k8s"
	"k8s.io/client-go/kubernetes/fake"
	"testing"

	"github.com/openfaas/faas/gateway/requests"
	apiv1 "k8s.io/api/core/v1"
)

func Test_buildAnnotations_Empty_In_CreateRequest(t *testing.T) {
	request := requests.CreateFunctionRequest{}

	annotations := buildAnnotations(request)

	if len(annotations) != 1 {
		t.Errorf("want: %d annotations got: %d", 1, len(annotations))
	}

	v, ok := annotations["prometheus.io.scrape"]
	if !ok {
		t.Errorf("missing prometheus.io.scrape key")
	}

	if v != "false" {
		t.Errorf("want: %s for annotation prometheus.io.scrape got: %s", "false", v)
	}
}

func Test_buildAnnotations_From_CreateRequest(t *testing.T) {
	request := requests.CreateFunctionRequest{
		Annotations: &map[string]string{
			"date-created": "Wed 25 Jul 21:26:22 BST 2018",
			"foo":          "bar",
		},
	}

	annotations := buildAnnotations(request)

	if len(annotations) != 3 {
		t.Errorf("want: %d annotations got: %d", 1, len(annotations))
	}

	v, ok := annotations["date-created"]
	if !ok {
		t.Errorf("missing date-created key")
	}

	if v != "Wed 25 Jul 21:26:22 BST 2018" {
		t.Errorf("want: %s for annotation date-created got: %s", "Wed 25 Jul 21:26:22 BST 2018", v)
	}
}

func Test_SetNonRootUser(t *testing.T) {

	scenarios := []struct {
		name       string
		setNonRoot bool
	}{
		{"does not set userid value when SetNonRootUser is false", false},
		{"does set userid to constant value when SetNonRootUser is true", true},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			request := requests.CreateFunctionRequest{Service: "testfunc", Image: "alpine:latest"}
			factory := k8s.NewFunctionFactory(fake.NewSimpleClientset(), k8s.DeploymentConfig{
				LivenessProbe:  &k8s.ProbeConfig{},
				ReadinessProbe: &k8s.ProbeConfig{},
				SetNonRootUser: s.setNonRoot,
			})
			deployment, err := makeDeploymentSpec(request, map[string]*apiv1.Secret{}, factory)
			if err != nil {
				t.Errorf("unexpected makeDeploymentSpec error: %s", err.Error())
			}

			functionContainer := deployment.Spec.Template.Spec.Containers[0]
			if functionContainer.SecurityContext == nil {
				t.Errorf("expected container %s to have a non-nil security context", functionContainer.Name)
			}

			if !s.setNonRoot && functionContainer.SecurityContext.RunAsUser != nil {
				t.Errorf("expected RunAsUser to be nil, got %d", functionContainer.SecurityContext.RunAsUser)
			}

			if s.setNonRoot && *functionContainer.SecurityContext.RunAsUser != k8s.SecurityContextUserID {
				t.Errorf("expected RunAsUser to be %d, got %d", k8s.SecurityContextUserID, functionContainer.SecurityContext.RunAsUser)
			}
		})
	}

}
