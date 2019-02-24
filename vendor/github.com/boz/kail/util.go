package kail

import (
	"fmt"

	"github.com/boz/kcache/nsname"
)

type EventSource interface {
	Namespace() string
	Name() string
	Container() string
	Node() string
}

type eventSource struct {
	id        nsname.NSName
	container string
	node      string
}

func (es eventSource) Namespace() string {
	return es.id.Namespace
}

func (es eventSource) Name() string {
	return es.id.Name
}

func (es eventSource) Container() string {
	return es.container
}

func (es eventSource) Node() string {
	return es.node
}

func (es eventSource) String() string {
	return fmt.Sprintf("%v/%v@%v",
		es.id.Namespace, es.id.Name, es.container)
}

type Event interface {
	Source() EventSource
	Log() []byte
}

func newEvent(source EventSource, log []byte) Event {
	return &event{source, log}
}

type event struct {
	source EventSource
	log    []byte
}

func (e *event) Source() EventSource {
	return e.source
}

func (e *event) Log() []byte {
	return e.log
}
