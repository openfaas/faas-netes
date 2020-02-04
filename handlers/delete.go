// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/openfaas/faas/gateway/requests"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// MakeDeleteHandler delete a function
func MakeDeleteHandler(defaultNamespace string, clientset *kubernetes.Clientset) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

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

		body, _ := ioutil.ReadAll(r.Body)

		request := requests.DeleteFunctionRequest{}
		err := json.Unmarshal(body, &request)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if len(request.FunctionName) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		getOpts := metav1.GetOptions{}

		// This makes sure we don't delete non-labelled deployments
		deployment, findDeployErr := clientset.AppsV1().
			Deployments(lookupNamespace).
			Get(request.FunctionName, getOpts)

		if findDeployErr != nil {
			if errors.IsNotFound(findDeployErr) {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}

			w.Write([]byte(findDeployErr.Error()))
			return
		}

		if isFunction(deployment) {
			err := deleteFunction(lookupNamespace, clientset, request, w)
			if err != nil {
				return
			}
		} else {
			w.WriteHeader(http.StatusBadRequest)

			w.Write([]byte("Not a function: " + request.FunctionName))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		return
	}
}

func isFunction(deployment *appsv1.Deployment) bool {
	if deployment != nil {
		if _, found := deployment.Labels["faas_function"]; found {
			return true
		}
	}
	return false
}

func deleteFunction(functionNamespace string, clientset *kubernetes.Clientset, request requests.DeleteFunctionRequest, w http.ResponseWriter) error {
	foregroundPolicy := metav1.DeletePropagationForeground
	opts := &metav1.DeleteOptions{PropagationPolicy: &foregroundPolicy}

	if deployErr := clientset.AppsV1().Deployments(functionNamespace).
		Delete(request.FunctionName, opts); deployErr != nil {

		if errors.IsNotFound(deployErr) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Write([]byte(deployErr.Error()))
		return fmt.Errorf("error deleting function's deployment")
	}

	if svcErr := clientset.CoreV1().
		Services(functionNamespace).
		Delete(request.FunctionName, opts); svcErr != nil {

		if errors.IsNotFound(svcErr) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		w.Write([]byte(svcErr.Error()))
		return fmt.Errorf("error deleting function's service")
	}
	return nil
}
