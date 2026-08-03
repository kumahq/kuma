package clusters

type ClusterItem struct {
	Name         string       `json:"name"`
	HostStatuses []HostStatus `json:"host_statuses"`
}

type HostStatus struct {
	HealthStatus *HealthStatus `json:"health_status,omitempty"`
	Priority     *int          `json:"priority,omitempty"`
	Locality     *Locality     `json:"locality,omitempty"`
}

type HealthStatus struct {
	FailedActiveHealthCheck bool `json:"failed_active_health_check"`
}

type Locality struct {
	Zone string `json:"zone"`
}

type Clusters struct {
	Clusters []ClusterItem `json:"cluster_statuses"`
}

func (c *Clusters) GetCluster(clusterName string) *ClusterItem {
	for _, cluster := range c.Clusters {
		if cluster.Name == clusterName {
			return &cluster
		}
	}
	return nil
}

func (ci *ClusterItem) GetPriorityForZone(zone string) int {
	for _, hs := range ci.HostStatuses {
		if hs.Locality != nil && hs.Locality.Zone == zone {
			if hs.Priority == nil {
				return 0
			} else {
				return *hs.Priority
			}
		}
	}
	return 0
}
