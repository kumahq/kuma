package issuer

import (
	"context"
	"strings"

	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
)

var RevocationsSecretKey = core_model.ResourceKey{
	Name: "user-token-revocations",
	Mesh: core_model.NoMesh,
}

type TokenRevocations interface {
	IsRevoked(id string) (bool, error)
}

func NewTokenRevocations(manager manager.ReadOnlyResourceManager) TokenRevocations {
	return &secretsTokenRevocations{
		manager: manager,
	}
}

type secretsTokenRevocations struct {
	manager manager.ReadOnlyResourceManager
}

func (s *secretsTokenRevocations) IsRevoked(id string) (bool, error) {
	secret := system.NewGlobalSecretResource()
	if err := s.manager.Get(context.Background(), secret, core_store.GetBy(RevocationsSecretKey)); err != nil {
		if core_store.IsResourceNotFound(err) { // todo gets not found are not cached, we should cache it by ourselves here
			return false, nil
		}
		return false, err
	}
	ids := strings.Split(string(secret.Spec.GetData().GetValue()), ",")
	for _, revokedId := range ids {
		if revokedId == id {
			return true, nil
		}
	}
	return false, nil
}
