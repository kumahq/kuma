package v1alpha1

// DataplaneOverview defines the projected state of a Dataplane. It is assembled for the
// API and never stored.
type DataplaneOverview struct {
	Dataplane        *Dataplane        `json:"dataplane,omitempty"`
	DataplaneInsight *DataplaneInsight `json:"dataplaneInsight,omitempty"`
}

func (o *DataplaneOverview) GetDataplane() *Dataplane {
	if o == nil {
		return nil
	}
	return o.Dataplane
}

func (o *DataplaneOverview) GetDataplaneInsight() *DataplaneInsight {
	if o == nil {
		return nil
	}
	return o.DataplaneInsight
}
