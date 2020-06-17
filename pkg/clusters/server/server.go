package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Kong/kuma/pkg/config/clusters"

	"github.com/Kong/kuma/pkg/core"
)

var (
	clusterStatusLog = core.Log.WithName("cluster-status")
)

type (
	ClusterStatusServer interface {
		Start(<-chan struct{}) error
		StatusHandler(writer http.ResponseWriter)
	}

	Cluster struct {
		Name   string `json:"name"`
		URL    string `json:"url"`
		Active bool   `json:"active"`
	}

	Clusters []*Cluster

	ClustersStatusPoller struct {
		sync.RWMutex
		clusters  Clusters
		client    http.Client
		newTicker func() *time.Ticker
	}
)

const (
	tickInterval = 1 * time.Second
	httpTimeout  = tickInterval / 100
)

func NewClustersStatusPoller(clusters *clusters.ClustersConfig) (ClusterStatusServer, error) {
	poller := &ClustersStatusPoller{
		clusters: []*Cluster{},
		client: http.Client{
			Timeout: httpTimeout,
		},
		newTicker: func() *time.Ticker {
			return time.NewTicker(tickInterval)
		},
	}

	for _, cluster := range clusters.Clusters {
		// ignore the Ingress for now
		poller.clusters = append(poller.clusters, &Cluster{
			Name:   cluster.Local.Address, // init the name of the cluster with its address
			URL:    cluster.Local.Address,
			Active: false,
		})
	}

	return poller, nil
}

func (p *ClustersStatusPoller) Start(stop <-chan struct{}) error {
	ticker := p.newTicker()
	defer ticker.Stop()

	// update the status before running the API
	p.pollClusters()

	clusterStatusLog.Info("starting the Clusters polling")
	for {
		select {
		case <-ticker.C:
			p.pollClusters()
		case <-stop:
			clusterStatusLog.Info("Stopping down API Server")
			return nil
		}
	}
}

func (p *ClustersStatusPoller) pollClusters() {
	p.Lock()
	defer p.Unlock()

	for _, cluster := range p.clusters {
		response, err := p.client.Get(cluster.URL)
		if err != nil {
			if cluster.Active {
				clusterStatusLog.Info(fmt.Sprintf("%s at %s did not respond", cluster.Name, cluster.URL))
				cluster.Active = false
			}

			continue
		}

		cluster.Active = response.StatusCode == http.StatusOK
		if !cluster.Active {
			clusterStatusLog.Info(fmt.Sprintf("%s at %s responded with %s", cluster.Name, cluster.URL, response.Status))
		}

		response.Body.Close()
	}
}

func (p *ClustersStatusPoller) StatusHandler(writer http.ResponseWriter) {
	p.RLock()
	defer p.RUnlock()
	writer.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(writer).Encode(p.clusters); err != nil {
		clusterStatusLog.Error(err, "failed marshaling response")
	}
}
