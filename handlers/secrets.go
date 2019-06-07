package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	typedV1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/openfaas/faas/gateway/requests"
	appsv1 "k8s.io/api/apps/v1beta2"
	apiv1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	secretsMountPath = "/var/openfaas/secrets"
	secretLabel      = "app.kubernetes.io/managed-by"
	secretLabelValue = "openfaas"
)

// MakeSecretHandler makes a handler for Create/List/Delete/Update of
//secrets in the Kubernetes API
func MakeSecretHandler(namespace string, kube kubernetes.Interface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			defer r.Body.Close()
		}

		client := kube.CoreV1().Secrets(namespace)

		switch r.Method {
		case http.MethodGet:
			getSecretsHandler(client, w, r)
		case http.MethodPost:
			createSecretHandler(client, namespace, w, r)
		case http.MethodPut:
			replaceSecretHandler(client, namespace, w, r)
		case http.MethodDelete:
			deleteSecretHandler(client, namespace, w, r)
		default:
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
}

func getSecretsHandler(kube typedV1.SecretInterface, w http.ResponseWriter, r *http.Request) {
	selector := fmt.Sprintf("%s=%s", secretLabel, secretLabelValue)
	res, err := kube.List(metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		status, reason := processErrorReasons(err)
		log.Printf("Secret list error reason: %s, %v\n", reason, err)
		w.WriteHeader(status)
		return
	}

	secrets := []requests.Secret{}
	for _, item := range res.Items {
		secret := requests.Secret{
			Name: item.Name,
		}
		secrets = append(secrets, secret)
	}
	secretsBytes, err := json.Marshal(secrets)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Secrets json marshal error: %v\n", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(secretsBytes)
}

func createSecretHandler(kube typedV1.SecretInterface, namespace string, w http.ResponseWriter, r *http.Request) {
	secret, err := parseSecret(namespace, r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Secret unmarshal error: %v\n", err)
		return
	}
	_, err = kube.Create(secret)
	if err != nil {
		status, reason := processErrorReasons(err)
		log.Printf("Secret create error reason: %s, %v\n", reason, err)
		w.WriteHeader(status)
		return
	}
	log.Printf("Secret %s create\n", secret.GetName())
	w.WriteHeader(http.StatusAccepted)
}

func replaceSecretHandler(kube typedV1.SecretInterface, namespace string, w http.ResponseWriter, r *http.Request) {
	newSecret, err := parseSecret(namespace, r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Secret unmarshal error: %v\n", err)
		return
	}
	secret, err := kube.Get(newSecret.GetName(), metav1.GetOptions{})
	if isNotFound(err) {
		w.WriteHeader(http.StatusNotFound)
		log.Printf("Secret update error: %s not found\n", newSecret.GetName())
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Secret query error: %v\n", err)
		return
	}
	secret.StringData = newSecret.StringData
	_, err = kube.Update(secret)
	if err != nil {
		status, reason := processErrorReasons(err)
		log.Printf("Secret update error reason: %s, %v\n", reason, err)
		w.WriteHeader(status)
		return
	}
	log.Printf("Secret %s updated", secret.GetName())
	w.WriteHeader(http.StatusAccepted)
}

func deleteSecretHandler(kube typedV1.SecretInterface, namespace string, w http.ResponseWriter, r *http.Request) {
	secret, err := parseSecret(namespace, r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Secret unmarshal error: %v\n", err)
		return
	}
	opts := &metav1.DeleteOptions{}
	err = kube.Delete(secret.GetName(), opts)
	if isNotFound(err) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		status, reason := processErrorReasons(err)
		log.Printf("Secret delete error reason: %s, %v\n", reason, err)
		w.WriteHeader(status)
		return
	}
	log.Printf("Secret %s deleted\n", secret.GetName())
	w.WriteHeader(http.StatusAccepted)
}

