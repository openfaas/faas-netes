package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	ofv1 "github.com/openfaas/faas-netes/pkg/apis/openfaas/v1"
	clientset "github.com/openfaas/faas-netes/pkg/client/clientset/versioned"
	"github.com/openfaas/faas-provider/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
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
