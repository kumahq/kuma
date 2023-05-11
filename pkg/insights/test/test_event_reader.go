package test

import "github.com/kumahq/kuma/pkg/events"

type TestEventReader struct {
	Ch chan events.Event
}

func (t *TestEventReader) Recv() <-chan events.Event {
	return t.Ch
}

func (t *TestEventReader) Close() {
	close(t.Ch)
}

type TestEventReaderFactory struct {
	Reader *TestEventReader
}

func (t *TestEventReaderFactory) New() events.Listener {
	return t.Reader
}
