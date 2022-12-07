package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	faasv1 "github.com/openfaas/faas-netes/pkg/apis/openfaas/v1"
	clientset "github.com/openfaas/faas-netes/pkg/client/clientset/versioned"
	"github.com/openfaas/faas-netes/pkg/handlers"
	"github.com/openfaas/faas-provider/types"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

func makeApplyHandler(defaultNamespace string, client clientset.Interface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Body != nil {
			defer r.Body.Close()
		}

		body, _ := ioutil.ReadAll(r.Body)
		req := types.FunctionDeployment{}

		if err := json.Unmarshal(body, &req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		if err := handlers.ValidateDeployRequest(&req); err != nil {
			wrappedErr := fmt.Errorf("validation failed: %s", err.Error())
			http.Error(w, wrappedErr.Error(), http.StatusBadRequest)
			return
		}

		klog.Infof("Deployment request for: %s\n", req.Service)

		namespace := defaultNamespace
		if len(req.Namespace) > 0 {
			namespace = req.Namespace
		}

		if namespace != defaultNamespace {
			http.Error(w, fmt.Sprintf("valid namespaces are: %s", defaultNamespace), http.StatusBadRequest)
			return
		}

		opts := metav1.GetOptions{}
		got, err := client.OpenfaasV1().Functions(namespace).Get(r.Context(), req.Service, opts)
		miss := false
		if err != nil {
			if errors.IsNotFound(err) {
				miss = true
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
		}

		// Sometimes there was a non-nil "got" value when the miss was
		// true.
		if miss == false && got != nil {
			updated := got.DeepCopy()
			klog.Infof("Updating %s\n", updated.ObjectMeta.Name)

			updated.Spec = toFunctionSpec(req)

			if _, err = client.OpenfaasV1().Functions(namespace).
				Update(r.Context(), updated, metav1.UpdateOptions{}); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("Error updating function: %s", err.Error())))
				return
			}
		} else {

			newFunc := &faasv1.Function{
				ObjectMeta: metav1.ObjectMeta{
					Name:      req.Service,
					Namespace: namespace,
				},
				Spec: toFunctionSpec(req),
			}

			if _, err = client.OpenfaasV1().Functions(namespace).
				Create(r.Context(), newFunc, metav1.CreateOptions{}); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("Error creating function: %s", err.Error())))
				return
			}
		}

		w.WriteHeader(http.StatusAccepted)
	}
}

func toFunctionSpec(req types.FunctionDeployment) faasv1.FunctionSpec {
	spec := faasv1.FunctionSpec{
		Name:                   req.Service,
		Image:                  req.Image,
		Handler:                req.EnvProcess,
		Labels:                 req.Labels,
		Annotations:            req.Annotations,
		Constraints:            req.Constraints,
		Secrets:                req.Secrets,
		Limits:                 getResources(req.Limits),
		Requests:               getResources(req.Requests),
		ReadOnlyRootFilesystem: req.ReadOnlyRootFilesystem,
	}

	if req.EnvVars != nil {
		spec.Environment = &req.EnvVars
	}

	return spec
}

func getResources(limits *types.FunctionResources) *faasv1.FunctionResources {
	if limits == nil {
		return nil
	}
	return &faasv1.FunctionResources{
		CPU:    limits.CPU,
		Memory: limits.Memory,
	}
}

func int32p(i int32) *int32 {
	return &i
}
