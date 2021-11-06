package k8s

import (
	"strings"

	"k8s.io/client-go/kubernetes"
)

// Capabilities is a map of kuberenetes api resources
type Capabilities map[string]bool

// Has returns true if the api resource is supported
func (c Capabilities) Has(wanted string) bool {
	return c[wanted]
}

// String implements the Stringer interface
func (c Capabilities) String() string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	return strings.Join(keys, ", ")
}

// getPreferredAvailableAPIs queries the cluster for the preferred resources information and returns a Capabilities
// instance containing those api groups that support the specified kind.
//
// kind should be the title case singular name of the kind. For example, "Ingress" is the kind for a resource "ingress".
func GetPreferredAvailableAPIs(client kubernetes.Interface, kind string) (Capabilities, error) {
	discoveryclient := client.Discovery()
	lists, err := discoveryclient.ServerPreferredResources()
	if err != nil {
		return nil, err
	}

	caps := Capabilities{}
	for _, list := range lists {
		if len(list.APIResources) == 0 {
			continue
		}
		for _, resource := range list.APIResources {
			if len(resource.Verbs) == 0 {
				continue
			}
			if resource.Kind == kind {
				caps[list.GroupVersion] = true
			}
		}
	}

	return caps, nil
}

// ProfilesEnabled returns if the Profile CRD is installed in the cluster.
func ProfilesEnabled(client kubernetes.Interface) (bool, error) {
	capabilities, err := GetPreferredAvailableAPIs(client, "Profile")

	return capabilities.Has("openfaas.com/v1"), err
}

// FunctionEnabled returns if the Function CRD is installed in the cluster.
func FunctionEnabled(client kubernetes.Interface) (bool, error) {
	capabilities, err := GetPreferredAvailableAPIs(client, "Function")

	return capabilities.Has("openfaas.com/v1"), err
}
