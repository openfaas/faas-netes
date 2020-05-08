package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/openfaas/faas-netes/k8s"
	faasv1 "github.com/openfaas/faas-netes/pkg/apis/openfaas/v1"
	v1 "github.com/openfaas/faas-netes/pkg/apis/openfaas/v1"
	clientset "github.com/openfaas/faas-netes/pkg/client/clientset/versioned"
	"github.com/openfaas/faas-provider/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	glog "k8s.io/klog"
)

func makeApplyHandler(namespace string, client clientset.Interface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Body != nil {
			defer r.Body.Close()
		}

		body, _ := ioutil.ReadAll(r.Body)
		req := types.FunctionDeployment{}
		err := json.Unmarshal(body, &req)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, writeErr := w.Write([]byte(err.Error()))
			if writeErr != nil {
				glog.Errorf("response writing error: %s", writeErr)
			}
			return
		}

		var newFunc *faasv1.Function
		spec := faasv1.FunctionSpec{
			Name:                   req.Service,
			Image:                  req.Image,
			Handler:                req.EnvProcess,
			Labels:                 req.Labels,
			Annotations:            req.Annotations,
			Environment:            &req.EnvVars,
			Constraints:            req.Constraints,
			Secrets:                req.Secrets,
			Limits:                 getResources(req.Limits),
			Requests:               getResources(req.Requests),
			ReadOnlyRootFilesystem: req.ReadOnlyRootFilesystem,
		}

		var method func(*v1.Function) (*v1.Function, error)
		method = client.OpenfaasV1().Functions(namespace).Update

		found, err := client.OpenfaasV1().Functions(namespace).Get(req.Service, metav1.GetOptions{})
		if k8s.IsNotFound(err) {
			method = client.OpenfaasV1().Functions(namespace).Create
			newFunc = &faasv1.Function{
				ObjectMeta: metav1.ObjectMeta{
					Name:      req.Service,
					Namespace: namespace,
				},
				Spec: spec,
			}
		} else {
			newFunc = found
			newFunc.Spec = spec
		}

		_, err = method(newFunc)
		if err != nil {
			glog.Errorf("Function %s update error: %v", req.Service, err)
			w.WriteHeader(http.StatusInternalServerError)
			_, writeErr := w.Write([]byte(err.Error()))
			if writeErr != nil {
				glog.Errorf("response writing error: %s", writeErr)
			}
			return
		}

		glog.Infof("Function %s updated", req.Service)
		w.WriteHeader(http.StatusAccepted)
	}
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
