package samples

import (
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

func SampleSigningKeySecretBuilder() *builders.SecretBuilder {
	return builders.Secret().
		WithName("dataplane-token-signing-key-1").
		WithStringValue(SampleSigningKeyValue)
}

func SampleSigningKeySecret() *system.SecretResource {
	return SampleSigningKeySecretBuilder().Build()
}

func SampleSecretBuilder() *builders.SecretBuilder {
	return builders.Secret().
		WithStringValue(SampleSigningKeyValue)
}

func SampleGlobalSecretAdminCa() *system.GlobalSecretResource {
	globalSecret := system.NewGlobalSecretResource()
	globalSecret.Spec.Data = &wrapperspb.BytesValue{
		Value: []byte{},
	}
	globalSecret.SetMeta(&test_model.ResourceMeta{
		Name: system.EnvoyAdminCA,
	})
	return globalSecret
}
