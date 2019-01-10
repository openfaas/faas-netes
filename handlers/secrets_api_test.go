package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/openfaas/faas/gateway/requests"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testclient "k8s.io/client-go/kubernetes/fake"
)

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

		secretList := []requests.Secret{}
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
