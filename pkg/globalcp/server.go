package globalcp

import (
	"net/http"
	"time"

	"github.com/Kong/kuma/pkg/core"
)

var (
	globalCPLog = core.Log.WithName("global-cp")
)

type (
	GlobalCP interface {
		Start(<-chan struct{}) error
	}

	LocalCPList map[string]string

	GlobalCPPoller struct {
		localCPList LocalCPList
		newTicker   func() *time.Ticker
	}
)

const (
	tickInterval = 500 * time.Millisecond
)

func NewGlobalCPPoller(localCPList LocalCPList) (GlobalCP, error) {
	return &GlobalCPPoller{
		localCPList: localCPList,
		newTicker: func() *time.Ticker {
			return time.NewTicker(tickInterval)
		},
	}, nil
}

func (g *GlobalCPPoller) Start(stop <-chan struct{}) error {
	ticker := g.newTicker()
	defer ticker.Stop()

	globalCPLog.Info("starting the Global CP polling")
	for {
		select {
		case <-ticker.C:
			g.pollLocalCPs()
		case <-stop:
			return nil
		}
	}
}

func (g *GlobalCPPoller) pollLocalCPs() {
	for name, url := range g.localCPList {
		response, err := http.Get(url)
		if err != nil {
			globalCPLog.Info(name + " at " + url + " did not respond")
		}

		if response.StatusCode != http.StatusOK {
			globalCPLog.Info(name + " at " + url + " responded with" + response.Status)
		}
		response.Body.Close()
	}
}
