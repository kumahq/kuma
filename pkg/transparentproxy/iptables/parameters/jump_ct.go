package parameters

type CtParameter WrappingParameter

func newCtParameter(param string, params ...string) *CtParameter {
	return (*CtParameter)(NewWrappingParameter(param, params...))
}

func Ct(params ...*CtParameter) *JumpParameter {
	var parameters []string
	for _, param := range params {
		parameters = append(parameters, param.parameters...)
	}

	return newJumpParameter("CT", parameters...)
}

func Zone(id string) *CtParameter {
	return newCtParameter("--zone", id)
}
