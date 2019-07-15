package v1alpha1

func (ds *DataplaneStatus) GetSubscription(id string) (int, *DiscoverySubscription) {
	for i, s := range ds.Subscriptions {
		if s.Id == id {
			return i, s
		}
	}
	return -1, nil
}

func (ds *DataplaneStatus) UpdateSubscription(s *DiscoverySubscription) {
	i, old := ds.GetSubscription(s.Id)
	if old != nil {
		ds.Subscriptions[i] = s
	} else {
		ds.Subscriptions = append(ds.Subscriptions, s)
	}
}
