package server

import (
	"context"
	"encoding/json"
	pk8s "github.com/openfaas/faas-netes/pkg/k8s"
	"io"
	"io/ioutil"
	"net/http"

	clientset "github.com/openfaas/faas-netes/pkg/client/clientset/versioned"
	"github.com/openfaas/faas/gateway/requests"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	glog "k8s.io/klog"
)

func makeDeleteHandler(namespace string, client clientset.Interface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Body != nil {
			defer r.Body.Close()
		}

		body, _ := ioutil.ReadAll(r.Body)
		request := requests.DeleteFunctionRequest{}
		err := json.Unmarshal(body, &request)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		if len(request.FunctionName) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "No function provided")
			return
		}

		functionName, namespace := pk8s.GetNamespace(request.FunctionName, namespace)

		err = client.OpenfaasV1().Functions(namespace).
			Delete(context.TODO(), functionName, metav1.DeleteOptions{})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			glog.Errorf("Function %s in %s delete error: %v", functionName, namespace, err)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	}
}
