package parameters

type JumpParameter WrappingParameter

func newJumpParameter(param string, params ...string) *JumpParameter {
	return (*JumpParameter)(NewWrappingParameter(param, params...))
}

func Jump(parameter *JumpParameter) *Parameter {
	return &Parameter{
		long:       "--jump",
		short:      "-j",
		parameters: []ParameterBuilder{(*WrappingParameter)(parameter)},
	}
}

func JumpConditional(
	condition bool,
	parameterTrue *JumpParameter,
	parameterFalse *JumpParameter,
) *Parameter {
	if !condition {
		return Jump(parameterFalse)
	}

	return Jump(parameterTrue)
}

func ToUserDefinedChain(chainName string) *JumpParameter {
	return newJumpParameter(chainName)
}

func Return() *JumpParameter {
	return newJumpParameter("RETURN")
}

func Drop() *JumpParameter {
	return newJumpParameter("DROP")
}
