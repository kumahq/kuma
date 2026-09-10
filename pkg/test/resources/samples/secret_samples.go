package samples

import (
	system_proto "github.com/kumahq/kuma/v3/api/system/v1alpha1"

	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
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
	globalSecret.Spec.Data = system_proto.Bytes([]byte{})
	globalSecret.SetMeta(&test_model.ResourceMeta{
		Name: system.EnvoyAdminCA,
	})
	return globalSecret
}
