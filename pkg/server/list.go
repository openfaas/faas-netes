package server

import (
	"encoding/json"
	"net/http"

	clientset "github.com/openfaas/faas-netes/pkg/client/clientset/versioned"
	"github.com/openfaas/faas-provider/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsv1 "k8s.io/client-go/listers/apps/v1"
	glog "k8s.io/klog"
)

func makeListHandler(defaultNamespace string,
	client clientset.Interface,
	deploymentLister appsv1.DeploymentLister) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			defer r.Body.Close()
		}

		q := r.URL.Query()
		namespace := q.Get("namespace")

		lookupNamespace := defaultNamespace

		if len(namespace) > 0 {
			lookupNamespace = namespace
		}

		if lookupNamespace == "kube-system" {
			http.Error(w, "unable to list within the kube-system namespace", http.StatusUnauthorized)
			return
		}

		opts := metav1.ListOptions{}
		res, err := client.OpenfaasV1().Functions(lookupNamespace).List(r.Context(), opts)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			glog.Errorf("Function listing error: %v", err)
			return
		}

		functions := []types.FunctionStatus{}
		for _, item := range res.Items {
			desiredReplicas, availableReplicas, err := getReplicas(item.Spec.Name, lookupNamespace, deploymentLister)
			if err != nil {
				glog.Warningf("Function listing getReplicas error: %v", err)
			}
			function := toFunctionStatus(item)
			function.AvailableReplicas = availableReplicas
			function.Replicas = desiredReplicas

			functions = append(functions, function)
		}

		functionBytes, err := json.Marshal(functions)
		if err != nil {
			glog.Errorf("Failed to marshal functions: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to marshal functions"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(functionBytes)
	}
}
