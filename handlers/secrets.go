// Copyright 2019 OpenFaaS Author(s)
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/openfaas/faas-netes/k8s"
	types "github.com/openfaas/faas-provider/types"
	"k8s.io/client-go/kubernetes"
)

// MakeSecretHandler makes a handler for Create/List/Delete/Update of
//secrets in the Kubernetes API
func MakeSecretHandler(defaultNamespace string, kube kubernetes.Interface) http.HandlerFunc {
	handler := secretsHandler{
		lookupNamespace: NewNamespaceResolver(defaultNamespace, kube),
		secrets:         k8s.NewSecretsClient(kube),
	}
	return handler.ServeHTTP
}

type secretsHandler struct {
	secrets         k8s.SecretsClient
	lookupNamespace NamespaceResolver
}

func (h secretsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}

	lookupNamespace, err := h.lookupNamespace(r)
	if err != nil {
		switch err.Error() {
		case "unable to unmarshal Secret request":
			http.Error(w, err.Error(), http.StatusBadRequest)
		case "unable to manage secrets within the specified namespace":
			http.Error(w, err.Error(), http.StatusUnauthorized)
		default:
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	switch r.Method {
	case http.MethodGet:
		h.listSecrets(lookupNamespace, w, r)
	case http.MethodPost:
		h.createSecret(lookupNamespace, w, r)
	case http.MethodPut:
		h.replaceSecret(lookupNamespace, w, r)
	case http.MethodDelete:
		h.deleteSecret(lookupNamespace, w, r)
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (h secretsHandler) listSecrets(namespace string, w http.ResponseWriter, r *http.Request) {
	res, err := h.secrets.List(namespace)
	if err != nil {
		status, reason := ProcessErrorReasons(err)
		log.Printf("Secret list error reason: %s, %v\n", reason, err)
		w.WriteHeader(status)
		return
	}

	secrets := make([]types.Secret, len(res))
	for idx, name := range res {
		secrets[idx] = types.Secret{
			Name:      name,
			Namespace: namespace,
		}
	}
	secretsBytes, err := json.Marshal(secrets)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Secrets json marshal error: %v\n", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(secretsBytes)
}

func (h secretsHandler) createSecret(namespace string, w http.ResponseWriter, r *http.Request) {
	secret := types.Secret{}
	err := json.NewDecoder(r.Body).Decode(&secret)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Secret unmarshal error: %v\n", err)
		return
	}

	secret.Namespace = namespace
	err = h.secrets.Create(secret)
	if err != nil {
		status, reason := ProcessErrorReasons(err)
		log.Printf("Secret create error reason: %s, %v\n", reason, err)
		w.WriteHeader(status)
		return
	}
	log.Printf("Secret %s create\n", secret.Name)
	w.WriteHeader(http.StatusAccepted)
}

func (h secretsHandler) replaceSecret(namespace string, w http.ResponseWriter, r *http.Request) {
	secret := types.Secret{}
	err := json.NewDecoder(r.Body).Decode(&secret)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Secret unmarshal error: %v\n", err)
		return
	}

	secret.Namespace = namespace
	err = h.secrets.Replace(secret)
	if err != nil {
		status, reason := ProcessErrorReasons(err)
		log.Printf("Secret update error reason: %s, %v\n", reason, err)
		w.WriteHeader(status)
		return
	}
	log.Printf("Secret %s updated", secret.Name)
	w.WriteHeader(http.StatusAccepted)
}

func (h secretsHandler) deleteSecret(namespace string, w http.ResponseWriter, r *http.Request) {
	secret := types.Secret{}
	err := json.NewDecoder(r.Body).Decode(&secret)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Secret unmarshal error: %v\n", err)
		return
	}

	err = h.secrets.Delete(namespace, secret.Name)
	if err != nil {
		status, reason := ProcessErrorReasons(err)
		log.Printf("Secret delete error reason: %s, %v\n", reason, err)
		w.WriteHeader(status)
		return
	}
	log.Printf("Secret %s deleted\n", secret.Name)
	w.WriteHeader(http.StatusAccepted)
}
