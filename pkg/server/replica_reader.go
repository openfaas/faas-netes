package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	ofv1 "github.com/openfaas/faas-netes/pkg/apis/openfaas/v1"
	clientset "github.com/openfaas/faas-netes/pkg/client/clientset/versioned"
	"github.com/openfaas/faas-provider/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/listers/apps/v1"
	glog "k8s.io/klog"
)

func makeReplicaReader(defaultNamespace string, client clientset.Interface, lister v1.DeploymentLister) http.HandlerFunc {
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

		opts := metav1.GetOptions{}
		k8sfunc, err := client.OpenfaasV1().Functions(lookupNamespace).
			Get(r.Context(), functionName, opts)
		if err != nil || k8sfunc == nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(err.Error()))
			return
		}

		desiredReplicas, availableReplicas, err := getReplicas(functionName, lookupNamespace, lister)
		if err != nil {
			glog.Warningf("Function replica reader error: %v", err)
		}

		result := toFunctionStatus(*k8sfunc)
		result.AvailableReplicas = availableReplicas
		result.Replicas = desiredReplicas

		res, err := json.Marshal(result)
		if err != nil {
			glog.Errorf("Failed to marshal function status: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to marshal function status"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(res)
	}
}

func getReplicas(functionName string, namespace string, lister v1.DeploymentLister) (uint64, uint64, error) {
	dep, err := lister.Deployments(namespace).Get(functionName)
	if err != nil {
		return 0, 0, err
	}

	desiredReplicas := uint64(dep.Status.Replicas)
	availableReplicas := uint64(dep.Status.AvailableReplicas)

	return desiredReplicas, availableReplicas, nil
}

func toFunctionStatus(item ofv1.Function) types.FunctionStatus {

	status := types.FunctionStatus{
		// AvailableReplicas: availableReplicas,
		// Replicas:          desiredReplicas,
		Labels:                 item.Spec.Labels,
		Annotations:            item.Spec.Annotations,
		Name:                   item.Spec.Name,
		EnvProcess:             item.Spec.Handler,
		Image:                  item.Spec.Image,
		Namespace:              item.Namespace,
		Secrets:                item.Spec.Secrets,
		Constraints:            item.Spec.Constraints,
		ReadOnlyRootFilesystem: item.Spec.ReadOnlyRootFilesystem,
		CreatedAt:              item.CreationTimestamp.Time,
	}

	if item.Spec.Environment != nil {
		status.EnvVars = *item.Spec.Environment
	}

	if item.Spec.Limits != nil {
		status.Limits = (*types.FunctionResources)(item.Spec.Limits)
	}

	if item.Spec.Requests != nil {
		status.Requests = (*types.FunctionResources)(item.Spec.Requests)
	}

	return status
}
