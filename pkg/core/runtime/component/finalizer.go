package component

import (
	"sync"

	"github.com/kumahq/kuma/pkg/util/channels"
)

// Finalizer is a helper for implementing GracefulComponent
type Finalizer struct {
	finishCh chan struct{}
	mutex    sync.Mutex
}

func (f *Finalizer) Running() {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if f.finishCh == nil || channels.IsClosed(f.finishCh) {
		f.finishCh = make(chan struct{})
	}
}

func (f *Finalizer) Done() {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if f.finishCh != nil && !channels.IsClosed(f.finishCh) {
		close(f.finishCh)
	}
}

func (f *Finalizer) WaitForDone() {
	f.mutex.Lock()
	if f.finishCh != nil && !channels.IsClosed(f.finishCh) {
		waitCh := f.finishCh
		f.mutex.Unlock()
		<-waitCh
	} else {
		f.mutex.Unlock()
	}
}
