// Copyright 2019 OpenFaaS Author(s)
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import "net/http"

// MakeHealthHandler returns 200/OK when healthy
func MakeHealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		w.WriteHeader(http.StatusOK)
	}
}
