package kail

import "bytes"

const bufferMaxRetainSize = logBufsiz

type buffer interface {
	process([]byte) []Event
}

type _buffer struct {
	source EventSource
	prev   *bytes.Buffer
}

func newBuffer(source EventSource) buffer {
	return &_buffer{source, new(bytes.Buffer)}
}

func (b *_buffer) process(log []byte) []Event {

	var events []Event

	for end := bytes.IndexRune(log, '\n'); end >= 0 && len(log) > 0; end = bytes.IndexRune(log, '\n') {
		var ebuf []byte

		if plen := b.prev.Len(); plen > 0 {
			ebuf = make([]byte, plen+end)
			copy(ebuf, b.prev.Bytes())
			copy(ebuf[plen:], log[:end])
			b.prev.Reset()
		} else {
			ebuf = make([]byte, end)
			copy(ebuf, log[:end])
		}

		events = append(events, newEvent(b.source, ebuf))
		log = log[end+1:]
	}

	if sz := len(log); sz > 0 {
		b.prev.Write(log)
		if plen := b.prev.Len(); plen >= bufferMaxRetainSize {
			ebuf := make([]byte, plen)
			copy(ebuf, b.prev.Bytes())
			events = append(events, newEvent(b.source, ebuf))
			b.prev.Reset()
		}
	}

	return events
}
