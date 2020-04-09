package server

import (
	"net/http"
)

// makeHealthHandler provides the healthz endpoint
func makeHealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			defer r.Body.Close()
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}
