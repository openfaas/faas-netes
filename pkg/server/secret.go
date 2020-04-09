package server

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	faastypes "github.com/openfaas/faas-provider/types"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	glog "k8s.io/klog"
)

const (
	secretLabel      = "app.kubernetes.io/managed-by"
	secretLabelValue = "openfaas"
)

// makeSecretHandler provides the secrets CRUD endpoint
func makeSecretHandler(namespace string, kube kubernetes.Interface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			defer r.Body.Close()
		}

		switch r.Method {
		case http.MethodGet:
			secrets, err := getSecrets(namespace, kube)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				glog.Errorf("Secrets query error: %v", err)
				return
			}

			secretsBytes, err := json.Marshal(secrets)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				glog.Errorf("Secrets json marshal error: %v", err)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(secretsBytes)
		case http.MethodPost:
			secret, err := parseSecret(namespace, r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				glog.Errorf("Secret unmarshal error: %v", err)
				return
			}

			if err := createSecret(namespace, kube, secret); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				glog.Errorf("Secret create error: %v", err)
				return
			}

			glog.Infof("Secret %s created", secret.GetName())
			w.WriteHeader(http.StatusAccepted)
		case http.MethodPut:
			secret, err := parseSecret(namespace, r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				glog.Errorf("Secret unmarshal error: %v", err)
				return
			}

			if err := updateSecret(namespace, kube, secret); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				glog.Errorf("Secret update error: %v", err)
				return
			}

			glog.Infof("Secret %s updated", secret.GetName())
			w.WriteHeader(http.StatusAccepted)
		case http.MethodDelete:
			secret, err := parseSecret(namespace, r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				glog.Errorf("Secret unmarshal error: %v", err)
				return
			}

			if err := deleteSecret(namespace, kube, secret); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				glog.Errorf("Secret %s delete error: %v", secret.GetName(), err)
				return
			}

			glog.Infof("Secret %s deleted", secret.GetName())
			w.WriteHeader(http.StatusAccepted)
		default:
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
}

func getSecrets(namespace string, kube kubernetes.Interface) ([]faastypes.Secret, error) {
	secrets := []faastypes.Secret{}
	selector := fmt.Sprintf("%s=%s", secretLabel, secretLabelValue)

	res, err := kube.CoreV1().Secrets(namespace).List(metav1.ListOptions{LabelSelector: selector})

	if err != nil {
		return secrets, err
	}

	for _, item := range res.Items {
		secret := faastypes.Secret{
			Name: item.Name,
		}
		secrets = append(secrets, secret)
	}

	return secrets, nil
}

func createSecret(namespace string, kube kubernetes.Interface, secret *corev1.Secret) error {
	_, err := kube.CoreV1().Secrets(namespace).Create(secret)
	if err != nil {
		return err
	}
	return nil
}

func updateSecret(namespace string, kube kubernetes.Interface, secret *corev1.Secret) error {
	s, err := kube.CoreV1().Secrets(namespace).Get(secret.GetName(), metav1.GetOptions{})
	if err != nil {
		return err
	}
	s.StringData = secret.StringData
	if _, err = kube.CoreV1().Secrets(namespace).Update(s); err != nil {
		return err
	}
	return nil
}

func deleteSecret(namespace string, kube kubernetes.Interface, secret *corev1.Secret) error {
	if err := kube.CoreV1().Secrets(namespace).Delete(secret.GetName(), &metav1.DeleteOptions{}); err != nil {
		return err
	}
	return nil
}

func parseSecret(namespace string, r io.Reader) (*corev1.Secret, error) {
	body, _ := ioutil.ReadAll(r)
	req := faastypes.Secret{}
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, err
	}
	secret := &corev1.Secret{
		Type: corev1.SecretTypeOpaque,
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: namespace,
			Labels: map[string]string{
				secretLabel: secretLabelValue,
			},
		},
		StringData: map[string]string{
			req.Name: req.Value,
		},
	}

	return secret, nil
}
