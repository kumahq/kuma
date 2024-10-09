package parameters

import (
	"fmt"
)

type RedirectParameter WrappingParameter

func newRedirectParameter(param string, params ...string) *RedirectParameter {
	return (*RedirectParameter)(NewWrappingParameter(param, params...))
}

func Redirect(parameter *RedirectParameter) *JumpParameter {
	return newJumpParameter("REDIRECT", parameter.parameters...)
}

func ToPort[T ~uint16](port T) *RedirectParameter {
	return newRedirectParameter("--to-ports", fmt.Sprintf("%d", port))
}
