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
	// Do a list operation instead of get because of the cache in ReadOnlyResourceManager
	// For the majority of cases, users do not set revocation secret.
	// We don't cache not found result of get operation, so with many execution to IsRevoked() we would send a lot of requests to a DB.
	// We could do get requests and preserve separate cache here (taking into account resource not found),
	// but there is a high chance that SecretResourceList is already in the cache, because of XDS reconciliation, so we can avoid I/O at all.
	var resources core_model.ResourceList
	if s.revocationKey.Mesh == "" {
		resources = &system.GlobalSecretResourceList{}
	} else {
		resources = &system.SecretResourceList{}
	}

	if err := s.manager.List(ctx, resources, core_store.ListByMesh(s.revocationKey.Mesh)); err != nil {
		return nil, err
	}
	for _, res := range resources.GetItems() {
		if res.GetMeta().GetName() == s.revocationKey.Name {
			return res.GetSpec().(*system_proto.Secret).GetData().GetValue(), nil
		}
	}
	return nil, nil
}
