// Copyright 2019 OpenFaaS Author(s)
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package k8s

import (
	"context"
	"log"
	"strings"

	"github.com/openfaas/faas-provider/logs"
	"k8s.io/client-go/kubernetes"
)

// LogRequestor implements the Requestor interface for k8s
type LogRequestor struct {
	client            kubernetes.Interface
	functionNamespace string
}

// NewLogRequestor returns a new logs.Requestor that uses kail to select and follow pod logs
func NewLogRequestor(client kubernetes.Interface, functionNamespace string) *LogRequestor {
	return &LogRequestor{
		client:            client,
		functionNamespace: functionNamespace,
	}
}

// Query implements the actual Swarm logs request logic for the Requestor interface
// This implementation ignores the r.Limit value because the OF-Provider already handles server side
// line limits.
func (l LogRequestor) Query(ctx context.Context, r logs.Request) (<-chan logs.Message, error) {
	ns := l.functionNamespace

	if len(r.Namespace) > 0 && strings.ToLower(r.Namespace) != "kube-system" {
		ns = r.Namespace
	}

	logStream, err := GetLogs(ctx, l.client, r.Name, ns, int64(r.Tail), r.Since, r.Follow)
	if err != nil {
		log.Printf("LogRequestor: get logs failed: %s\n", err)
		return nil, err
	}

	msgStream := make(chan logs.Message, LogBufferSize)
	go func() {
		defer close(msgStream)
		// here we depend on the fact that logStream will close when the context is cancelled,
		// this ensures that the go routine will resolve
		for msg := range logStream {
			msgStream <- logs.Message{
				Timestamp: msg.Timestamp,
				Text:      msg.Text,
				Name:      msg.FunctionName,
				Instance:  msg.PodName,
				Namespace: msg.Namespace,
			}
		}
	}()

	return msgStream, nil
}
