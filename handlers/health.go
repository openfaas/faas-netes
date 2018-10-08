package handlers

import (
	"log"
	"net/http"

	"k8s.io/client-go/kubernetes"
)

// MakeHealthHandler returns 200/OK when healthy
func MakeHealthHandler(clientset *kubernetes.Clientset) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		_, err := clientset.Discovery().ServerVersion()
		if err != nil {
			log.Printf("client communicating with apiserver: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusOK)
	}
}
