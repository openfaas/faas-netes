package handlers

import (
	"log"
	"strconv"
)

const (
	// initialReplicasCount how many replicas to start of creating for a function, this is
	// also used as the default return value for getMinReplicaCount
	initialReplicasCount = 1

	// OFFunctionNameLabel is the label key used by OpenFaaS to store the function name
	// on the resources managed by OpenFaaS for that function.  This key is also used to
	// denote that a resource is a "Function"
	OFFunctionNameLabel = "faas_function"
	// OFFunctionMinReplicaCount is a label that user's can set and will be passed to Kubernetes
	// as the Deployment replicas value.
	OFFunctionMinReplicaCount = "com.openfaas.scale.min"
)

// parseLabels will copy the user request labels and ensure that any required internal labels
// are set appropriately.
func parseLabels(functionName string, requestLables *map[string]string) map[string]string {
	labels := map[string]string{}
	if requestLables != nil {
		for k, v := range *requestLables {
			labels[k] = v
		}
	}

	labels[OFFunctionNameLabel] = functionName

	return labels
}

// getMinReplicaCount extracts the functions minimum replica count from the user's
// request labels. If the value is not found, this will return the default value, 1.
func getMinReplicaCount(labels *map[string]string) *int32 {
	if labels == nil {
		return int32p(initialReplicasCount)
	}

	l := *labels
	if value, exists := l[OFFunctionMinReplicaCount]; exists {
		minReplicas, err := strconv.Atoi(value)
		if err == nil && minReplicas > 0 {
			return int32p(int32(minReplicas))
		}

		log.Println(err)
	}

	return int32p(initialReplicasCount)
}
