package kail

import (
	"context"

	logutil "github.com/boz/go-logutil"
	"github.com/boz/kcache/types/daemonset"
	"github.com/boz/kcache/types/deployment"
	"github.com/boz/kcache/types/ingress"
	"github.com/boz/kcache/types/node"
	"github.com/boz/kcache/types/pod"
	"github.com/boz/kcache/types/replicaset"
	"github.com/boz/kcache/types/replicationcontroller"
	"github.com/boz/kcache/types/service"
)

type DS interface {
	Pods() pod.Controller
	Ready() <-chan struct{}
	Done() <-chan struct{}
	Close()
}

type datastore struct {
	podBase         pod.Controller
	servicesBase    service.Controller
	nodesBase       node.Controller
	rcsBase         replicationcontroller.Controller
	rssBase         replicaset.Controller
	dssBase         daemonset.Controller
	deploymentsBase deployment.Controller
	ingressesBase   ingress.Controller

	pods        pod.Controller
	services    service.Controller
	nodes       node.Controller
	rcs         replicationcontroller.Controller
	rss         replicaset.Controller
	dss         daemonset.Controller
	deployments deployment.Controller
	ingresses   ingress.Controller

	readych chan struct{}
	donech  chan struct{}
	log     logutil.Log
}

type cacheController interface {
	Close()
	Done() <-chan struct{}
	Ready() <-chan struct{}
}

func (ds *datastore) Pods() pod.Controller {
	return ds.pods
}

func (ds *datastore) Ready() <-chan struct{} {
	return ds.readych
}

func (ds *datastore) Done() <-chan struct{} {
	return ds.donech
}

func (ds *datastore) Close() {
	ds.closeAll()
}

func (ds *datastore) run(ctx context.Context) {
	go func() {
		select {
		case <-ctx.Done():
			ds.Close()
		case <-ds.Done():
		}
	}()

	go ds.waitReadyAll()
	go ds.waitDoneAll()
}

func (ds *datastore) waitReadyAll() {
	for _, c := range ds.controllers() {
		select {
		case <-c.Done():
			return
		case <-c.Ready():
		}
	}
	close(ds.readych)
}

func (ds *datastore) closeAll() {
	for _, c := range ds.controllers() {
		c.Close()
	}
}

func (ds *datastore) waitDoneAll() {
	defer close(ds.donech)
	for _, c := range ds.controllers() {
		<-c.Done()
	}
}

func (ds *datastore) controllers() []cacheController {

	potential := []cacheController{
		ds.podBase,
		ds.servicesBase,
		ds.nodesBase,
		ds.rcsBase,
		ds.rssBase,
		ds.dssBase,
		ds.deploymentsBase,
		ds.ingressesBase,
		ds.pods,
		ds.services,
		ds.nodes,
		ds.rcs,
		ds.rss,
		ds.dss,
		ds.deployments,
		ds.ingresses,
	}

	var existing []cacheController
	for _, c := range potential {
		if c != nil {
			existing = append(existing, c)
		}
	}
	return existing
}
