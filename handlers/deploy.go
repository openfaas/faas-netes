// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/openfaas/faas/gateway/requests"
	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

// watchdogPort for the OpenFaaS function watchdog
const watchdogPort = 8080

// initialReplicasCount how many replicas to start of creating for a function
const initialReplicasCount = 1

// ValidateDeployRequest validates that the service name is valid for Kubernetes
func ValidateDeployRequest(request *requests.CreateFunctionRequest) error {
	// Regex for RFC-1123 validation:
	// 	k8s.io/kubernetes/pkg/util/validation/validation.go
	var validDNS = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
	matched := validDNS.MatchString(request.Service)
	if matched {
		return nil
	}

	return fmt.Errorf("(%s) must be a valid DNS entry for service name", request.Service)
}

// DeployHandlerConfig specify options for Deployments
type DeployHandlerConfig struct {
	EnableFunctionReadinessProbe bool
	ImagePullPolicy              string
}

// MakeDeployHandler creates a handler to create new functions in the cluster
func MakeDeployHandler(functionNamespace string, clientset *kubernetes.Clientset, config *DeployHandlerConfig) http.HandlerFunc {
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

		existingSecrets, err := getSecrets(clientset, functionNamespace, request.Secrets)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		deploymentSpec, specErr := makeDeploymentSpec(request, existingSecrets, config)

		if specErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(specErr.Error()))
			return
		}

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

func makeDeploymentSpec(request requests.CreateFunctionRequest, existingSecrets map[string]*apiv1.Secret, config *DeployHandlerConfig) (*v1beta1.Deployment, error) {
	envVars := buildEnvVars(&request)
	path := filepath.Join(os.TempDir(), ".lock")
	probe := &apiv1.Probe{
		Handler: apiv1.Handler{
			Exec: &apiv1.ExecAction{
				Command: []string{"cat", path},
			},
		},
		InitialDelaySeconds: 3,
		TimeoutSeconds:      1,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}
	if !config.EnableFunctionReadinessProbe {
		probe = nil
	}

	initialReplicas := int32p(initialReplicasCount)
	labels := map[string]string{
		"faas_function": request.Service,
	}

	if request.Labels != nil {
		if min := getMinReplicaCount(*request.Labels); min != nil {
			initialReplicas = min
		}
		for k, v := range *request.Labels {
			labels[k] = v
		}
	}

	nodeSelector := createSelector(request.Constraints)

	resources, resourceErr := createResources(request)

	if resourceErr != nil {
		return nil, resourceErr
	}

	var imagePullPolicy apiv1.PullPolicy
	switch config.ImagePullPolicy {
	case "Never":
		imagePullPolicy = apiv1.PullNever
	case "IfNotPresent":
		imagePullPolicy = apiv1.PullIfNotPresent
	default:
		imagePullPolicy = apiv1.PullAlways
	}

	deploymentSpec := &v1beta1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: request.Service,
		},
		Spec: v1beta1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"faas_function": request.Service,
				},
			},
			Replicas: initialReplicas,
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
					Name:        request.Service,
					Labels:      labels,
					Annotations: map[string]string{"prometheus.io.scrape": "false"},
				},
				Spec: apiv1.PodSpec{
					NodeSelector: nodeSelector,
					Containers: []apiv1.Container{
						{
							Name:  request.Service,
							Image: request.Image,
							Ports: []apiv1.ContainerPort{
								{ContainerPort: int32(watchdogPort), Protocol: v1.ProtocolTCP},
							},
							Env:             envVars,
							Resources:       *resources,
							ImagePullPolicy: imagePullPolicy,
							LivenessProbe:   probe,
							ReadinessProbe:  probe,
						},
					},
					RestartPolicy: v1.RestartPolicyAlways,
					DNSPolicy:     v1.DNSClusterFirst,
				},
			},
		},
	}

	if err := UpdateSecrets(request, deploymentSpec, existingSecrets); err != nil {
		return nil, err
	}

	return deploymentSpec, nil
}

func makeServiceSpec(request requests.CreateFunctionRequest) *v1.Service {
	serviceSpec := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        request.Service,
			Annotations: map[string]string{"prometheus.io.scrape": "false"},
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeClusterIP,
			Selector: map[string]string{
				"faas_function": request.Service,
			},
			Ports: []v1.ServicePort{
				{
					Protocol: v1.ProtocolTCP,
					Port:     watchdogPort,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: int32(watchdogPort),
					},
				},
			},
		},
	}
	return serviceSpec
}

func buildEnvVars(request *requests.CreateFunctionRequest) []v1.EnvVar {
	envVars := []v1.EnvVar{}

	if len(request.EnvProcess) > 0 {
		envVars = append(envVars, v1.EnvVar{
			Name:  "fprocess",
			Value: request.EnvProcess,
		})
	}

	for k, v := range request.EnvVars {
		envVars = append(envVars, v1.EnvVar{
			Name:  k,
			Value: v,
		})
	}

	return envVars
}

func int32p(i int32) *int32 {
	return &i
}

func createSelector(constraints []string) map[string]string {
	selector := make(map[string]string)

	log.Println(constraints)
	if len(constraints) > 0 {
		for _, constraint := range constraints {
			parts := strings.Split(constraint, "=")

			if len(parts) == 2 {
				selector[parts[0]] = parts[1]
			}
		}
	}

	// log.Println("selector: ", selector)
	return selector
}

func createResources(request requests.CreateFunctionRequest) (*apiv1.ResourceRequirements, error) {
	resources := &apiv1.ResourceRequirements{
		Limits:   apiv1.ResourceList{},
		Requests: apiv1.ResourceList{},
	}

	// Set Memory limits
	if request.Limits != nil && len(request.Limits.Memory) > 0 {
		qty, err := resource.ParseQuantity(request.Limits.Memory)
		if err != nil {
			return resources, err
		}
		resources.Limits[apiv1.ResourceMemory] = qty
	}

	if request.Requests != nil && len(request.Requests.Memory) > 0 {
		qty, err := resource.ParseQuantity(request.Requests.Memory)
		if err != nil {
			return resources, err
		}
		resources.Requests[apiv1.ResourceMemory] = qty
	}

	// Set CPU limits
	if request.Limits != nil && len(request.Limits.CPU) > 0 {
		qty, err := resource.ParseQuantity(request.Limits.CPU)
		if err != nil {
			return resources, err
		}
		resources.Limits[apiv1.ResourceCPU] = qty
	}

	if request.Requests != nil && len(request.Requests.CPU) > 0 {
		qty, err := resource.ParseQuantity(request.Requests.CPU)
		if err != nil {
			return resources, err
		}
		resources.Requests[apiv1.ResourceCPU] = qty
	}

	return resources, nil
}

func getMinReplicaCount(labels map[string]string) *int32 {
	if value, exists := labels["com.openfaas.scale.min"]; exists {
		minReplicas, err := strconv.Atoi(value)
		if err == nil && minReplicas > 0 {
			return int32p(int32(minReplicas))
		}

		log.Println(err)
	}

	return nil
}
