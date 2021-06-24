package v1alpha1

func (x *ZoneIngressInsight) GetSubscription(id string) (int, *DiscoverySubscription) {
	for i, s := range x.GetSubscriptions() {
		if s.Id == id {
			return i, s
		}
	}
	return -1, nil
}

func (x *ZoneIngressInsight) UpdateSubscription(s *DiscoverySubscription) {
	if x == nil {
		return
	}
	i, old := x.GetSubscription(s.Id)
	if old != nil {
		x.Subscriptions[i] = s
	} else {
		x.Subscriptions = append(x.Subscriptions, s)
	}
}
