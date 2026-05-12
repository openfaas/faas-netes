package k8s

import (
	"io/ioutil"
	"os"
	"strings"
)

// CurrentNamespace attempts to return the current namespace from the environment
// or from the service account file. If it cannot find the namespace, it returns
// an empty string. This will be empty when the not running in-cluster.
//
// This implementation is based on the clientcmd.inClusterClientConfig.Namespace method.
// This is not exported and not accessible via other methods, so we have to copy it.
func CurrentNamespace() (namespace string, found bool) {
	if ns := os.Getenv("POD_NAMESPACE"); ns != "" {
		return ns, true
	}

	if data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
			return ns, true
		}
	}

	return "", false
}
