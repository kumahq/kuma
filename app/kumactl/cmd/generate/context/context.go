package context

import (
	"github.com/kumahq/kuma/pkg/core/tokens"
)

type GenerateContext struct {
	NewSigningKey func() ([]byte, error)
}

func DefaultGenerateContext() GenerateContext {
	return GenerateContext{
		NewSigningKey: tokens.NewSigningKey,
	}
}
