package kail

import (
	"context"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	lifecycle "github.com/boz/go-lifecycle"
	logutil "github.com/boz/go-logutil"
	"github.com/boz/kcache"
	"github.com/boz/kcache/nsname"
	"github.com/boz/kcache/types/pod"
)

const (
	eventBufsiz = 500
)

type Controller interface {
	Events() <-chan Event
	Close()
	Done() <-chan struct{}
}

func NewController(
	ctx context.Context,
	cs kubernetes.Interface,
	rc *rest.Config,
	pcontroller pod.Controller,
	filter ContainerFilter,
	since time.Duration) (Controller, error) {

	pods, err := pcontroller.Subscribe()
	if err != nil {
		return nil, err
	}

	initial, err := pods.Cache().List()
	if err != nil {
		pods.Close()
		return nil, err
	}

	lc := lifecycle.New()
	go lc.WatchContext(ctx)

	log := logutil.FromContextOrDefault(ctx)
	log = log.WithComponent("kail.controller")

	c := &controller{
		cs:        cs,
		rc:        rc,
		pods:      pods,
		filter:    filter,
		mconfig:   monitorConfig{since: since},
		eventch:   make(chan Event, eventBufsiz),
		monitorch: make(chan eventSource),
		monitors:  make(map[nsname.NSName]podMonitors),
		log:       log,
		ctx:       ctx,
		lc:        lc,
	}

	go c.run(initial)

	return c, nil
}

type controller struct {
	cs     kubernetes.Interface
	rc     *rest.Config
	pods   pod.Subscription
	filter ContainerFilter

	eventch   chan Event
	monitorch chan eventSource

	monitors monitors
	mconfig  monitorConfig

	log logutil.Log
	ctx context.Context
	lc  lifecycle.Lifecycle
}

type podMonitors map[eventSource]monitor
type monitors map[nsname.NSName]podMonitors

func (c *controller) Events() <-chan Event {
	return c.eventch
}

func (c *controller) Done() <-chan struct{} {
	return c.lc.Done()
}

func (c *controller) Close() {
	c.lc.Shutdown(nil)
}

func (c *controller) run(initial []*v1.Pod) {
	defer c.log.Un(c.log.Trace("run"))
	defer c.lc.ShutdownCompleted()

	peventch := c.pods.Events()
	shutdownch := c.lc.ShutdownRequest()
	draining := false

	c.createInitialMonitors(initial)

	for {

		c.log.Debugf("loop draining:%v monitors:%v", draining, len(c.monitors))

		if draining && len(c.monitors) == 0 {
			break
		}

		select {

		case err := <-shutdownch:
			c.log.Debugf("shutdown requested: %v", err)

			c.lc.ShutdownInitiated(err)
			shutdownch = nil
			draining = true

		case ev, ok := <-peventch:
			if !ok {
				c.log.Debugf("pods closed")

				peventch = nil

				if !draining {
					c.lc.ShutdownInitiated(nil)
					shutdownch = nil
					draining = true
				}

				break
			}

			if !draining {
				c.handlePodEvent(ev)
			}

		case source := <-c.monitorch:
			if pms, ok := c.monitors[source.id]; ok {
				if _, ok := pms[source]; ok {
					c.log.Debugf("removing source %v", source)
					delete(pms, source)
					if len(pms) == 0 {
						c.log.Debugf("removing pod %v", source.id)
						delete(c.monitors, source.id)
					}
					break
				}
			}
			c.log.Warnf("attempted to remove unknown source: %v", source)
		}
	}

	c.pods.Close()
	<-c.pods.Done()
}

func (c *controller) handlePodEvent(ev pod.Event) {
	pod := ev.Resource()
	id := nsname.ForObject(pod)

	c.log.Debugf("event %v %v/%v",
		ev.Type(), ev.Resource().GetName(), ev.Resource().GetNamespace())

	if ev.Type() == kcache.EventTypeDelete {
		if pms, ok := c.monitors[id]; ok {
			for _, pm := range pms {
				pm.Shutdown()
			}
		}
		return
	}

	c.ensureMonitorsForPod(pod)
}

func (c *controller) ensureMonitorsForPod(pod *v1.Pod) {
	id, sources := sourcesForPod(c.filter, pod)

	c.log.Debugf("pod %v/%v: %v containers ready",
		pod.GetNamespace(), pod.GetName(), len(sources))

	// delete monitors of not-ready containers
	if pms, ok := c.monitors[id]; ok {
		for source, pm := range pms {
			if !sources[source] {
				pm.Shutdown()
			}
		}
	}

	if len(sources) == 0 {
		return
	}

	pms, ok := c.monitors[id]
	if !ok {
		pms = make(map[eventSource]monitor)
	}

	for source, _ := range sources {
		if _, ok := pms[source]; ok {
			continue
		}
		pms[source] = c.createMonitor(source)
	}

	c.monitors[id] = pms
}

func (c *controller) createMonitor(source eventSource) monitor {
	defer c.log.Un(c.log.Trace("createMonitor(%v)", source))

	m := newMonitor(c, &source, c.mconfig)

	go func() {

		select {
		case <-m.Done():
		case <-c.lc.ShuttingDown():
			m.Shutdown()
			<-m.Done()
		}

		select {
		case c.monitorch <- source:
		case <-c.lc.Done():
			c.log.Warnf("done before monitor %v unregistered", source)
		}
	}()

	return m
}

func (c *controller) createInitialMonitors(pods []*v1.Pod) {
	defer c.log.Un(c.log.Trace("createInitialMonitors(pods=%v)", len(pods)))
	for _, pod := range pods {
		c.ensureMonitorsForPod(pod)
	}
}
