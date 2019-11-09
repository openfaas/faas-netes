// Copyright 2019 OpenFaaS Author(s)
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/openfaas/faas-netes/k8s"

	types "github.com/openfaas/faas-provider/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MakeUpdateHandler update specified function
func MakeUpdateHandler(defaultNamespace string, factory k8s.FunctionFactory) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			defer r.Body.Close()
		}

		body, _ := ioutil.ReadAll(r.Body)

		request := types.FunctionDeployment{}
		err := json.Unmarshal(body, &request)
		if err != nil {
			wrappedErr := fmt.Errorf("unable to unmarshal request: %s", err.Error())
			http.Error(w, wrappedErr.Error(), http.StatusBadRequest)
			return
		}

		lookupNamespace := defaultNamespace
		if len(request.Namespace) > 0 {
			lookupNamespace = request.Namespace
		}

		if lookupNamespace == "kube-system" {
			http.Error(w, "unable to list within the kube-system namespace", http.StatusUnauthorized)
			return
		}

		annotations := buildAnnotations(request)
		if err, status := updateDeploymentSpec(lookupNamespace, factory, request, annotations); err != nil {
			if !k8s.IsNotFound(err) {
				log.Printf("error updating deployment: %s.%s, error: %s\n", request.Service, lookupNamespace, err)

				return
			}

			wrappedErr := fmt.Errorf("unable update Deployment: %s.%s, error: %s", request.Service, lookupNamespace, err.Error())
			http.Error(w, wrappedErr.Error(), status)
			return
		}

		if err, status := updateService(lookupNamespace, factory, request, annotations); err != nil {
			if !k8s.IsNotFound(err) {
				log.Printf("error updating service: %s.%s, error: %s\n", request.Service, lookupNamespace, err)
			}

			wrappedErr := fmt.Errorf("unable update Service: %s.%s, error: %s", request.Service, request.Namespace, err.Error())
			http.Error(w, wrappedErr.Error(), status)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	}
}

func updateDeploymentSpec(
	functionNamespace string,
	factory k8s.FunctionFactory,
	request types.FunctionDeployment,
	annotations map[string]string) (err error, httpStatus int) {

	getOpts := metav1.GetOptions{}

	deployment, findDeployErr := factory.Client.AppsV1().
		Deployments(functionNamespace).
		Get(request.Service, getOpts)

	if findDeployErr != nil {
		return findDeployErr, http.StatusNotFound
	}

	if len(deployment.Spec.Template.Spec.Containers) > 0 {
		deployment.Spec.Template.Spec.Containers[0].Image = request.Image

		// Disabling update support to prevent unexpected mutations of deployed functions,
		// since imagePullPolicy is now configurable. This could be reconsidered later depending
		// on desired behavior, but will need to be updated to take config.
		//deployment.Spec.Template.Spec.Containers[0].ImagePullPolicy = v1.PullAlways

		deployment.Spec.Template.Spec.Containers[0].Env = buildEnvVars(&request)

		factory.ConfigureReadOnlyRootFilesystem(request, deployment)
		factory.ConfigureContainerUserID(deployment)

		deployment.Spec.Template.Spec.NodeSelector = createSelector(request.Constraints)

		labels := map[string]string{
			"faas_function": request.Service,
			"uid":           fmt.Sprintf("%d", time.Now().Nanosecond()),
		}

		if request.Labels != nil {
			if min := getMinReplicaCount(*request.Labels); min != nil {
				deployment.Spec.Replicas = min
			}

			for k, v := range *request.Labels {
				labels[k] = v
			}
		}

		// deployment.Labels = labels
		deployment.Spec.Template.ObjectMeta.Labels = labels

		deployment.Annotations = annotations
		deployment.Spec.Template.Annotations = annotations
		deployment.Spec.Template.ObjectMeta.Annotations = annotations

		resources, resourceErr := createResources(request)
		if resourceErr != nil {
			return resourceErr, http.StatusBadRequest
		}

		deployment.Spec.Template.Spec.Containers[0].Resources = *resources

		var serviceAccount string

		if request.Annotations != nil {
			annotations := *request.Annotations
			if val, ok := annotations["com.openfaas.serviceaccount"]; ok && len(val) > 0 {
				serviceAccount = val
			}
		}

		deployment.Spec.Template.Spec.ServiceAccountName = serviceAccount

		secrets := k8s.NewSecretsClient(factory.Client)
		existingSecrets, err := secrets.GetSecrets(functionNamespace, request.Secrets)
		if err != nil {
			return err, http.StatusBadRequest
		}

		err = factory.ConfigureSecrets(request, deployment, existingSecrets)
		if err != nil {
			log.Println(err)
			return err, http.StatusBadRequest
		}

		probes, err := factory.MakeProbes(request)
		if err != nil {
			return err, http.StatusBadRequest
		}

		deployment.Spec.Template.Spec.Containers[0].LivenessProbe = probes.Liveness
		deployment.Spec.Template.Spec.Containers[0].ReadinessProbe = probes.Readiness
	}

	if _, updateErr := factory.Client.AppsV1().
		Deployments(functionNamespace).
		Update(deployment); updateErr != nil {

		return updateErr, http.StatusInternalServerError
	}

	return nil, http.StatusAccepted
}

func updateService(
	functionNamespace string,
	factory k8s.FunctionFactory,
	request types.FunctionDeployment,
	annotations map[string]string) (err error, httpStatus int) {

	getOpts := metav1.GetOptions{}

	service, findServiceErr := factory.Client.CoreV1().
		Services(functionNamespace).
		Get(request.Service, getOpts)

	if findServiceErr != nil {
		return findServiceErr, http.StatusNotFound
	}

	service.Annotations = annotations

	if _, updateErr := factory.Client.CoreV1().
		Services(functionNamespace).
		Update(service); updateErr != nil {

		return updateErr, http.StatusInternalServerError
	}

	return nil, http.StatusAccepted
}
