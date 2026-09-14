package parameters

func MatchUdp() *MatchParameter {
	return &MatchParameter{
		name:       "udp",
		parameters: []ParameterBuilder{},
	}
}