func parseSecret(namespace string, r io.Reader) (*apiv1.Secret, error) {
	body, _ := ioutil.ReadAll(r)
	req := requests.Secret{}
	err := json.Unmarshal(body, &req)
	if err != nil {
		return nil, err
	}
	secret := &apiv1.Secret{
		Type: apiv1.SecretTypeOpaque,
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: namespace,
			Labels: map[string]string{
				secretLabel: secretLabelValue,
			},
		},
		StringData: map[string]string{
			req.Name: req.Value,
		},
	}

	return secret, nil
}

// getSecrets queries Kubernetes for a list of secrets by name in the given k8s namespace.
func getSecrets(clientset kubernetes.Interface, namespace string, secretNames []string) (map[string]*apiv1.Secret, error) {
	secrets := map[string]*apiv1.Secret{}

	for _, secretName := range secretNames {
		secret, err := clientset.Core().Secrets(namespace).Get(secretName, metav1.GetOptions{})
		if err != nil {
			return secrets, err
		}
		secrets[secretName] = secret
	}

	return secrets, nil
}

// UpdateSecrets will update the Deployment spec to include secrets that have beenb deployed
// in the kubernetes cluster.  For each requested secret, we inspect the type and add it to the
// deployment spec as appropriat: secrets with type `SecretTypeDockercfg/SecretTypeDockerjson`
// are added as ImagePullSecrets all other secrets are mounted as files in the deployments containers.
func UpdateSecrets(request requests.CreateFunctionRequest, deployment *appsv1.Deployment, existingSecrets map[string]*apiv1.Secret) error {
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

			break

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

			break
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

// removeVolume returns a Volume slice with any volumes matching volumeName removed.
// Uses the filter without allocation technique
// https://github.com/golang/go/wiki/SliceTricks#filtering-without-allocating
func removeVolume(volumeName string, volumes []apiv1.Volume) []apiv1.Volume {
	if volumes == nil {
		return []apiv1.Volume{}
	}

	newVolumes := volumes[:0]
	for _, v := range volumes {
		if v.Name != volumeName {
			newVolumes = append(newVolumes, v)
		}
	}

	return newVolumes
}

// removeVolumeMount returns a VolumeMount slice with any mounts matching volumeName removed
// Uses the filter without allocation technique
// https://github.com/golang/go/wiki/SliceTricks#filtering-without-allocating
func removeVolumeMount(volumeName string, mounts []apiv1.VolumeMount) []apiv1.VolumeMount {
	if mounts == nil {
		return []apiv1.VolumeMount{}
	}

	newMounts := mounts[:0]
	for _, v := range mounts {
		if v.Name != volumeName {
			newMounts = append(newMounts, v)
		}
	}

	return newMounts
}

// isNotFound tests if the error is a kubernetes API error that indicates that the object
// was not found or does not exist
func isNotFound(err error) bool {
	return k8serrors.IsNotFound(err) || k8serrors.IsGone(err)
}

// processErrorReasons maps k8serrors.ReasonForError to http status codes
func processErrorReasons(err error) (int, metav1.StatusReason) {
	switch {
	case k8serrors.ReasonForError(err) == metav1.StatusReasonConflict:
		return http.StatusConflict, metav1.StatusReasonConflict
	case k8serrors.ReasonForError(err) == metav1.StatusReasonInvalid:
		return http.StatusUnprocessableEntity, metav1.StatusReasonInvalid
	case k8serrors.ReasonForError(err) == metav1.StatusReasonBadRequest:
		return http.StatusBadRequest, metav1.StatusReasonBadRequest
	case k8serrors.ReasonForError(err) == metav1.StatusReasonForbidden:
		return http.StatusForbidden, metav1.StatusReasonForbidden
	case k8serrors.ReasonForError(err) == metav1.StatusReasonTimeout:
		return http.StatusRequestTimeout, metav1.StatusReasonTimeout
	default:
		return http.StatusInternalServerError, metav1.StatusReasonInternalError
	}
}
