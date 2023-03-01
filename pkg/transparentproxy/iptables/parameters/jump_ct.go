package parameters

type CtParameter struct {
	name  string
	value string
}

func (p *CtParameter) Build() []string {
	return []string{p.name, p.value}
}

func Ct(ctParameters ...*CtParameter) *JumpParameter {
	parameters := []string{"CT"}

	for _, parameter := range ctParameters {
		parameters = append(parameters, parameter.Build()...)
	}

	return &JumpParameter{
		parameters: parameters,
	}
}

func Zone(id string) *CtParameter {
	return &CtParameter{
		name:  "--zone",
		value: id,
	}
}
