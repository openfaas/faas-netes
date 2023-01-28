// Copyright 2020 OpenFaaS Author(s)
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	types "github.com/openfaas/faas-provider/types"
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

		actualSecret, err := kube.CoreV1().Secrets(namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
		if err != nil {
			t.Errorf("error validting secret: %s", err)
		}

		if actualSecret.Name != secretName {
			t.Errorf("want secret with name: %q, got: %q", secretName, actualSecret.Name)
		}

		managedBy := actualSecret.Labels[secretLabel]
		if managedBy != secretLabelValue {
			t.Errorf("want secret to be managed by '%s', got: '%s'", secretLabelValue, managedBy)
		}

		actualValue := actualSecret.Data[secretName]
		if bytes.Equal(actualValue, []byte(secretValue)) == false {
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

	t.Run("create raw managed secrets", func(t *testing.T) {
		secretValue := []byte("testsecretvalue")
		secretName := "raw-secret"

		payload := fmt.Sprintf(`{"name": "%s", "value": "%s"}`, secretName, secretValue)
		req := httptest.NewRequest("POST", "http://example.com/foo", strings.NewReader(payload))
		w := httptest.NewRecorder()

		secretsHandler(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusAccepted {
			t.Errorf("want status code '%d', got '%d'", http.StatusAccepted, resp.StatusCode)
		}

		actualSecret, err := kube.CoreV1().Secrets(namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
		if err != nil {
			t.Errorf("error validting secret: %s", err)
		}

		if actualSecret.Name != secretName {
			t.Errorf("want secret with name: 'testsecret', got: '%s'", actualSecret.Name)
		}

		managedBy := actualSecret.Labels[secretLabel]
		if managedBy != secretLabelValue {
			t.Errorf("want secret to be managed by '%s', got: '%s'", secretLabelValue, managedBy)
		}

		actualValue := actualSecret.Data[secretName]
		if bytes.Equal(actualValue, secretValue) == false {
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

		actualSecret, err := kube.CoreV1().Secrets(namespace).Get(context.TODO(), "testsecret", metav1.GetOptions{})
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

		actualValue := actualSecret.Data[secretName]
		if bytes.Equal(actualValue, []byte(newSecretValue)) == false {
			t.Errorf("want secret value: '%s', got: '%s'", newSecretValue, actualValue)
		}
	})

	t.Run("update raw managed secrets", func(t *testing.T) {
		newSecretValue := []byte("newtestsecretvalue")
		secretName := "raw-secret"
		payload := fmt.Sprintf(`{"name": "%s", "value": "%s"}`, secretName, newSecretValue)
		req := httptest.NewRequest("PUT", "http://example.com/foo", strings.NewReader(payload))
		w := httptest.NewRecorder()

		secretsHandler(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusAccepted {
			t.Errorf("want status code '%d', got '%d'", http.StatusAccepted, resp.StatusCode)
		}

		actualSecret, err := kube.CoreV1().Secrets(namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
		if err != nil {
			t.Errorf("error validting secret: %s", err)
		}

		if actualSecret.Name != secretName {
			t.Errorf("want secret with name: 'testsecret', got: '%s'", actualSecret.Name)
		}

		managedBy := actualSecret.Labels[secretLabel]
		if managedBy != secretLabelValue {
			t.Errorf("want secret to be managed by '%s', got: '%s'", secretLabelValue, managedBy)
		}

		actualValue := actualSecret.Data[secretName]
		if bytes.Equal(actualValue, newSecretValue) == false {
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
			t.Errorf("want status code '%d', got '%d'", http.StatusNotFound, resp.StatusCode)
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

		secretWant := 2
		if len(secretList) != secretWant {
			t.Errorf("want %d secret, got %d", secretWant, len(secretList))
		}

		want := []string{"testsecret", "raw-secret"}
		got := []string{}
		for _, s := range secretList {
			for _, w := range want {
				if w == s.Name {
					got = append(got, w)
				}
			}
		}

		if len(want) != len(got) {
			t.Fatalf("wanted to match %v, but only found %v", want, got)
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

		actualSecret, err := kube.CoreV1().Secrets(namespace).Get(context.TODO(), "testsecret", metav1.GetOptions{})
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
			t.Errorf("want status code '%d', got '%d'", http.StatusNotFound, resp.StatusCode)
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
	body, _ := io.ReadAll(resp.Body)

	if string(body) == "null" {
		t.Errorf(`want empty list to be valid json i.e. "[]", but was %q`, string(body))
	}
}
