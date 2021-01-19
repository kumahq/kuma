package test

import "github.com/kumahq/kuma/pkg/events"

type TestEventReader struct {
	Ch chan events.Event
}

func (t *TestEventReader) Recv(stop <-chan struct{}) (events.Event, error) {
	return <-t.Ch, nil
}

type TestEventReaderFactory struct {
	Reader *TestEventReader
}

func (t *TestEventReaderFactory) New() events.Listener {
	return t.Reader
}
