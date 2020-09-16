package v1alpha1

// Matches is simply an alias for MatchTags to make source code more aesthetic.
func (es *ExternalService) Matches(selector TagSelector) bool {
	if es != nil {
		return es.MatchTags(selector)
	}
	return false
}

func (es *ExternalService) MatchTags(selector TagSelector) bool {
	return selector.Matches(es.Tags)
}
func (es *ExternalService) GetService() string {
	if es == nil {
		return ""
	}
	return es.Tags[ServiceTag]
}

func (es *ExternalService) GetProtocol() string {
	if es == nil {
		return ""
	}
	return es.Tags[ProtocolTag]
}

func (es *ExternalService) TagSet() MultiValueTagSet {
	tags := MultiValueTagSet{}
	for tag, value := range es.Tags {
		_, exists := tags[tag]
		if !exists {
			tags[tag] = map[string]bool{}
		}
		tags[tag][value] = true
	}

	return tags
}

func (es *ExternalService) GetIdentifyingService() string {
	services := es.TagSet().Values(ServiceTag)
	if len(services) > 0 {
		return services[0]
	}
	return ServiceUnknown
}
