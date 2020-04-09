package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	clientset "github.com/openfaas/faas-netes/pkg/client/clientset/versioned/fake"

	types "github.com/openfaas/faas-provider/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_makeApplyHandler(t *testing.T) {
	namespace := "openfaas-fn"
	createVal := "test1"
	updateVal := "test2"
	fn := types.FunctionDeployment{
		Service:                "nodeinfo",
		Image:                  "functions/nodeinfo",
		ReadOnlyRootFilesystem: true,
		Labels: &map[string]string{
			"test": createVal,
		},
		Secrets: []string{createVal},
	}

	kube := clientset.NewSimpleClientset()
	applyHandler := makeApplyHandler(namespace, kube).ServeHTTP

	// test create fn
	fnJson, _ := json.Marshal(fn)
	req := httptest.NewRequest("POST", "http://system/functions", bytes.NewBuffer(fnJson))
	w := httptest.NewRecorder()

	applyHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("expected status code '%d', got '%d'", http.StatusAccepted, resp.StatusCode)
	}

	newFunction, err := kube.OpenfaasV1().Functions(namespace).Get(fn.Service, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("error validating function: %v", err)
	}

	if !newFunction.Spec.ReadOnlyRootFilesystem {
		t.Errorf("expected ReadOnlyRootFilesystem '%v' got: '%v'",
			true, newFunction.Spec.ReadOnlyRootFilesystem)
	}

	if newFunction.Spec.Labels == nil || len(*newFunction.Spec.Labels) < 1 {
		t.Fatal("expected one label got none")
	}

	newLabels := *newFunction.Spec.Labels
	if newLabels["test"] != createVal {
		t.Errorf("expected label '%s' got: '%s'", createVal, newLabels["test"])
	}

	if len(newFunction.Spec.Secrets) < 1 {
		t.Fatal("expected one secret got none")
	}

	if newFunction.Spec.Secrets[0] != createVal {
		t.Errorf("expected secret '%s' got: '%s'", createVal, newFunction.Spec.Secrets[0])
	}

	// test update fn
	fn.Image += ":v1.0"
	fn.ReadOnlyRootFilesystem = false
	fn.Labels = &map[string]string{
		"test": updateVal,
	}
	fn.Secrets = []string{updateVal}

	updatedJson, _ := json.Marshal(fn)
	reqUp := httptest.NewRequest("POST", "http://system/functions", bytes.NewBuffer(updatedJson))
	wUp := httptest.NewRecorder()

	applyHandler(wUp, reqUp)

	respUp := wUp.Result()
	if respUp.StatusCode != http.StatusAccepted {
		t.Errorf("expected status code '%d', got '%d'", http.StatusAccepted, resp.StatusCode)
	}

	updatedFunction, err := kube.OpenfaasV1().Functions(namespace).Get(fn.Service, metav1.GetOptions{})
	if err != nil {
		t.Errorf("error validating function: %v", err)
	}

	if updatedFunction.Spec.Image != fn.Image {
		t.Errorf("expected image '%s' got: '%s'",
			fn.Image, updatedFunction.Spec.Image)
	}

	if updatedFunction.Spec.ReadOnlyRootFilesystem {
		t.Errorf("expected ReadOnlyRootFilesystem '%v' got: '%v'",
			false, updatedFunction.Spec.ReadOnlyRootFilesystem)
	}

	updatedLabels := *updatedFunction.Spec.Labels
	if updatedLabels["test"] != updateVal {
		t.Errorf("expected label '%s' got: '%s'", updateVal, updatedLabels["test"])
	}

	if updatedFunction.Spec.Secrets[0] != updateVal {
		t.Errorf("expected secret '%s' got: '%s'", updateVal, updatedFunction.Spec.Secrets[0])
	}
}
