package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/openfaas/faas-provider/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	glog "k8s.io/klog"
)

func makeReplicaHandler(defaultNamespace string, kube kubernetes.Interface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		functionName := vars["name"]

		q := r.URL.Query()
		namespace := q.Get("namespace")

		lookupNamespace := defaultNamespace

		if len(namespace) > 0 {
			lookupNamespace = namespace
		}

		if namespace != defaultNamespace {
			http.Error(w, fmt.Sprintf("valid namespaces are: %s", defaultNamespace), http.StatusBadRequest)
			return
		}

		if lookupNamespace == "kube-system" {
			http.Error(w, "unable to list within the kube-system namespace", http.StatusUnauthorized)
			return
		}

		req := types.ScaleServiceRequest{}
		if r.Body != nil {
			defer r.Body.Close()
			bytesIn, _ := ioutil.ReadAll(r.Body)
			if err := json.Unmarshal(bytesIn, &req); err != nil {
				glog.Errorf("Function %s replica invalid JSON: %v", functionName, err)
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				return
			}
		}

		if req.Replicas == 0 {
			http.Error(w, fmt.Errorf("for OpenFaaS CE, replicas cannot be set to 0").Error(), http.StatusBadRequest)
		}

		opts := metav1.GetOptions{}
		dep, err := kube.AppsV1().Deployments(lookupNamespace).Get(r.Context(), functionName, opts)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			glog.Errorf("Function %s get error: %v", functionName, err)
			return
		}

		dep.Spec.Replicas = int32p(int32(req.Replicas))
		_, err = kube.AppsV1().Deployments(lookupNamespace).Update(r.Context(), dep, metav1.UpdateOptions{})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			glog.Errorf("Function %s update error: %v", functionName, err)
			return
		}

		glog.Infof("Function %v replica updated to %v", functionName, req.Replicas)
		w.WriteHeader(http.StatusAccepted)
	}
}
