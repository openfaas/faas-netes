package handlers

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/alexellis/faas-netes/types"
	"github.com/alexellis/faas/gateway/requests"
	"github.com/gorilla/mux"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// MakeReplicaUpdater updates desired count of replicas
func MakeReplicaUpdater(clientset *kubernetes.Clientset) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		functionName := vars["name"]

		req := types.ScaleServiceRequest{}
		if r.Body != nil {
			defer r.Body.Close()
			bytesIn, _ := ioutil.ReadAll(r.Body)
			marshalErr := json.Unmarshal(bytesIn, &req)
			if marshalErr != nil {
				w.WriteHeader(http.StatusBadRequest)
				msg := "Cannot parse request. Please pass valid JSON."
				w.Write([]byte(msg))
				log.Println(msg, marshalErr)
				return
			}
		}

		options := metav1.GetOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Deployment",
				APIVersion: "extensions/v1beta1",
			},
		}
		deployment, err := clientset.ExtensionsV1beta1().Deployments("default").Get(functionName, options)

		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Unable to lookup function deployment " + functionName))
			log.Println(err)
			return
		}

		var replicas int32
		replicas = int32(req.Replicas)
		deployment.Spec.Replicas = &replicas
		_, err = clientset.ExtensionsV1beta1().Deployments("default").Update(deployment)

		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Unable to update function deployment " + functionName))
			log.Println(err)
		}

	}
}

// MakeReplicaReader reads the amount of replicas for a deployment
func MakeReplicaReader(clientset *kubernetes.Clientset) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)
		functionName := vars["name"]

		functions, err := getServiceList(clientset)
		if err != nil {
			w.WriteHeader(500)
			return
		}
		var found *requests.Function
		for _, function := range functions {
			if function.Name == functionName {
				found = &function
				break
			}
		}

		if found == nil {
			w.WriteHeader(404)
		} else {
			w.WriteHeader(200)
			functionBytes, _ := json.Marshal(found)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			w.Write(functionBytes)
		}
	}
}
