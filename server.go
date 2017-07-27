package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/alexellis/faas-netes/handlers"
	"github.com/gorilla/mux"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	r := mux.NewRouter()

	r.HandleFunc("/system/functions", handlers.MakeFunctionReader(clientset)).Methods("GET")
	r.HandleFunc("/system/functions", handlers.MakeDeployHandler(clientset)).Methods("POST")

	r.HandleFunc("/system/function/{name:[-a-zA-Z_0-9]+}", handlers.MakeReplicaReader(clientset)).Methods("GET")
	r.HandleFunc("/system/scale-function/{name:[-a-zA-Z_0-9]+}", handlers.MakeReplicaUpdater(clientset)).Methods("POST")

	functionProxy := handlers.MakeProxy()
	r.HandleFunc("/function/{name:[-a-zA-Z_0-9]+}", functionProxy)
	r.HandleFunc("/function/{name:[-a-zA-Z_0-9]+}/", functionProxy)

	readTimeout := 8 * time.Second
	writeTimeout := 8 * time.Second
	tcpPort := 8080

	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", tcpPort),
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: http.DefaultMaxHeaderBytes, // 1MB - can be overridden by setting Server.MaxHeaderBytes.
		Handler:        r,
	}

	log.Fatal(s.ListenAndServe())
}
