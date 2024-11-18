// License: OpenFaaS Community Edition (CE) EULA
// Copyright (c) 2017,2019-2024 OpenFaaS Author(s)

package k8s

import (
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

// isNotFound tests if the error is a kubernetes API error that indicates that the object
// was not found or does not exist
func IsNotFound(err error) bool {
	return k8serrors.IsNotFound(err) || k8serrors.IsGone(err)
}
