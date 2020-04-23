package k8s

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"log"
	"strings"
	"time"

	"github.com/pkg/errors"
	"k8s.io/client-go/informers/internalinterfaces"

	"k8s.io/client-go/tools/cache"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	// podInformerResync is the period between cache syncs in the pod informer
	podInformerResync = 5 * time.Second

	// defaultLogSince is the fallback log stream history
	defaultLogSince = 5 * time.Minute

	// LogBufferSize number of log messages that may be buffered
	LogBufferSize = 500 * 2
)

// Log is the object which will be used together with the template to generate
// the output.
type Log struct {
	// Text is the log message itself
	Text string `json:"text"`

	// Namespace of the pod
	Namespace string `json:"namespace"`

	// PodName of the instance
	PodName string `json:"podName"`

	// FunctionName of the pod
	FunctionName string `json:"FunctionName"`

	// Timestamp of the message
	Timestamp time.Time `json:"timestamp"`
}

// GetLogs returns a channel of logs for the given function
func GetLogs(ctx context.Context, client kubernetes.Interface, functionName, namespace string, tail int64, since *time.Time, follow bool) (<-chan Log, error) {
	added, err := startFunctionPodInformer(ctx, client, functionName, namespace)
	if err != nil {
		return nil, err
	}

	logs := make(chan Log, LogBufferSize)

	go func() {
		var watching uint
		defer close(logs)

		finished := make(chan error)

		for {
			select {
			case <-ctx.Done():
				return
			case <-finished:
				watching--
				if watching == 0 && !follow {
					return
				}
			case p := <-added:
				watching++
				go func() {
					finished <- podLogs(ctx, client.CoreV1().Pods(namespace), p, functionName, namespace, tail, since, follow, logs)
				}()
			}
		}
	}()

	return logs, nil
}

// podLogs returns a stream of logs lines from the specified pod
func podLogs(ctx context.Context, i v1.PodInterface, pod, container, namespace string, tail int64, since *time.Time, follow bool, dst chan<- Log) error {
	log.Printf("Logger: starting log stream for %s\n", pod)
	defer log.Printf("Logger: stopping log stream for %s\n", pod)

	opts := &corev1.PodLogOptions{
		Follow:     follow,
		Timestamps: true,
		Container:  container,
	}

	if tail > 0 {
		opts.TailLines = &tail
	}

	if opts.TailLines == nil || since != nil {
		opts.SinceSeconds = parseSince(since)
	}

	stream, err := i.GetLogs(pod, opts).Stream()
	if err != nil {
		return err
	}
	defer stream.Close()

	done := make(chan error)
	go func() {
		reader := bufio.NewReader(stream)
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				done <- err
				return
			}
			msg, ts := extractTimestampAndMsg(string(bytes.Trim(line, "\x00")))
			dst <- Log{Timestamp: ts, Text: msg, PodName: pod, FunctionName: container}
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		if err != io.EOF {
			return err
		}
		return nil
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

// parseSince returns the time.Duration of the requested Since value _or_ 5 minutes
func parseSince(r *time.Time) *int64 {
	var since int64
	if r == nil || r.IsZero() {
		since = int64(defaultLogSince.Seconds())
		return &since
	}
	since = int64(time.Since(*r).Seconds())
	return &since
}

// startFunctionPodInformer will gather the list of existing Pods for the function, it will then watch
// and watch for newly added or deleted function instances.
func startFunctionPodInformer(ctx context.Context, client kubernetes.Interface, functionName, namespace string) (<-chan string, error) {
	functionSelector := &metav1.LabelSelector{
		MatchLabels: map[string]string{"faas_function": functionName},
	}
	selector, err := metav1.LabelSelectorAsSelector(functionSelector)
	if err != nil {
		err = errors.Wrap(err, "unable to build function selector")
		log.Printf("PodInformer: %s", err)
		return nil, err
	}

	log.Printf("PodInformer: starting informer for %s in: %s\n", selector.String(), namespace)
	factory := informers.NewFilteredSharedInformerFactory(
		client,
		podInformerResync,
		namespace,
		withLabels(selector.String()),
	)

	podInformer := factory.Core().V1().Pods()
	podsResp, err := client.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		log.Printf("PodInformer: %s", err)
		return nil, err
	}

	pods := podsResp.Items
	if len(pods) == 0 {
		err = errors.New("no matching instances found")
		log.Printf("PodInformer: %s", err)
		return nil, err
	}

	// prepare channel with enough space for the current instance set
	added := make(chan string, len(pods))
	podInformer.Informer().AddEventHandler(&podLoggerEventHandler{
		added: added,
	})

	// will add existing pods to the chan and then listen for any new pods
	go podInformer.Informer().Run(ctx.Done())
	go func() {
		<-ctx.Done()
		close(added)
	}()

	return added, nil
}

func withLabels(selector string) internalinterfaces.TweakListOptionsFunc {
	return func(opts *metav1.ListOptions) {
		opts.LabelSelector = selector
	}
}

type podLoggerEventHandler struct {
	cache.ResourceEventHandler
	added   chan<- string
	deleted chan<- string
}

func (h *podLoggerEventHandler) OnAdd(obj interface{}) {
	pod := obj.(*corev1.Pod)
	log.Printf("PodInformer: adding instance: %s", pod.Name)
	h.added <- pod.Name
}

func (h *podLoggerEventHandler) OnUpdate(oldObj, newObj interface{}) {
	// purposefully empty, we don't need to do anything for logs on update
}

func (h *podLoggerEventHandler) OnDelete(obj interface{}) {
	// this may not be needed, the log stream Reader _should_ close on its own without
	// us needing to watch and close it
	// pod := obj.(*corev1.Pod)
	// h.deleted <- pod.Name
}
