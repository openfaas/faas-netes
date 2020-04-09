package server

import (
	"encoding/json"
	"net/http"
)

func makeListNamespaceHandler(defaultNamespace string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		defer r.Body.Close()

		res, _ := json.Marshal([]string{defaultNamespace})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(res)
	}
}
