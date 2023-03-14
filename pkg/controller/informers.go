package controller

import (
	"context"
	"fmt"

	"github.com/openfaas/faas-netes/pkg/handlers"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	v1apps "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

func RegisterEventHandlers(deploymentInformer v1apps.DeploymentInformer, kubeClient *kubernetes.Clientset, namespace string) {
	deploymentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			deployment, ok := obj.(*appsv1.Deployment)
			if !ok || deployment == nil {
				return
			}
			if err := applyValidation(deployment, kubeClient); err != nil {
				klog.Info(err)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			deployment, ok := newObj.(*appsv1.Deployment)
			if !ok || deployment == nil {
				return
			}
			if err := applyValidation(deployment, kubeClient); err != nil {
				klog.Info(err)
			}
		},
	})

	list, err := deploymentInformer.Lister().Deployments(namespace).List(labels.Everything())
	if err != nil {
		klog.Info(err)
		return
	}

	for _, deployment := range list {
		if err := applyValidation(deployment, kubeClient); err != nil {
			klog.Info(err)
		}
	}
}

func applyValidation(deployment *appsv1.Deployment, kubeClient *kubernetes.Clientset) error {
	if deployment.Spec.Replicas == nil {
		return nil
	}

	if _, ok := deployment.Spec.Template.Labels["faas_function"]; !ok {
		return nil
	}

	current := *deployment.Spec.Replicas
	var target int
	if current == 0 {
		target = 1
	} else if current > handlers.MaxReplicas {
		target = handlers.MaxReplicas
	} else {
		return nil
	}
	clone := deployment.DeepCopy()

	value := int32(target)
	clone.Spec.Replicas = &value

	if _, err := kubeClient.AppsV1().Deployments(deployment.Namespace).
		Update(context.Background(), clone, metav1.UpdateOptions{}); err != nil {
		if errors.IsConflict(err) {
			return nil
		}
		return fmt.Errorf("error scaling %s to %d replicas: %w", deployment.Name, value, err)
	}

	return nil
}
