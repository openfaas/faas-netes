package server

import (
	"encoding/json"
	"net/http"

	"github.com/openfaas/faas-netes/version"
	"github.com/openfaas/faas-provider/types"
)

// makeInfoHandler provides the system/info endpoint
func makeInfoHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			defer r.Body.Close()
		}

		sha, release := version.GetReleaseInfo()
		info := types.InfoResponse{
			Orchestration: "kubernetes",
			Provider:      "openfaas-operator",
			Version: types.ProviderVersion{
				SHA:     sha,
				Release: release,
			},
		}

		infoBytes, err := json.Marshal(info)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(infoBytes)
	}

}
