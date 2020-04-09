package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	clientset "github.com/openfaas/faas-netes/pkg/client/clientset/versioned"
	"github.com/openfaas/faas-provider/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/listers/apps/v1"
	glog "k8s.io/klog"
)

func makeReplicaReader(namespace string, client clientset.Interface, lister v1.DeploymentNamespaceLister) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		functionName := vars["name"]

		opts := metav1.GetOptions{}
		k8sfunc, err := client.OpenfaasV1().Functions(namespace).Get(functionName, opts)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(err.Error()))
			return
		}

		desiredReplicas, availableReplicas, err := getReplicas(functionName, namespace, lister)
		if err != nil {
			glog.Warningf("Function replica reader error: %v", err)
		}

		result := &types.FunctionStatus{
			AvailableReplicas: availableReplicas,
			Replicas:          desiredReplicas,
			Labels:            k8sfunc.Spec.Labels,
			Annotations:       k8sfunc.Spec.Annotations,
			Name:              k8sfunc.Spec.Name,
			EnvProcess:        k8sfunc.Spec.Handler,
			Image:             k8sfunc.Spec.Image,
			Namespace:         namespace,
		}

		res, _ := json.Marshal(result)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(res)
	}
}

func getReplicas(functionName string, namespace string, lister v1.DeploymentNamespaceLister) (uint64, uint64, error) {
	dep, err := lister.Get(functionName)
	if err != nil {
		return 0, 0, err
	}
	desiredReplicas := uint64(dep.Status.Replicas)
	availableReplicas := uint64(dep.Status.AvailableReplicas)

	return desiredReplicas, availableReplicas, nil
}

func makeReplicaHandler(namespace string, kube kubernetes.Interface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		functionName := vars["name"]

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
		dep, err := kube.AppsV1().Deployments(namespace).Get(functionName, opts)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			glog.Errorf("Function %s get error: %v", functionName, err)
			return
		}

		dep.Spec.Replicas = int32p(int32(req.Replicas))
		_, err = kube.AppsV1().Deployments(namespace).Update(dep)
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
