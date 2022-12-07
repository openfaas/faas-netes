package server

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	clientset "github.com/openfaas/faas-netes/pkg/client/clientset/versioned"
	"github.com/openfaas/faas-provider/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	glog "k8s.io/klog"
)

func makeDeleteHandler(defaultNamespace string, client clientset.Interface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		q := r.URL.Query()
		namespace := q.Get("namespace")

		lookupNamespace := defaultNamespace
		if len(namespace) > 0 {
			lookupNamespace = namespace
		}

		if lookupNamespace != defaultNamespace {
			http.Error(w, fmt.Sprintf("valid namespaces are: %s", defaultNamespace), http.StatusBadRequest)
			return
		}

		if namespace == "kube-system" {
			http.Error(w, "unable to list within the kube-system namespace", http.StatusUnauthorized)
			return
		}

		if r.Body != nil {
			defer r.Body.Close()
		}

		body, _ := ioutil.ReadAll(r.Body)
		request := types.DeleteFunctionRequest{}
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

		err = client.OpenfaasV1().Functions(lookupNamespace).
			Delete(r.Context(), request.FunctionName, metav1.DeleteOptions{})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			glog.Errorf("Function %s delete error: %v", request.FunctionName, err)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	}
}
