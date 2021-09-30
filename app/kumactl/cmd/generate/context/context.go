package context

import "github.com/kumahq/kuma/pkg/tokens/builtin/issuer"

type GenerateContext struct {
	NewSigningKey func() ([]byte, error)
}

func DefaultGenerateContext() GenerateContext {
	return GenerateContext{
		NewSigningKey: issuer.NewSigningKey,
	}
}
