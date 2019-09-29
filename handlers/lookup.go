// Copyright 2019 OpenFaaS Author(s)
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"fmt"
	"net/http"
	"net/url"
)

type FunctionLookup struct {
}

func (l *FunctionLookup) Resolve(name string) (url.URL, error) {
	namespace := "openfaas-fn"

	urlRes, errRes := url.Parse(fmt.Sprintf("http://%s.%s:8080", name, namespace))

	return *urlRes, errRes
}

func NewFunctionLookup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		w.WriteHeader(http.StatusOK)

	}
}
