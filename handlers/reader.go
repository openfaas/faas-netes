// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	types "github.com/openfaas/faas-provider/types"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/openfaas/faas-netes/k8s"
)

// MakeFunctionReader handler for reading functions deployed in the cluster as deployments.
func MakeFunctionReader(defaultNamespace string, clientset *kubernetes.Clientset) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

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

		functions, err := getServiceList(lookupNamespace, clientset)
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

func getServiceList(functionNamespace string, clientset *kubernetes.Clientset) ([]types.FunctionStatus, error) {
	functions := []types.FunctionStatus{}

	listOpts := metav1.ListOptions{
		LabelSelector: "faas_function",
	}

	res, err := clientset.AppsV1().Deployments(functionNamespace).List(listOpts)

	if err != nil {
		return nil, err
	}

	for _, item := range res.Items {
		function := k8s.AsFunctionStatus(item)
		if function != nil {
			functions = append(functions, *function)
		}
	}
	return functions, nil
}

// getService returns a function/service or nil if not found
func getService(functionNamespace string, functionName string, clientset *kubernetes.Clientset) (*types.FunctionStatus, error) {

	getOpts := metav1.GetOptions{}

	item, err := clientset.AppsV1().Deployments(functionNamespace).Get(functionName, getOpts)

	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}

		return nil, err
	}

	if item != nil {
		function := k8s.AsFunctionStatus(*item)
		if function != nil {
			return function, nil
		}
	}

	return nil, fmt.Errorf("function: %s not found", functionName)
}
