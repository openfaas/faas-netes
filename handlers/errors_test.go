// Copyright 2019 OpenFaaS Author(s)
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"errors"
	"net/http"
	"testing"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func Test_DefaultErrorReason(t *testing.T) {
	err := errors.New("A default error mapping")
	status, reason := ProcessErrorReasons(err)
	if status != http.StatusInternalServerError {
		t.Errorf("Unexpected default status code: %d", status)
	}
	if reason != metav1.StatusReasonInternalError {
		t.Errorf("Unexpected default reason: %s", reason)
	}
}

func Test_ConflictErrorReason(t *testing.T) {
	gv := schema.GroupResource{Group: "testing", Resource: "testing"}
	statusError := k8serrors.NewConflict(gv, "conflict_test", errors.New("A test error for confict"))

	status, reason := ProcessErrorReasons(statusError)
	if status != http.StatusConflict {
		t.Errorf("Unexpected default status code: %d", status)
	}
	if !k8serrors.IsConflict(statusError) {
		t.Errorf("Unexpected default reason: %s", reason)
	}
}
