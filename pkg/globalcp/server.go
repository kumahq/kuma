package globalcp

import (
	"context"
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
	}

	LocalCP struct {
		URL    string `json:"url"`
		Active bool   `json:"active"`
	}

	LocalCPMap map[string]*LocalCP

	GlobalCPPoller struct {
		sync.RWMutex
		localCPMap LocalCPMap
		server     *http.Server
		newTicker  func() *time.Ticker
	}
)

const (
	tickInterval            = 5 * time.Second
	globalCPStatisticsAddrt = ":5656"
)

func NewGlobalCPPoller(localCPList map[string]string) (GlobalCP, error) {
	mux := http.NewServeMux()
	poller := &GlobalCPPoller{
		localCPMap: LocalCPMap{},
		server: &http.Server{
			Addr:    globalCPStatisticsAddrt,
			Handler: mux,
		},
		newTicker: func() *time.Ticker {
			return time.NewTicker(tickInterval)
		},
	}

	for name, url := range localCPList {
		poller.localCPMap[name] = &LocalCP{URL: url, Active: true}
	}

	mux.HandleFunc("/", poller.StatusHandler)
	return poller, nil
}

func (g *GlobalCPPoller) Start(stop <-chan struct{}) error {
	ticker := g.newTicker()
	defer ticker.Stop()

	// update the status before running the API
	g.pollLocalCPs()

	errChan := make(chan error)
	go func() {
		err := g.server.ListenAndServe()
		if err != nil {
			switch err {
			case http.ErrServerClosed:
				globalCPLog.Info("Shutting down server")
			default:
				globalCPLog.Error(err, "Could not start an HTTP Server")
				errChan <- err
			}
		}
	}()

	globalCPLog.Info("starting the Global CP polling")
	for {
		select {
		case <-ticker.C:
			g.pollLocalCPs()
		case <-stop:
			globalCPLog.Info("Stopping down API Server")
			return g.server.Shutdown(context.Background())
		case err := <-errChan:
			return err
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

func (g *GlobalCPPoller) StatusHandler(writer http.ResponseWriter, request *http.Request) {
	g.RLock()
	defer g.RUnlock()
	writer.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(writer).Encode(g.localCPMap); err != nil {
		globalCPLog.Error(err, "failed marshaling response")
	}
}
