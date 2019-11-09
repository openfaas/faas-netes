// Copyright 2019 OpenFaaS Author(s)
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"net/http"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ProcessErrorReasons maps k8serrors.ReasonForError to http status codes
func ProcessErrorReasons(err error) (int, metav1.StatusReason) {
	reason := k8serrors.ReasonForError(err)
	switch reason {
	case metav1.StatusReasonGone, metav1.StatusReasonNotFound:
		return http.StatusNotFound, reason
	case metav1.StatusReasonAlreadyExists:
		return http.StatusConflict, reason
	case metav1.StatusReasonConflict:
		return http.StatusConflict, reason
	case metav1.StatusReasonInvalid:
		return http.StatusUnprocessableEntity, reason
	case metav1.StatusReasonBadRequest:
		return http.StatusBadRequest, reason
	case metav1.StatusReasonForbidden:
		return http.StatusForbidden, reason
	case metav1.StatusReasonTimeout:
		return http.StatusRequestTimeout, reason
	default:
		return http.StatusInternalServerError, metav1.StatusReasonInternalError
	}
}
