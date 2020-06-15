// +build tools
package tools

import (
	// Embedded as a tool
	// https://github.com/go-modules-by-example/index/blob/master/010_tools/README.md#tools-as-dependencies
	_ "sigs.k8s.io/controller-tools/cmd/controller-gen"
)