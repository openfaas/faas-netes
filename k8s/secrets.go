// Copyright 2019 OpenFaaS Authors
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package k8s

import (
	"fmt"
	"log"
	"strings"

	types "github.com/openfaas/faas-provider/types"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	typedV1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	secretsMountPath = "/var/openfaas/secrets"
	secretLabel      = "app.kubernetes.io/managed-by"
	secretLabelValue = "openfaas"
)

// SecretsClient exposes the standardized CRUD behaviors for Kubernetes secrets.  These methods
// will ensure that the secrets are structured and labelled correctly for use by the OpenFaaS system.
type SecretsClient interface {
	// List returns a list of available function secrets.  Only the names are returned
	// to ensure we do not accidentally read or print the sensitive values during
	// read operations.
	List(namespace string) (names []string, err error)
	// Create adds a new secret, with the appropriate labels and structure to be
	// used as a function secret.
	Create(secret types.Secret) error
	// Replace updates the value of a function secret
	Replace(secret types.Secret) error
	// Delete removes a function secret
	Delete(name string, namespace string) error
	// GetSecrets queries Kubernetes for a list of secrets by name in the given k8s namespace.
	// This should only be used if you need access to the actual secret structure/value. Specifically,
	// inside the FunctionFactory.
	GetSecrets(namespace string, secretNames []string) (map[string]*apiv1.Secret, error)
}

// SecretsInterfacer exposes the SecretInterface getter for the k8s client.
// This is implemented by the CoreV1Interface() interface in the Kubernetes client.
// The SecretsClient only needs this one interface, but needs to be able to set the
// namespaces when the interface is instantiated, meaning, we need the Getter and not the
// SecretInterface itself.
type SecretInterfacer interface {
	// Secrets returns a SecretInterface scoped to the specified namespace
	Secrets(namespace string) typedV1.SecretInterface
}

type secretClient struct {
	kube SecretInterfacer
}

// NewSecretsClient constructs a new SecretsClient using the provided Kubernetes client.
func NewSecretsClient(kube kubernetes.Interface) SecretsClient {
	return &secretClient{
		kube: kube.CoreV1(),
	}
}

func (c secretClient) List(namespace string) (names []string, err error) {
	res, err := c.kube.Secrets(namespace).List(c.selector())
	if err != nil {
		log.Printf("failed to list secrets in %s: %v\n", namespace, err)
		return nil, err
	}

	names = make([]string, len(res.Items))
	for idx, item := range res.Items {
		// this is safe because size of names matches res.Items exactly
		names[idx] = item.Name
	}
	return names, nil
}

func (c secretClient) Create(secret types.Secret) error {
	err := c.validateSecret(secret)
	if err != nil {
		return err
	}

	req := &apiv1.Secret{
		Type: apiv1.SecretTypeOpaque,
		ObjectMeta: metav1.ObjectMeta{
			Name:      secret.Name,
			Namespace: secret.Namespace,
			Labels: map[string]string{
				secretLabel: secretLabelValue,
			},
		},
		StringData: map[string]string{
			secret.Name: secret.Value,
		},
	}

	_, err = c.kube.Secrets(secret.Namespace).Create(req)
	if err != nil {
		log.Printf("failed to create secret %s.%s: %v\n", secret.Name, secret.Namespace, err)
		return err
	}

	log.Printf("created secret %s.%s\n", secret.Name, secret.Namespace)

	return nil
}

func (c secretClient) Replace(secret types.Secret) error {
	err := c.validateSecret(secret)
	if err != nil {
		return err
	}

	kube := c.kube.Secrets(secret.Namespace)
	found, err := kube.Get(secret.Name, metav1.GetOptions{})
	if err != nil {
		log.Printf("can not retrieve secret for update %s.%s: %v\n", secret.Name, secret.Namespace, err)
		return err
	}

	found.StringData = map[string]string{
		secret.Name: secret.Value,
	}
	_, err = kube.Update(found)
	if err != nil {
		log.Printf("can not update secret %s.%s: %v\n", secret.Name, secret.Namespace, err)
		return err
	}

	return nil
}

