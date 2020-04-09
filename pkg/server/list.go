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

func makeListHandler(namespace string,
	client clientset.Interface,
	deploymentLister appsv1.DeploymentNamespaceLister) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			defer r.Body.Close()
		}

		functions := []types.FunctionStatus{}

		opts := metav1.ListOptions{}
		res, err := client.OpenfaasV1().Functions(namespace).List(opts)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			glog.Errorf("Function listing error: %v", err)
			return
		}

		for _, item := range res.Items {

			desiredReplicas, availableReplicas, err := getReplicas(item.Spec.Name, namespace, deploymentLister)
			if err != nil {
				glog.Warningf("Function listing getReplicas error: %v", err)
			}

			function := types.FunctionStatus{
				Name:              item.Spec.Name,
				Replicas:          desiredReplicas,
				AvailableReplicas: availableReplicas,
				Image:             item.Spec.Image,
				Labels:            item.Spec.Labels,
				Annotations:       item.Spec.Annotations,
				Namespace:         namespace,
			}

			functions = append(functions, function)
		}

		functionBytes, _ := json.Marshal(functions)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(functionBytes)

	}
}
