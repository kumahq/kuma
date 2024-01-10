package samples

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
)

func SampleSigningKeySecretBuilder() *builders.SecretBuilder {
	return builders.Secret().
		WithName("dataplane-token-signing-key-1").
		WithStringValue(SampleSigningKeyValue)
}

func SampleSigningKeySecret() *system.SecretResource {
	return SampleSigningKeySecretBuilder().Build()
}
