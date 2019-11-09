// Copyright 2019 OpenFaaS Author(s)
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	types "github.com/openfaas/faas-provider/types"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testclient "k8s.io/client-go/kubernetes/fake"
)

const secretLabel = "app.kubernetes.io/managed-by"
const secretLabelValue = "openfaas"

func Test_SecretsHandler(t *testing.T) {
	namespace := "of-fnc"
	kube := testclient.NewSimpleClientset()
	secretsHandler := MakeSecretHandler(namespace, kube).ServeHTTP
	secretName := "testsecret"

	t.Run("create managed secrets", func(t *testing.T) {
		secretValue := "testsecretvalue"
		payload := fmt.Sprintf(`{"name": "%s", "value": "%s"}`, secretName, secretValue)
		req := httptest.NewRequest("POST", "http://example.com/foo", strings.NewReader(payload))
		w := httptest.NewRecorder()

		secretsHandler(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusAccepted {
			t.Errorf("want status code '%d', got '%d'", http.StatusAccepted, resp.StatusCode)
		}

		actualSecret, err := kube.CoreV1().Secrets(namespace).Get("testsecret", metav1.GetOptions{})
		if err != nil {
			t.Errorf("error validting secret: %s", err)
		}

		if actualSecret.Name != "testsecret" {
			t.Errorf("want secret with name: 'testsecret', got: '%s'", actualSecret.Name)
		}

		managedBy := actualSecret.Labels[secretLabel]
		if managedBy != secretLabelValue {
			t.Errorf("want secret to be managed by '%s', got: '%s'", secretLabelValue, managedBy)
		}

		actualValue := actualSecret.StringData[secretName]
		if actualValue != secretValue {
			t.Errorf("want secret value: '%s', got: '%s'", secretValue, actualValue)
		}

		// Attempt to re-create the secret and make sure that "409 CONFLICT" is returned.
		newReq := httptest.NewRequest("POST", "http://example.com/foo", strings.NewReader(payload))
		newW := httptest.NewRecorder()
		secretsHandler(newW, newReq)
		newResp := newW.Result()
		if newResp.StatusCode != http.StatusConflict {
			t.Errorf("want status code '%d', got '%d'", http.StatusConflict, newResp.StatusCode)
		}
	})

	t.Run("update managed secrets", func(t *testing.T) {
		newSecretValue := "newtestsecretvalue"
		payload := fmt.Sprintf(`{"name": "%s", "value": "%s"}`, secretName, newSecretValue)
		req := httptest.NewRequest("PUT", "http://example.com/foo", strings.NewReader(payload))
		w := httptest.NewRecorder()

		secretsHandler(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusAccepted {
			t.Errorf("want status code '%d', got '%d'", http.StatusAccepted, resp.StatusCode)
		}

		actualSecret, err := kube.CoreV1().Secrets(namespace).Get("testsecret", metav1.GetOptions{})
		if err != nil {
			t.Errorf("error validting secret: %s", err)
		}

		if actualSecret.Name != "testsecret" {
			t.Errorf("want secret with name: 'testsecret', got: '%s'", actualSecret.Name)
		}

		managedBy := actualSecret.Labels[secretLabel]
		if managedBy != secretLabelValue {
			t.Errorf("want secret to be managed by '%s', got: '%s'", secretLabelValue, managedBy)
		}

		actualValue := actualSecret.StringData[secretName]
		if actualValue != newSecretValue {
			t.Errorf("want secret value: '%s', got: '%s'", newSecretValue, actualValue)
		}
	})

	t.Run("update non-existing secrets returns 404", func(t *testing.T) {
		newSecretValue := "newtestsecretvalue"
		secretName := "testsecret-doesnotexist"
		payload := fmt.Sprintf(`{"name": "%s", "value": "%s"}`, secretName, newSecretValue)
		req := httptest.NewRequest("PUT", "http://example.com/foo", strings.NewReader(payload))
		w := httptest.NewRecorder()

		secretsHandler(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status code '%d', got '%d'", http.StatusNotFound, resp.StatusCode)
		}
	})

	t.Run("list managed secrets only", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com/foo", nil)
		w := httptest.NewRecorder()

		secretsHandler(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("want status code '%d', got '%d'", http.StatusOK, resp.StatusCode)
		}

		decoder := json.NewDecoder(resp.Body)

		secretList := []types.Secret{}
		err := decoder.Decode(&secretList)
		if err != nil {
			t.Error(err)
		}

		secretWant := 1
		if len(secretList) != secretWant {
			t.Errorf("want %d secret, got %d", secretWant, len(secretList))
		}

		actualSecret := secretList[0]
		if actualSecret.Name != secretName {
			t.Errorf("want secret name: '%s', got: '%s'", secretName, actualSecret.Name)
		}
	})

	t.Run("delete managed secrets", func(t *testing.T) {
		secretName := "testsecret"
		payload := fmt.Sprintf(`{"name": "%s"}`, secretName)
		req := httptest.NewRequest("DELETE", "http://example.com/foo", strings.NewReader(payload))
		w := httptest.NewRecorder()

		secretsHandler(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusAccepted {
			t.Errorf("want status code '%d', got '%d'", http.StatusAccepted, resp.StatusCode)
		}

		actualSecret, err := kube.CoreV1().Secrets(namespace).Get("testsecret", metav1.GetOptions{})
		if err == nil {
			t.Errorf("want not found error, got secret payload '%s'", actualSecret)
		}
	})

	t.Run("delete missing secret returns 404", func(t *testing.T) {
		secretName := "testsecret-doesnotexist"
		payload := fmt.Sprintf(`{"name": "%s"}`, secretName)
		req := httptest.NewRequest("DELETE", "http://example.com/foo", strings.NewReader(payload))
		w := httptest.NewRecorder()

		secretsHandler(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status code '%d', got '%d'", http.StatusNotFound, resp.StatusCode)
		}
	})
}

func Test_SecretsHandler_ListEmpty(t *testing.T) {
	namespace := "of-fnc"
	kube := testclient.NewSimpleClientset()
	secretsHandler := MakeSecretHandler(namespace, kube).ServeHTTP

	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()

	secretsHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("want status code '%d', got '%d'", http.StatusOK, resp.StatusCode)
	}
	body, _ := ioutil.ReadAll(resp.Body)

	if string(body) == "null" {
		t.Errorf(`want empty list to be valid json i.e. "[]", but was %q`, string(body))
	}
}

func Test_NamespaceResolver(t *testing.T) {
	defaultNamespace := "openfaas-fn"
	kube := testclient.NewSimpleClientset()
	getLookupNamespace := NewNamespaceResolver(defaultNamespace, kube)

	allowedAnnotation := map[string]string{"openfaas": "true"}
	allowedNamespace := "allowed-ns"
	ns := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: allowedNamespace, Annotations: allowedAnnotation}}
	kube.CoreV1().Namespaces().Create(ns)

	t.Run("no custom namespace", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com/foo", nil)
		ns, err := getLookupNamespace(req)

		if ns != defaultNamespace {
			t.Errorf("expected %s, got %s", defaultNamespace, ns)
		}

		if err != nil {
			t.Errorf("expected %v, got %v", nil, err)
		}
	})

	t.Run("allowed namespace in GET query parameters", func(t *testing.T) {
		url := fmt.Sprintf("http://example.com/foo?namespace=%s", allowedNamespace)

		req := httptest.NewRequest("GET", url, nil)
		ns, err := getLookupNamespace(req)

		if ns != allowedNamespace {
			t.Errorf("expected %s, got %s", allowedNamespace, ns)
		}

		if err != nil {
			t.Errorf("expected %v, got %v", nil, err)
		}
	})

	t.Run("forbidden namespace in GET query parameters", func(t *testing.T) {
		forbiddenNamespace := "forbidden"
		url := fmt.Sprintf("http://example.com/foo?namespace=%s", forbiddenNamespace)

		req := httptest.NewRequest("GET", url, nil)
		_, err := getLookupNamespace(req)

		errMsg := "unable to manage secrets within the forbidden namespace"
		if err.Error() != errMsg {
			t.Errorf("expected %s, got %s", errMsg, err)
		}
	})

	t.Run("allowed namespace in POST body", func(t *testing.T) {
		url := "http://example.com/foo"
		payload := fmt.Sprintf(`{"name": "secret", "value": "value", "namespace": "%s"}`, allowedNamespace)

		req := httptest.NewRequest("POST", url, strings.NewReader(payload))
		ns, err := getLookupNamespace(req)

		if ns != allowedNamespace {
			t.Errorf("expected %s, got %s", allowedNamespace, ns)
		}

		if err != nil {
			t.Errorf("expected %v, got %v", nil, err)
		}
	})

	t.Run("forbidden namespace in POST body", func(t *testing.T) {
		forbiddenNamespace := "forbidden"
		url := "http://example.com/foo"
		payload := fmt.Sprintf(`{"name": "secret", "value": "value", "namespace": "%s"}`, forbiddenNamespace)

		req := httptest.NewRequest("POST", url, strings.NewReader(payload))
		_, err := getLookupNamespace(req)

		errMsg := "unable to manage secrets within the forbidden namespace"
		if err.Error() != errMsg {
			t.Errorf("expected %s, got %s", errMsg, err)
		}
	})
}
