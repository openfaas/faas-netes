package server

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/openfaas/faas/gateway/requests"
	glog "k8s.io/klog"
)

// makeProxy creates a proxy for HTTP web requests which can be routed to a function.
func makeProxy(functionNamespace string, timeout time.Duration) http.HandlerFunc {
	proxyClient := http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   timeout,
				KeepAlive: 1 * time.Second,
			}).DialContext,
			IdleConnTimeout:       120 * time.Millisecond,
			ExpectContinueTimeout: 1500 * time.Millisecond,
		},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			defer r.Body.Close()
		}

		switch r.Method {
		case http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodGet:

			vars := mux.Vars(r)
			service := vars["name"]

			defer func(when time.Time) {
				seconds := time.Since(when).Seconds()
				glog.V(2).Infof("%s took %f seconds", service, seconds)
			}(time.Now())

			forwardReq := requests.NewForwardRequest(r.Method, *r.URL)

			url := forwardReq.ToURL(fmt.Sprintf("%s.%s", service, functionNamespace), 8080)

			request, _ := http.NewRequest(r.Method, url, r.Body)

			copyHeaders(&request.Header, &r.Header)

			defer request.Body.Close()

			response, err := proxyClient.Do(request)

			if err != nil {
				glog.Errorf("%s error: %s", service, err.Error())
				writeHead(service, http.StatusInternalServerError, w)
				buf := bytes.NewBufferString("Can't reach service: " + service)
				w.Write(buf.Bytes())
				return
			}

			clientHeader := w.Header()
			copyHeaders(&clientHeader, &response.Header)

			writeHead(service, http.StatusOK, w)
			io.Copy(w, response.Body)
		}
	}
}

func writeHead(service string, code int, w http.ResponseWriter) {
	w.WriteHeader(code)
}

func copyHeaders(destination *http.Header, source *http.Header) {
	for k, v := range *source {
		vClone := make([]string, len(v))
		copy(vClone, v)
		(*destination)[k] = vClone
	}
}
