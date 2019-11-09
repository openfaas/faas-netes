// Copyright 2019 OpenFaaS Authors
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package k8s

import (
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

// isNotFound tests if the error is a kubernetes API error that indicates that the object
// was not found or does not exist
func IsNotFound(err error) bool {
	return k8serrors.IsNotFound(err) || k8serrors.IsGone(err)
}
