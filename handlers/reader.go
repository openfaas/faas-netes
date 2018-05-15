// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/openfaas/faas/gateway/requests"
	"k8s.io/api/apps/v1beta2"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func getService(name, functionNamespace string, clientset *kubernetes.Clientset) (*requests.Function, error) {
	deployment, err := clientset.AppsV1beta2().Deployments(functionNamespace).Get(name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return deployment2function(deployment), nil
}

func getServiceList(functionNamespace string, clientset *kubernetes.Clientset) ([]requests.Function, error) {
	listOpts := metav1.ListOptions{
		LabelSelector: "faas_function",
	}

	res, err := clientset.AppsV1beta2().Deployments(functionNamespace).List(listOpts)
	if err != nil {
		return nil, err
	}

	var functions []requests.Function
	for i := range res.Items {
		functions = append(functions, *deployment2function(&res.Items[i])) // &item would be the same pointer!
	}
	return functions, nil
}

func deployment2function(item *v1beta2.Deployment) *requests.Function {
	if item == nil {
		return nil
	}

	var replicas uint64
	if item.Spec.Replicas != nil {
		replicas = uint64(*item.Spec.Replicas)
	}

	labels := item.Spec.Template.Labels
	return &requests.Function{
		Name:              item.Name,
		Replicas:          replicas,
		Image:             item.Spec.Template.Spec.Containers[0].Image,
		AvailableReplicas: uint64(item.Status.AvailableReplicas),
		InvocationCount:   0,
		Labels:            &labels,
	}
}

// MakeFunctionReader handler for reading functions deployed in the cluster as deployments.
func MakeFunctionReader(functionNamespace string, clientset *kubernetes.Clientset) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		functions, err := getServiceList(functionNamespace, clientset)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		functionBytes, _ := json.Marshal(functions)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(functionBytes)
	}
}
