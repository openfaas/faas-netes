// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"math/rand"

	"github.com/openfaas-incubator/openfaas-operator/pkg/signals"
	kubeinformers "k8s.io/client-go/informers"

	"github.com/gorilla/mux"
	"github.com/openfaas/faas/gateway/requests"
	"k8s.io/client-go/kubernetes"
)

// watchdogPort for the OpenFaaS function watchdog
const watchdogPort = 8080

func newHTTPClient(timeout time.Duration, maxIdleConns, maxIdleConnsPerHost int) *http.Client {

	client := http.DefaultClient
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	// These overrides for the default client enable re-use of connections and prevent
	// CoreDNS from rate limiting the gateway under high traffic
	//
	// See also two similar projects where this value was updated:
	// https://github.com/prometheus/prometheus/pull/3592
	// https://github.com/minio/minio/pull/5860

	// Taken from http.DefaultTransport in Go 1.11
	client.Transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   timeout,
			KeepAlive: timeout,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          maxIdleConns,
		MaxIdleConnsPerHost:   maxIdleConnsPerHost,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return client
}

// MakeProxy creates a proxy for HTTP web requests which can be routed to a function.
func MakeProxy(functionNamespace string, timeout time.Duration, clientset *kubernetes.Clientset) http.HandlerFunc {

	defaultResync := time.Second * 5
	kubeInformerOpt := kubeinformers.WithNamespace(functionNamespace)
	kubeInformerFactory := kubeinformers.NewSharedInformerFactoryWithOptions(clientset, defaultResync, kubeInformerOpt)

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	endpointsInformer := kubeInformerFactory.Core().V1().Endpoints()
	go kubeInformerFactory.Start(stopCh)
	lister := endpointsInformer.Lister()

	client := newHTTPClient(timeout, 1024, 1024)
	nsEndpointLister := lister.Endpoints(functionNamespace)

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

			svc, err := nsEndpointLister.Get(service)
			if err != nil {
				log.Printf("error listing %s %s\n", service, err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			all := len(svc.Subsets[0].Addresses)
			target := rand.Intn(all)

			serviceIP := svc.Subsets[0].Addresses[target].IP

			forwardReq := requests.NewForwardRequest(r.Method, *r.URL)

			url := forwardReq.ToURL(fmt.Sprintf("%s", serviceIP), watchdogPort)

			log.Printf("Invoke [%d/%d] via: %s\n", target+1, all, url)

			request, _ := http.NewRequest(r.Method, url, r.Body)

			copyHeaders(&request.Header, &r.Header)

			defer request.Body.Close()

			response, err := client.Do(request)

			if err != nil {
				log.Println(err.Error())
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
