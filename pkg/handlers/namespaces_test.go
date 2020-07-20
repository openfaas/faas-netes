// Copyright 2019 OpenFaaS Author(s)
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import "testing"

func Test_findNamespace_Found(t *testing.T) {
	got := findNamespace("fn", []string{"fn", "openfaas-fn"})
	want := true

	if got != want {
		t.Errorf("findNamespace - want: %v, got %v", want, got)
	}
}

func Test_findNamespace_NotFound(t *testing.T) {
	got := findNamespace("fn", []string{"openfaas-fn"})
	want := false

	if got != want {
		t.Errorf("findNamespace - want: %v, got %v", want, got)
	}
}
