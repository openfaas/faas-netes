package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/alexellis/faas/gateway/requests"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// MakeFunctionReader handler for reading functions deployed in the cluster as deployments.
func MakeFunctionReader(clientset *kubernetes.Clientset) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var functions []requests.Function

		listOpts := metav1.ListOptions{
			LabelSelector: "faas_function=true",
		}

		res, err := clientset.ExtensionsV1beta1().Deployments("default").List(listOpts)

		if err != nil {
			log.Println(err)
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}

		// fmt.Printf("%d\n", len(res.Items))

		for _, item := range res.Items {
			var replicas uint64
			if item.Spec.Replicas != nil {
				replicas = uint64(*item.Spec.Replicas)
			}

			function := requests.Function{
				Name:            item.Name,
				Replicas:        replicas,
				Image:           item.Spec.Template.Spec.Containers[0].Image,
				InvocationCount: 0,
			}
			functions = append(functions, function)
		}

		functionBytes, _ := json.Marshal(functions)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(functionBytes)
	}
}
