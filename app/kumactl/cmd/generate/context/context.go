package context

import (
	"github.com/kumahq/kuma/v2/pkg/core/tokens"
)

type GenerateContext struct {
	NewSigningKey func() ([]byte, error)
}

func DefaultGenerateContext() GenerateContext {
	return GenerateContext{
		NewSigningKey: tokens.NewSigningKey,
	}
}
