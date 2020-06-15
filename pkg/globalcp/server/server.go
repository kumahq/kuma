package server

import (
	"encoding/json"
	"fmt"
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
		client     http.Client
		newTicker  func() *time.Ticker
	}
)

const (
	tickInterval = 1 * time.Second
	httpTimeout  = tickInterval / 100
)

func NewGlobalCPPoller(localCPList map[string]string) (GlobalCP, error) {
	poller := &GlobalCPPoller{
		localCPMap: LocalCPMap{},
		client: http.Client{
			Timeout: httpTimeout,
		},
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
	g.Lock()
	defer g.Unlock()

	for name, localCP := range g.localCPMap {
		response, err := g.client.Get(localCP.URL)
		if err != nil {
			if localCP.Active {
				globalCPLog.Info(fmt.Sprintf("%s at %s did not respond", name, localCP.URL))
				localCP.Active = false
			}

			continue
		}

		localCP.Active = response.StatusCode == http.StatusOK
		if !localCP.Active {
			globalCPLog.Info(fmt.Sprintf("%s at %s responded with %s", name, localCP.URL, response.Status))
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
