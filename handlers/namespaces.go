// Copyright (c) Alex Ellis 2017. All rights reserved.
// Copyright 2019 OpenFaaS Author(s)
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// MakeNamespacesLister builds a list of namespaces with an "openfaas" tag, or the default name
func MakeNamespacesLister(defaultNamespace string, clientset kubernetes.Interface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Query namespaces")

		res := list(defaultNamespace, clientset)

		out, _ := json.Marshal(res)
		w.Header().Set("Content-Type", "application/json")

		w.WriteHeader(http.StatusAccepted)
		w.Write(out)
	}
}

func list(defaultNamespace string, clientset kubernetes.Interface) []string {
	listOptions := metav1.ListOptions{}
	namespaces, err := clientset.CoreV1().Namespaces().List(listOptions)

	set := []string{}

	// Assume that an error means that a Role, instead of ClusterRole is being used
	// the Role will not be able to list namespaces, so all functions are in the
	// defaultNamespace
	if err != nil {
		log.Printf("Error listing namespaces: %s", err.Error())
		set = append(set, defaultNamespace)
		return set
	}

	for _, n := range namespaces.Items {
		if _, ok := n.Annotations["openfaas"]; ok {
			set = append(set, n.Name)
		}
	}

	if !findNamespace(defaultNamespace, set) {
		set = append(set, defaultNamespace)
	}

	return set
}

func findNamespace(target string, items []string) bool {
	for _, n := range items {
		if n == target {
			return true
		}
	}
	return false
}
