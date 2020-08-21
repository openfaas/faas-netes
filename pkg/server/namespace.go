package server

import (
	"encoding/json"
	"net/http"

	"github.com/openfaas/faas-netes/pkg/handlers"
	"k8s.io/client-go/kubernetes"
	glog "k8s.io/klog"
)

func makeListNamespaceHandler(defaultNamespace string, clientset kubernetes.Interface) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		res := handlers.ListNamespaces(defaultNamespace, clientset)

		out, err := json.Marshal(res)
		if err != nil {
			glog.Errorf("Failed to marshal namespaces: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to marshal namespaces"))
			return
		}

		w.Header().Set("Content-Type", "application/json")

		w.WriteHeader(http.StatusOK)
		w.Write(out)
	}
}