func (c secretClient) Delete(namespace string, name string) error {
	err := c.kube.Secrets(namespace).Delete(name, &metav1.DeleteOptions{})
	if err != nil {
		log.Printf("can not delete %s.%s: %v\n", name, namespace, err)
	}
	return err
}

func (c secretClient) GetSecrets(namespace string, secretNames []string) (map[string]*apiv1.Secret, error) {
	kube := c.kube.Secrets(namespace)
	opts := metav1.GetOptions{}

	secrets := map[string]*apiv1.Secret{}
	for _, secretName := range secretNames {
		secret, err := kube.Get(secretName, opts)
		if err != nil {
			return nil, err
		}
		secrets[secretName] = secret
	}

	return secrets, nil
}

func (c secretClient) selector() metav1.ListOptions {
	return metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", secretLabel, secretLabelValue),
	}
}

func (c secretClient) validateSecret(secret types.Secret) error {
	if strings.TrimSpace(secret.Namespace) == "" {
		return errors.New("namespace may not be empty")
	}

	if strings.TrimSpace(secret.Name) == "" {
		return errors.New("name may not be empty")
	}

	return nil
}

// ConfigureSecrets will update the Deployment spec to include secrets that have been deployed
// in the kubernetes cluster.  For each requested secret, we inspect the type and add it to the
// deployment spec as appropriate: secrets with type `SecretTypeDockercfg/SecretTypeDockerjson`
// are added as ImagePullSecrets all other secrets are mounted as files in the deployments containers.
func (f *FunctionFactory) ConfigureSecrets(request types.FunctionDeployment, deployment *appsv1.Deployment, existingSecrets map[string]*apiv1.Secret) error {
	// Add / reference pre-existing secrets within Kubernetes
	secretVolumeProjections := []apiv1.VolumeProjection{}

	for _, secretName := range request.Secrets {
		deployedSecret, ok := existingSecrets[secretName]
		if !ok {
			return fmt.Errorf("Required secret '%s' was not found in the cluster", secretName)
		}

		switch deployedSecret.Type {

		case apiv1.SecretTypeDockercfg,
			apiv1.SecretTypeDockerConfigJson:

			deployment.Spec.Template.Spec.ImagePullSecrets = append(
				deployment.Spec.Template.Spec.ImagePullSecrets,
				apiv1.LocalObjectReference{
					Name: secretName,
				},
			)
		default:

			projectedPaths := []apiv1.KeyToPath{}
			for secretKey := range deployedSecret.Data {
				projectedPaths = append(projectedPaths, apiv1.KeyToPath{Key: secretKey, Path: secretKey})
			}

			projection := &apiv1.SecretProjection{Items: projectedPaths}
			projection.Name = secretName
			secretProjection := apiv1.VolumeProjection{
				Secret: projection,
			}
			secretVolumeProjections = append(secretVolumeProjections, secretProjection)
		}
	}

	volumeName := fmt.Sprintf("%s-projected-secrets", request.Service)
	projectedSecrets := apiv1.Volume{
		Name: volumeName,
		VolumeSource: apiv1.VolumeSource{
			Projected: &apiv1.ProjectedVolumeSource{
				Sources: secretVolumeProjections,
			},
		},
	}

	// remove the existing secrets volume, if we can find it. The update volume will be
	// added below
	existingVolumes := removeVolume(volumeName, deployment.Spec.Template.Spec.Volumes)
	deployment.Spec.Template.Spec.Volumes = existingVolumes
	if len(secretVolumeProjections) > 0 {
		deployment.Spec.Template.Spec.Volumes = append(existingVolumes, projectedSecrets)
	}

	// add mount secret as a file
	updatedContainers := []apiv1.Container{}
	for _, container := range deployment.Spec.Template.Spec.Containers {
		mount := apiv1.VolumeMount{
			Name:      volumeName,
			ReadOnly:  true,
			MountPath: secretsMountPath,
		}

		// remove the existing secrets volume mount, if we can find it. We update it later.
		container.VolumeMounts = removeVolumeMount(volumeName, container.VolumeMounts)
		if len(secretVolumeProjections) > 0 {
			container.VolumeMounts = append(container.VolumeMounts, mount)
		}

		updatedContainers = append(updatedContainers, container)
	}

	deployment.Spec.Template.Spec.Containers = updatedContainers

	return nil
}
