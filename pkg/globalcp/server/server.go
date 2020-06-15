package server

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/Kong/kuma/pkg/core"
)

var (
	globalCPLog = core.Log.WithName("global-cp")
)

type (
	GlobalCP interface {
		Start(<-chan struct{}) error
		StatusHandler(writer http.ResponseWriter)
	}

	LocalCP struct {
		URL    string `json:"url"`
		Active bool   `json:"active"`
	}

	LocalCPMap map[string]*LocalCP

	GlobalCPPoller struct {
		sync.RWMutex
		localCPMap LocalCPMap
		newTicker  func() *time.Ticker
	}
)

const (
	tickInterval = 5 * time.Second
)

func NewGlobalCPPoller(localCPList map[string]string) (GlobalCP, error) {
	poller := &GlobalCPPoller{
		localCPMap: LocalCPMap{},
		newTicker: func() *time.Ticker {
			return time.NewTicker(tickInterval)
		},
	}

	for name, url := range localCPList {
		poller.localCPMap[name] = &LocalCP{URL: url, Active: true}
	}

	return poller, nil
}

func (g *GlobalCPPoller) Start(stop <-chan struct{}) error {
	ticker := g.newTicker()
	defer ticker.Stop()

	// update the status before running the API
	g.pollLocalCPs()

	globalCPLog.Info("starting the Global CP polling")
	for {
		select {
		case <-ticker.C:
			g.pollLocalCPs()
		case <-stop:
			globalCPLog.Info("Stopping down API Server")
			return nil
		}
	}
}

func (g *GlobalCPPoller) pollLocalCPs() {
	for name, localCP := range g.localCPMap {
		response, err := http.Get(localCP.URL)
		if err != nil {
			if localCP.Active {
				globalCPLog.Info(name + " at " + localCP.URL + " did not respond")
				g.Lock()
				localCP.Active = false
				g.Unlock()
			}

			continue
		}

		g.Lock()
		localCP.Active = response.StatusCode == http.StatusOK
		g.Unlock()
		if !localCP.Active {
			globalCPLog.Info(name + " at " + localCP.URL + " responded with" + response.Status)
		}

		response.Body.Close()
	}
}

func (g *GlobalCPPoller) StatusHandler(writer http.ResponseWriter) {
	g.RLock()
	defer g.RUnlock()
	writer.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(writer).Encode(g.localCPMap); err != nil {
		globalCPLog.Error(err, "failed marshaling response")
	}
}
