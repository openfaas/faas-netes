package kail

import (
	"sort"

	"github.com/boz/kcache/nsname"
	"k8s.io/api/core/v1"
)

type ContainerFilter interface {
	Accept(cs v1.ContainerStatus) bool
}

func NewContainerFilter(names []string) ContainerFilter {
	return containerFilter(names)
}

type containerFilter []string

func (cf containerFilter) Accept(cs v1.ContainerStatus) bool {
	if cs.State.Running == nil {
		return false
	}
	if len(cf) == 0 {
		return true
	}
	for _, name := range cf {
		if name == cs.Name {
			return true
		}
	}
	return false
}

func sourcesForPod(filter ContainerFilter, pod *v1.Pod) (nsname.NSName, map[eventSource]bool) {
	id := nsname.ForObject(pod)
	sources := make(map[eventSource]bool)

	for _, cstatus := range pod.Status.ContainerStatuses {
		if filter.Accept(cstatus) {
			source := eventSource{id, cstatus.Name, pod.Spec.NodeName}
			sources[source] = true
		}
	}

	for _, cstatus := range pod.Status.InitContainerStatuses {
		if filter.Accept(cstatus) {
			source := eventSource{id, cstatus.Name, pod.Spec.NodeName}
			sources[source] = true
		}
	}

	return id, sources
}

func SourcesForPod(
	filter ContainerFilter, pod *v1.Pod) (nsname.NSName, []EventSource) {

	id, internal := sourcesForPod(filter, pod)
	sources := make([]EventSource, 0, len(internal))

	for source, _ := range internal {
		sources = append(sources, source)
	}

	sort.Slice(sources, func(a, b int) bool {
		na := sources[a].Namespace() + sources[a].Name()
		nb := sources[b].Namespace() + sources[b].Name()
		return na < nb
	})

	return id, sources
}
