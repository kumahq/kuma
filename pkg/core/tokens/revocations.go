package tokens

import (
	"context"
	"strings"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
)

// Revocations keeps track of revoked tokens.
// If only one token is compromised, it's more convenient to revoke it instead of rotate signing key and regenerate all tokens.
// Revocation list is stored as Secret (in case of mesh scoped tokens) or GlobalSecret (global scoped tokens).
// IDs of token are stored in secret in comma separated format: "id1,id2".
type Revocations interface {
	IsRevoked(ctx context.Context, id string) (bool, error)
}

func NewRevocations(manager manager.ReadOnlyResourceManager, revocationKey core_model.ResourceKey) Revocations {
	return &secretRevocations{
		manager:       manager,
		revocationKey: revocationKey,
	}
}

type secretRevocations struct {
	manager       manager.ReadOnlyResourceManager
	revocationKey core_model.ResourceKey
}

func (s *secretRevocations) IsRevoked(ctx context.Context, id string) (bool, error) {
	data, err := s.getSecretData(ctx)
	if err != nil {
		return false, err
	}
	if len(data) == 0 {
		return false, nil
	}
	rawIds := strings.TrimSuffix(string(data), "\n")
	ids := strings.Split(rawIds, ",")
	for _, revokedId := range ids {
		if revokedId == id {
			return true, nil
		}
	}
	return false, nil
}

func (s *secretRevocations) getSecretData(ctx context.Context) ([]byte, error) {
	var resource core_model.Resource
	if s.revocationKey.Mesh == "" {
		resource = system.NewGlobalSecretResource()
	} else {
		resource = system.NewSecretResource()
	}
	if err := s.manager.Get(ctx, resource, core_store.GetBy(s.revocationKey)); err != nil {
		if core_store.IsResourceNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return resource.GetSpec().(*system_proto.Secret).GetData().GetValue(), nil
}
