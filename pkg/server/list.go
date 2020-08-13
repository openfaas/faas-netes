package server

import (
	"context"
	"encoding/json"
	appsinformer "k8s.io/client-go/informers/apps/v1"
	"net/http"

	clientset "github.com/openfaas/faas-netes/pkg/client/clientset/versioned"
	"github.com/openfaas/faas-provider/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	glog "k8s.io/klog"
)

func makeListHandler(namespace string,
	client clientset.Interface, deploymentsInformer appsinformer.DeploymentInformer) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			defer r.Body.Close()
		}

		query := r.URL.Query()

		if query.Get("namespace") != "" {
			namespace = query.Get("namespace")
		}

		functions := []types.FunctionStatus{}

		opts := metav1.ListOptions{}
		res, err := client.OpenfaasV1().Functions(namespace).List(context.TODO(), opts)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			glog.Errorf("Function listing error: %v", err)
			return
		}

		for _, item := range res.Items {
			deploymentLister := deploymentsInformer.Lister().Deployments(item.Namespace)

			desiredReplicas, availableReplicas, err := getReplicas(item.Spec.Name, item.Namespace, deploymentLister)
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
				Namespace:         item.Namespace,
			}

			functions = append(functions, function)
		}

		functionBytes, _ := json.Marshal(functions)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(functionBytes)

	}
}
