// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/openfaas/faas-netes/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
)

// MakeReplicaUpdater updates desired count of replicas
func MakeReplicaUpdater(defaultNamespace string, clientset *kubernetes.Clientset) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Update replicas")

		vars := mux.Vars(r)

		functionName := vars["name"]
		q := r.URL.Query()
		namespace := q.Get("namespace")

		lookupNamespace := defaultNamespace

		if len(namespace) > 0 {
			lookupNamespace = namespace
		}

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

		if len(req.ServiceNamespace) > 0 {
			lookupNamespace = req.ServiceNamespace
		}

		options := metav1.GetOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Deployment",
				APIVersion: "apps/v1",
			},
		}

		deployment, err := clientset.Apps().Deployments(lookupNamespace).Get(functionName, options)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Unable to lookup function deployment " + functionName))
			log.Println(err)
			return
		}

		oldReplicas := *deployment.Spec.Replicas
		replicas := int32(req.Replicas)

		log.Printf("Set replicas - %s %s, %d/%d\n", functionName, lookupNamespace, replicas, oldReplicas)

		deployment.Spec.Replicas = &replicas

		_, err = clientset.Apps().Deployments(lookupNamespace).Update(deployment)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Unable to update function deployment " + functionName))
			log.Println(err)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	}
}

// MakeReplicaReader reads the amount of replicas for a deployment
func MakeReplicaReader(defaultNamespace string, clientset *kubernetes.Clientset) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)

		functionName := vars["name"]
		q := r.URL.Query()
		namespace := q.Get("namespace")

		lookupNamespace := defaultNamespace

		if len(namespace) > 0 {
			lookupNamespace = namespace
		}
		function, err := getService(lookupNamespace, functionName, clientset)
		if err != nil {
			log.Printf("Unable to fetch service: %s %s\n", functionName, namespace)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if function.AvailableReplicas == 0 && function.Replicas > 0 {
			log.Printf("Checking Replicas > 0 and AvailableReplicas = 0\n")
			opts := metav1.GetOptions{}
			endpoints, err := clientset.CoreV1().Endpoints(lookupNamespace).Get(functionName, opts)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if len(endpoints.Subsets) > 0 {
				log.Printf("Endpoints for %s, subnets %d notready %d ready %d\n", function.Name,
					len(endpoints.Subsets), len(endpoints.Subsets[0].NotReadyAddresses), len(endpoints.Subsets[0].Addresses))

				if len(endpoints.Subsets[0].NotReadyAddresses) > 0 ||
					len(endpoints.Subsets[0].Addresses) > 0 {
					function.AvailableReplicas = 1
				}
			}
		}

		if function == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		log.Printf("Read replicas - %s %s, %d/%d\n", functionName, lookupNamespace, function.AvailableReplicas, function.Replicas)

		functionBytes, _ := json.Marshal(function)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(functionBytes)
	}
}
