// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"

	"github.com/alexellis/faas/gateway/requests"
	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

const DefaultFunctionNamespace string = "default"

// ValidateDeployRequest validates that the service name is valid for Kubernetes
func ValidateDeployRequest(request *requests.CreateFunctionRequest) error {
	var validDNS = regexp.MustCompile(`^[a-zA-Z\-]+$`)
	matched := validDNS.MatchString(request.Service)
	if matched {
		return nil
	}

	return fmt.Errorf("(%s) must be a valid DNS entry for service name", request.Service)
}

// MakeDeployHandler creates a handler to create new functions in the cluster
func MakeDeployHandler(functionNamespace string, clientset *kubernetes.Clientset) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		body, _ := ioutil.ReadAll(r.Body)

		request := requests.CreateFunctionRequest{}
		err := json.Unmarshal(body, &request)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := ValidateDeployRequest(&request); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		deploymentSpec := makeDeploymentSpec(request)
		deploy := clientset.Extensions().Deployments(functionNamespace)

		_, err = deploy.Create(deploymentSpec)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		log.Println("Created deployment - " + request.Service)

		service := clientset.Core().Services(functionNamespace)
		serviceSpec := makeServiceSpec(request)
		_, err = service.Create(serviceSpec)

		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		log.Println("Created service - " + request.Service)
		log.Println(string(body))

		w.WriteHeader(http.StatusAccepted)

	}
}

func makeDeploymentSpec(request requests.CreateFunctionRequest) *v1beta1.Deployment {
	envVars := buildEnvVars(request)

	deploymentSpec := &v1beta1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: request.Service,
		},
		Spec: v1beta1.DeploymentSpec{
			Replicas: int32p(1),
			Strategy: v1beta1.DeploymentStrategy{
				Type: v1beta1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &v1beta1.RollingUpdateDeployment{
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: int32(0),
					},
					MaxSurge: &intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: int32(1),
					},
				},
			},
			RevisionHistoryLimit: int32p(10),
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: request.Service,
					Labels: map[string]string{
						"faas_function": request.Service,
						// "uid":           fmt.Sprintf("%d", time.Now().Nanosecond()),
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  request.Service,
							Image: request.Image,
							Ports: []apiv1.ContainerPort{
								{ContainerPort: int32(8080), Protocol: v1.ProtocolTCP},
							},
							Env: envVars,
							Resources: apiv1.ResourceRequirements{
								Limits: apiv1.ResourceList{
								//v1.ResourceCPU:    resource.MustParse("100m"),
								//v1.ResourceMemory: resource.MustParse("256Mi"),
								},
							},
							ImagePullPolicy: v1.PullAlways,
						},
					},
					RestartPolicy: v1.RestartPolicyAlways,
					DNSPolicy:     v1.DNSClusterFirst,
				},
			},
		},
	}
	return deploymentSpec
}

func makeServiceSpec(request requests.CreateFunctionRequest) *v1.Service {
	serviceSpec := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: request.Service,
		},
		Spec: v1.ServiceSpec{
			Type:     v1.ServiceTypeClusterIP,
			Selector: map[string]string{"faas_function": request.Service},
			Ports: []v1.ServicePort{
				{
					Protocol: v1.ProtocolTCP,
					Port:     8080,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: int32(8080),
					},
				},
			},
		},
	}
	return serviceSpec
}

func buildEnvVars(request requests.CreateFunctionRequest) []v1.EnvVar {
	envVars := []v1.EnvVar{}

	if len(request.EnvProcess) > 0 {
		envVar := v1.EnvVar{
			Name:  "fprocess",
			Value: request.EnvProcess,
		}
		envVars = append(envVars, envVar)
	}

	for k, v := range request.EnvVars {
		if len(request.EnvProcess) > 0 {
			envVar := v1.EnvVar{
				Name:  k,
				Value: v,
			}
			envVars = append(envVars, envVar)
		}
	}
	return envVars
}

func int32p(i int32) *int32 {
	return &i
}
