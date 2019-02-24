package kail

import (
	"context"

	logutil "github.com/boz/go-logutil"
	"github.com/boz/kcache/filter"
	"github.com/boz/kcache/join"
	"github.com/boz/kcache/nsname"
	"github.com/boz/kcache/types/daemonset"
	"github.com/boz/kcache/types/deployment"
	"github.com/boz/kcache/types/ingress"
	"github.com/boz/kcache/types/pod"
	"github.com/boz/kcache/types/replicaset"
	"github.com/boz/kcache/types/replicationcontroller"
	"github.com/boz/kcache/types/service"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

type DSBuilder interface {
	WithIgnore(selectors ...labels.Selector) DSBuilder
	WithSelectors(selectors ...labels.Selector) DSBuilder
	WithPods(id ...nsname.NSName) DSBuilder
	WithNamespace(name ...string) DSBuilder
	WithService(id ...nsname.NSName) DSBuilder
	WithNode(name ...string) DSBuilder
	WithRC(id ...nsname.NSName) DSBuilder
	WithRS(id ...nsname.NSName) DSBuilder
	WithDS(id ...nsname.NSName) DSBuilder
	WithDeployment(id ...nsname.NSName) DSBuilder
	WithIngress(id ...nsname.NSName) DSBuilder

	Create(ctx context.Context, cs kubernetes.Interface) (DS, error)
}

func NewDSBuilder() DSBuilder {
	return &dsBuilder{}
}

type dsBuilder struct {
	ignore      []labels.Selector
	selectors   []labels.Selector
	pods        []nsname.NSName
	namespaces  []string
	services    []nsname.NSName
	nodes       []string
	rcs         []nsname.NSName
	rss         []nsname.NSName
	dss         []nsname.NSName
	deployments []nsname.NSName
	ingresses   []nsname.NSName
}

func (b *dsBuilder) WithIgnore(selector ...labels.Selector) DSBuilder {
	b.ignore = append(b.ignore, selector...)
	return b
}

func (b *dsBuilder) WithSelectors(selectors ...labels.Selector) DSBuilder {
	b.selectors = append(b.selectors, selectors...)
	return b
}

func (b *dsBuilder) WithPods(id ...nsname.NSName) DSBuilder {
	b.pods = append(b.pods, id...)
	return b
}

func (b *dsBuilder) WithNamespace(name ...string) DSBuilder {
	b.namespaces = append(b.namespaces, name...)
	return b
}

func (b *dsBuilder) WithService(id ...nsname.NSName) DSBuilder {
	b.services = append(b.services, id...)
	return b
}

func (b *dsBuilder) WithNode(name ...string) DSBuilder {
	b.nodes = append(b.nodes, name...)
	return b
}

func (b *dsBuilder) WithRC(id ...nsname.NSName) DSBuilder {
	b.rcs = append(b.rcs, id...)
	return b
}

func (b *dsBuilder) WithRS(id ...nsname.NSName) DSBuilder {
	b.rss = append(b.rss, id...)
	return b
}

func (b *dsBuilder) WithDS(id ...nsname.NSName) DSBuilder {
	b.dss = append(b.dss, id...)
	return b
}

func (b *dsBuilder) WithDeployment(id ...nsname.NSName) DSBuilder {
	b.deployments = append(b.deployments, id...)
	return b
}

func (b *dsBuilder) WithIngress(id ...nsname.NSName) DSBuilder {
	b.ingresses = append(b.ingresses, id...)
	return b
}

func (b *dsBuilder) Create(ctx context.Context, cs kubernetes.Interface) (DS, error) {
	log := logutil.FromContextOrDefault(ctx)

	ds := &datastore{
		readych: make(chan struct{}),
		donech:  make(chan struct{}),
		log:     log.WithComponent("kail.ds"),
	}

	log = log.WithComponent("kail.ds.builder")

	base, err := pod.NewController(ctx, log, cs, "")
	if err != nil {
		return nil, log.Err(err, "base pod controller")
	}

	ds.podBase = base
	ds.pods, err = base.CloneWithFilter(filter.Null())
	if err != nil {
		ds.closeAll()
		return nil, log.Err(err, "null filter")
	}

	if len(b.ignore) != 0 {
		filters := make([]filter.Filter, 0, len(b.ignore))
		for _, selector := range b.ignore {
			filters = append(filters, filter.Not(filter.Selector(selector)))
		}
		ds.pods, err = ds.pods.CloneWithFilter(filter.And(filters...))
		if err != nil {
			ds.closeAll()
			return nil, log.Err(err, "labels filter")
		}
	}

	if len(b.selectors) != 0 {
		filters := make([]filter.Filter, 0, len(b.selectors))
		for _, selector := range b.selectors {
			filters = append(filters, filter.Selector(selector))
		}
		ds.pods, err = ds.pods.CloneWithFilter(filter.And(filters...))
		if err != nil {
			ds.closeAll()
			return nil, log.Err(err, "labels filter")
		}
	}

	if len(b.pods) != 0 {
		ds.pods, err = ds.pods.CloneWithFilter(filter.NSName(b.pods...))
		if err != nil {
			ds.closeAll()
			return nil, log.Err(err, "pods filter")
		}
	}

	if sz := len(b.namespaces); sz > 0 {
		ids := make([]nsname.NSName, 0, sz)
		for _, ns := range b.namespaces {
			ids = append(ids, nsname.New(ns, ""))
		}

		ds.pods, err = ds.pods.CloneWithFilter(filter.NSName(ids...))
		if err != nil {
			ds.closeAll()
			return nil, log.Err(err, "namespace filter")
		}
	}

	if len(b.nodes) != 0 {
		ds.pods, err = ds.pods.CloneWithFilter(pod.NodeFilter(b.nodes...))
		if err != nil {
			ds.closeAll()
			return nil, log.Err(err, "node filter")
		}
	}

	if len(b.services) != 0 {
		ds.servicesBase, err = service.NewController(ctx, log, cs, "")
		if err != nil {
			ds.closeAll()
			return nil, log.Err(err, "service base controller")
		}

		ds.services, err = ds.servicesBase.CloneWithFilter(filter.NSName(b.services...))
		if err != nil {
			ds.closeAll()
			return nil, log.Err(err, "service controller")
		}

		ds.pods, err = join.ServicePods(ctx, ds.services, ds.pods)
		if err != nil {
			ds.closeAll()
			return nil, log.Err(err, "service join")
		}
	}

	if len(b.rcs) != 0 {
		ds.rcsBase, err = replicationcontroller.NewController(ctx, log, cs, "")
		if err != nil {
			ds.closeAll()
			return nil, log.Err(err, "rc base controller")
		}

		ds.rcs, err = ds.rcsBase.CloneWithFilter(filter.NSName(b.rcs...))
		if err != nil {
			ds.closeAll()
			return nil, log.Err(err, "rc controller")
		}

		ds.pods, err = join.RCPods(ctx, ds.rcs, ds.pods)
		if err != nil {
			ds.closeAll()
			return nil, log.Err(err, "rc join")
		}
	}

	if len(b.rss) != 0 {
		ds.rssBase, err = replicaset.NewController(ctx, log, cs, "")
		if err != nil {
			ds.closeAll()
			return nil, log.Err(err, "rs base controller")
		}

		ds.rss, err = ds.rssBase.CloneWithFilter(filter.NSName(b.rss...))
		if err != nil {
			ds.closeAll()
			return nil, log.Err(err, "rs controller")
		}

		ds.pods, err = join.RSPods(ctx, ds.rss, ds.pods)
		if err != nil {
			ds.closeAll()
			return nil, log.Err(err, "rs join")
		}
	}

	if len(b.dss) != 0 {
		ds.dssBase, err = daemonset.NewController(ctx, log, cs, "")
		if err != nil {
			ds.closeAll()
			return nil, log.Err(err, "ds base controller")
		}

		ds.dss, err = ds.dssBase.CloneWithFilter(filter.NSName(b.dss...))
		if err != nil {
			ds.closeAll()
			return nil, log.Err(err, "ds controller")
		}

		ds.pods, err = join.DaemonSetPods(ctx, ds.dss, ds.pods)
		if err != nil {
			ds.closeAll()
			return nil, log.Err(err, "ds join")
		}
	}

	if len(b.deployments) != 0 {
		ds.deploymentsBase, err = deployment.NewController(ctx, log, cs, "")
		if err != nil {
			ds.closeAll()
			return nil, log.Err(err, "deployment base controller")
		}

		ds.deployments, err = ds.deploymentsBase.CloneWithFilter(filter.NSName(b.deployments...))
		if err != nil {
			ds.closeAll()
			return nil, log.Err(err, "deployment controller")
		}

		ds.pods, err = join.DeploymentPods(ctx, ds.deployments, ds.pods)
		if err != nil {
			ds.closeAll()
			return nil, log.Err(err, "deployment join")
		}
	}

	if len(b.ingresses) != 0 {
		ds.ingressesBase, err = ingress.NewController(ctx, log, cs, "")
		if err != nil {
			ds.closeAll()
			return nil, log.Err(err, "ingress base controller")
		}

		if ds.servicesBase == nil {
			ds.servicesBase, err = service.NewController(ctx, log, cs, "")
			if err != nil {
				ds.closeAll()
				return nil, log.Err(err, "service base controller")
			}
			ds.services = ds.servicesBase
		}

		ds.ingresses, err = ds.ingressesBase.CloneWithFilter(filter.NSName(b.ingresses...))
		if err != nil {
			ds.closeAll()
			return nil, log.Err(err, "ingresses controller")
		}

		ds.pods, err = join.IngressPods(ctx, ds.ingresses, ds.services, ds.pods)
		if err != nil {
			ds.closeAll()
			return nil, log.Err(err, "ingress join")
		}
	}

	ds.run(ctx)

	return ds, nil
}
