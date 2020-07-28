package server

import (
	"net/http"

	"github.com/openfaas/faas-netes/pkg/k8s"

	"github.com/openfaas/faas-netes/pkg/handlers"
	"k8s.io/client-go/kubernetes"
)

const (
	secretLabel      = "app.kubernetes.io/managed-by"
	secretLabelValue = "openfaas"
)

// makeSecretHandler provides the secrets CRUD endpoint
func makeSecretHandler(namespace string, kube kubernetes.Interface) http.HandlerFunc {
	handler := handlers.SecretsHandler{
		LookupNamespace: handlers.NewNamespaceResolver(namespace, kube),
		Secrets:         k8s.NewSecretsClient(kube),
	}
	return handler.ServeHTTP
}
