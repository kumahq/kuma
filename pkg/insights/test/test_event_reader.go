package test

import "github.com/kumahq/kuma/pkg/events"

type TestEventReader struct {
	Ch chan events.Event
}

func (t *TestEventReader) Recv() <-chan events.Event {
	return t.Ch
}

// This method should be called only once. In tests, we
// have only one call to subscribe, and we want to avoid
// closing the channel twice, as it may lead to a panic.
func (t *TestEventReader) Close() {
	close(t.Ch)
}

type TestEventReaderFactory struct {
	Reader *TestEventReader
}

func (t *TestEventReaderFactory) Subscribe(...events.Predicate) events.Listener {
	return t.Reader
}
