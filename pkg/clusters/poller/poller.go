package poller

import (
	"fmt"
	"net"
	"net/url"
	"sync"
	"time"

	"github.com/Kong/kuma/pkg/config/mode"

	"github.com/Kong/kuma/pkg/core"
)

var (
	clusterStatusLog = core.Log.WithName("cluster-status")
)

type (
	ClusterStatusPoller interface {
		Start(<-chan struct{}) error
		NeedLeaderElection() bool
		Clusters() Clusters
	}

	Cluster struct {
		Name   string `json:"name"`
		URL    string `json:"url"`
		Active bool   `json:"active"`
	}

	Clusters []Cluster

	ClustersStatusPoller struct {
		sync.RWMutex
		clusters  Clusters
		newTicker func() *time.Ticker
	}
)

const (
	tickInterval = 15 * time.Second
	dialTimeout  = 100 * time.Millisecond
)

func NewClustersStatusPoller(globalConfig *mode.GlobalConfig) (ClusterStatusPoller, error) {
	poller := &ClustersStatusPoller{
		clusters: []Cluster{},
		newTicker: func() *time.Ticker {
			return time.NewTicker(tickInterval)
		},
	}

	for _, zone := range globalConfig.Zones {
		// ignore the Ingress for now
		poller.clusters = append(poller.clusters, Cluster{
			Name:   zone.Remote.Address, // init the name of the cluster with its address
			URL:    zone.Remote.Address,
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

func (p *ClustersStatusPoller) NeedLeaderElection() bool {
	return false
}

func (p *ClustersStatusPoller) pollClusters() {
	p.Lock()
	defer p.Unlock()

	for i, cluster := range p.clusters {
		u, err := url.Parse(cluster.URL)
		if err != nil {
			clusterStatusLog.Info(fmt.Sprintf("failed to parse URL %s", cluster.URL))
			continue
		}
		conn, err := net.DialTimeout("tcp", u.Host, dialTimeout)
		if err != nil {
			if cluster.Active {
				clusterStatusLog.Info(fmt.Sprintf("%s at %s did not respond", cluster.Name, cluster.URL))
				p.clusters[i].Active = false
			}
			continue
		}
		defer conn.Close()

		if !p.clusters[i].Active {
			clusterStatusLog.Info(fmt.Sprintf("%s responded", cluster.URL))
			p.clusters[i].Active = true
		}
	}
}

func (p *ClustersStatusPoller) Clusters() Clusters {
	p.RLock()
	defer p.RUnlock()
	newClusters := Clusters{}
	return append(newClusters, p.clusters...)
}
