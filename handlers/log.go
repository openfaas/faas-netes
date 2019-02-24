package handlers

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/boz/kcache/nsname"
	"github.com/pkg/errors"

	"github.com/boz/kail"
	"github.com/openfaas/faas-provider/logs"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	// number of log messages that may be buffered
	logBufferSize = 500 * 2 // double the event buffer size in kail
	// defaultLogSince is the fallback log stream history
	defaultLogSince = 5 * time.Minute
)

var (
	acceptAllFilter = kail.NewContainerFilter(nil)
)

// LogRequestor implements the Requestor interface for k8s
type LogRequestor struct {
	client            kubernetes.Interface
	rc                *rest.Config
	functionNamespace string
}

// NewLogRequestor returns a new logs.Requestor that uses kail to select and follow pod logs
func NewLogRequestor(client kubernetes.Interface, rc *rest.Config, functionNamespace string) *LogRequestor {
	return &LogRequestor{
		client:            client,
		rc:                rc,
		functionNamespace: functionNamespace,
	}
}

// Query implements the actual Swarm logs request logic for the Requestor interface
// This implementation ignores the r.Limit value because the OF-Provider already handles server side
// line limits.
//
// TODO:
// * implement Follow=false logic by stopping when timestamp hits time.Now?
func (l LogRequestor) Query(ctx context.Context, r logs.Request) (<-chan logs.Message, error) {
	logSourceBuilder := kail.NewDSBuilder()
	logSourceBuilder = logSourceBuilder.WithNamespace(l.functionNamespace)
	logSourceBuilder = logSourceBuilder.WithDeployment(nsname.NSName{Namespace: l.functionNamespace, Name: r.Name})

	logSource, err := createLogSource(ctx, l.client, logSourceBuilder)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create log source")
	}

	controller, err := kail.NewController(ctx, l.client, l.rc, logSource.Pods(), acceptAllFilter, parseSince(r))
	if err != nil {
		return nil, errors.Wrap(err, "unable start log controller")
	}

	msgStream := make(chan logs.Message, logBufferSize)
	go startLogStream(ctx, r.Name, r.Follow, controller, msgStream)

	return msgStream, nil
}

func createLogSource(ctx context.Context, cs kubernetes.Interface, dsb kail.DSBuilder) (kail.DS, error) {
	ds, err := dsb.Create(ctx, cs)
	if err != nil {
		return nil, err
	}

	select {
	case <-ds.Ready():
	case <-ds.Done():
		return nil, errors.New("Unable to initialize data source")
	}

	return ds, nil
}

// parseSince returns the time.Duration of the requested Since value _or_ 5 minutes
func parseSince(r logs.Request) time.Duration {
	if r.Since != nil {
		return time.Since(*r.Since)
	}

	return defaultLogSince
}

// startLogStream will start parsing the log events from the kail controller, parse them into log
//  messages and send them on the msgStream.  The msgStream channel will be closed when the context
// is cancelled or when the log stream finishes.
func startLogStream(ctx context.Context, name string, follow bool, controller kail.Controller, msgStream chan<- logs.Message) {
	defer close(msgStream)
	defer controller.Close()

	msgCount := 0

	deadline, ok := ctx.Deadline()
	if ok {
		log.Printf("LogRequestor: stream logs until %s\n", deadline.Format(time.RFC3339))
	} else {
		log.Println("LogRequestor: no stream deadline")
	}

	now := time.Now()

	for {
		// double check if the stream was cancelled
		if ctx.Err() != nil {
			log.Println("LogRequestor: log stream context cancelled or done")
			return
		}

		deadline, ok := ctx.Deadline()
		if ok {
			log.Printf("LogRequestor: stream logs until %s\n", deadline.Format(time.RFC3339))
		}

		timer := time.NewTimer(time.Second)

		select {
		case <-timer.C:
			if !follow {
				log.Println("LogRequestor: log stream follow timeout reached")
				return
			}
		case ev := <-controller.Events():
			source := ev.Source()
			msg, ts := extractTimestampAndMsg(string(bytes.Trim(ev.Log(), "\x00")))

			if !follow && ts.After(now) {
				log.Println("LogRequestor: log stream reached now")
				return
			}

			msgStream <- logs.Message{
				// name will be the function name
				Name: name,
				// source.Name will be the container name
				Instance:  fmt.Sprintf("%s/%s", source.Node(), source.Name()),
				Text:      msg,
				Timestamp: ts,
			}
			msgCount = msgCount + 1
		case <-ctx.Done():
			log.Println("LogRequestor: log stream context done")
			return
		case <-controller.Done():
			log.Println("LogRequestor: log controller done")
			return
		}
	}
}

func extractTimestampAndMsg(logText string) (string, time.Time) {
	// first 32 characters is the k8s timestamp
	parts := strings.SplitN(logText, " ", 2)
	ts, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		log.Printf("error: invalid timestamp '%s'\n", parts[0])
		return "", time.Time{}
	}

	if len(parts) == 2 {
		return parts[1], ts
	}

	return "", ts
}
