package v1alpha1

func (c *TrafficRoute_Conf) HasWildcard() bool {
	for _, destination := range c.GetSplit() {
		if destination.Destination[ServiceTag] == MatchAllTag {
			return true
		}
	}
	return false
}
