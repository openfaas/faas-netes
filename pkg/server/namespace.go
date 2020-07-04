package server

import (
	"encoding/json"
	"net/http"

	"github.com/openfaas/faas-netes/handlers"
	"k8s.io/client-go/kubernetes"
)

func makeListNamespaceHandler(defaultNamespace string, clientset kubernetes.Interface) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		res := handlers.ListNamespaces(defaultNamespace, clientset)

		out, _ := json.Marshal(res)
		w.Header().Set("Content-Type", "application/json")

		w.WriteHeader(http.StatusOK)
		w.Write(out)
	}
}
