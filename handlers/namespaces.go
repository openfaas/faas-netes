// Copyright (c) Alex Ellis 2017. All rights reserved.
// Copyright 2019 OpenFaaS Author(s)
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

		w.WriteHeader(http.StatusOK)
		w.Write(out)
	}
}

// NamespaceResolver is a method that determines the requested namespace corresponding to an
// HTTP request. It then validates that OpenFaaS is permitted to operate in that the namespace.
type NamespaceResolver func(r *http.Request) (namespace string, err error)

// NewNamespaceResolver returns a generic namespace resolver that will inspect both the GET query
// parameters and the request body. It looks for the query param or json key "namespace".
func NewNamespaceResolver(defaultNamespace string, kube kubernetes.Interface) NamespaceResolver {
	return func(r *http.Request) (string, error) {
		req := struct{ Namespace string }{Namespace: defaultNamespace}

		switch r.Method {
		case http.MethodGet:
			q := r.URL.Query()
			namespace := q.Get("namespace")

			if len(namespace) > 0 {
				req.Namespace = namespace
			}

		case http.MethodPost, http.MethodPut, http.MethodDelete:
			body, _ := ioutil.ReadAll(r.Body)
			err := json.Unmarshal(body, &req)
			if err != nil {
				log.Printf("error while getting namespace: %s\n", err)
				return "", fmt.Errorf("unable to unmarshal json request")
			}

			// Reconstruct Body
			r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		}

		allowedNamespaces := list(defaultNamespace, kube)
		ok := findNamespace(req.Namespace, allowedNamespaces)
		if !ok {
			return req.Namespace, fmt.Errorf("unable to manage secrets within the %s namespace", req.Namespace)
		}

		return req.Namespace, nil
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
