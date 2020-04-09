package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	faastypes "github.com/openfaas/faas-provider/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testclient "k8s.io/client-go/kubernetes/fake"
)

func Test_makeSecretHandler(t *testing.T) {
	namespace := "openfaas-fn"
	kube := testclient.NewSimpleClientset()
	secretsHandler := makeSecretHandler(namespace, kube).ServeHTTP

	secretName := "testsecret"

	t.Run("create managed secrets", func(t *testing.T) {
		secretValue := "testsecretvalue"
		payload := fmt.Sprintf(`{"name": "%s", "value": "%s"}`, secretName, secretValue)
		req := httptest.NewRequest("POST", "http://system/secrets", strings.NewReader(payload))
		w := httptest.NewRecorder()

		secretsHandler(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusAccepted {
			t.Errorf("expected status code '%d', got '%d'", http.StatusAccepted, resp.StatusCode)
		}

		actualSecret, err := kube.CoreV1().Secrets(namespace).Get("testsecret", metav1.GetOptions{})
		if err != nil {
			t.Errorf("error validating secret: %s", err)
		}

		if actualSecret.Name != "testsecret" {
			t.Errorf("expected secret with name: 'testsecret', got: '%s'", actualSecret.Name)
		}

		managedBy := actualSecret.Labels[secretLabel]
		if managedBy != secretLabelValue {
			t.Errorf("expected secret to be managed by '%s', got: '%s'", secretLabelValue, managedBy)
		}

		actualValue := actualSecret.StringData[secretName]
		if actualValue != secretValue {
			t.Errorf("expected secret value: '%s', got: '%s'", secretValue, actualValue)
		}
	})

	t.Run("update managed secrets", func(t *testing.T) {
		newSecretValue := "newtestsecretvalue"
		payload := fmt.Sprintf(`{"name": "%s", "value": "%s"}`, secretName, newSecretValue)
		req := httptest.NewRequest("PUT", "http://system/secrets", strings.NewReader(payload))
		w := httptest.NewRecorder()

		secretsHandler(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusAccepted {
			t.Errorf("expected status code '%d', got '%d'", http.StatusAccepted, resp.StatusCode)
		}

		actualSecret, err := kube.CoreV1().Secrets(namespace).Get("testsecret", metav1.GetOptions{})
		if err != nil {
			t.Errorf("error validting secret: %s", err)
		}

		if actualSecret.Name != "testsecret" {
			t.Errorf("expected secret with name: 'testsecret', got: '%s'", actualSecret.Name)
		}

		managedBy := actualSecret.Labels[secretLabel]
		if managedBy != secretLabelValue {
			t.Errorf("expected secret to be managed by '%s', got: '%s'", secretLabelValue, managedBy)
		}

		actualValue := actualSecret.StringData[secretName]
		if actualValue != newSecretValue {
			t.Errorf("expected secret value: '%s', got: '%s'", newSecretValue, actualValue)
		}
	})

	t.Run("list managed secrets only", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://system/secrets", nil)
		w := httptest.NewRecorder()

		secretsHandler(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status code '%d', got '%d'", http.StatusOK, resp.StatusCode)
		}

		decoder := json.NewDecoder(resp.Body)

		secretList := []faastypes.Secret{}
		err := decoder.Decode(&secretList)
		if err != nil {
			t.Error(err)
		}

		if len(secretList) != 1 {
			t.Errorf("expected 1 secret, got %d", len(secretList))
		}

		actualSecret := secretList[0]
		if actualSecret.Name != secretName {
			t.Errorf("expected secret name: '%s', got: '%s'", secretName, actualSecret.Name)
		}
	})

	t.Run("delete managed secrets", func(t *testing.T) {
		secretName := "testsecret"
		payload := fmt.Sprintf(`{"name": "%s"}`, secretName)
		req := httptest.NewRequest("DELETE", "http://system/secrets", strings.NewReader(payload))
		w := httptest.NewRecorder()

		secretsHandler(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusAccepted {
			t.Errorf("expected status code '%d', got '%d'", http.StatusAccepted, resp.StatusCode)
		}

		actualSecret, err := kube.CoreV1().Secrets(namespace).Get("testsecret", metav1.GetOptions{})
		if err == nil {
			t.Errorf("expected not found error, got secret payload '%s'", actualSecret)
		}
	})
}

func Test_makeSecretHandler_WithEmptyList(t *testing.T) {
	namespace := "openfaas-fn"
	kube := testclient.NewSimpleClientset()
	secretsHandler := makeSecretHandler(namespace, kube).ServeHTTP

	req := httptest.NewRequest("GET", "http://system/secrets", nil)
	w := httptest.NewRecorder()

	secretsHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status code '%d', got '%d'", http.StatusOK, resp.StatusCode)
	}

	body, _ := ioutil.ReadAll(resp.Body)

	if string(body) == "null" {
		t.Errorf(`want empty list to be valid json i.e. "[]", but was %q`, string(body))
	}
}
