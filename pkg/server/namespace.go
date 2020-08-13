package server

import (
	"context"
	"encoding/json"
	clientset "github.com/openfaas/faas-netes/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	glog "k8s.io/klog"
	"net/http"
)

func makeListNamespaceHandler(
	client clientset.Interface) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		defer r.Body.Close()

		opts := metav1.ListOptions{}
		res, err := client.OpenfaasV1().Functions(v1.NamespaceAll).List(context.TODO(), opts)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			glog.Errorf("Namespace listing error: %v", err)
			return
		}

		m := make(map[string]interface{})
		for _, item := range res.Items {
			m[item.Namespace] = nil
		}

		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}

		res2, _ := json.Marshal(keys)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(res2)
	}
}
