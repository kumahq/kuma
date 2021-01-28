package admin

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"github.com/pkg/errors"

	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/tokens/builtin/issuer"
)

type EnvoyAdmin interface {
	GenerateAPIToken(dataplane *mesh_core.DataplaneResource) (string, error)
}

type envoyAdmin struct {
	rm manager.ResourceManager
}

func NewEnvoyAdmin(rm manager.ResourceManager) EnvoyAdmin {
	return &envoyAdmin{
		rm: rm,
	}
}

func (a *envoyAdmin) GenerateAPIToken(dataplane *mesh_core.DataplaneResource) (string, error) {
	mesh := dataplane.Meta.GetMesh()
	key, err := a.getSigningKeyString(mesh)
	if err != nil {
		return "", err
	}

	mac := hmac.New(sha256.New, []byte(key))
	_, err = mac.Write([]byte(dataplane.Meta.GetName()))
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(mac.Sum(nil)), nil
}
func (a *envoyAdmin) getSigningKeyString(mesh string) (string, error) {
	signingKey, err := issuer.CreateSigningKey()
	if err != nil {
		return "", errors.Wrap(err, "could not create a signing key")
	}
	key := issuer.SigningKeyResourceKey(mesh)
	err = a.rm.Get(context.Background(), signingKey, core_store.GetBy(key))
	if err == nil {
		return "", nil
	}
	if !core_store.IsResourceNotFound(err) {
		return "", errors.Wrap(err, "could not retrieve a resource")
	}
	if err := a.rm.Create(context.Background(), signingKey, core_store.CreateBy(key)); err != nil {
		return "", errors.Wrap(err, "could not create a resource")
	}
	return signingKey.Spec.GetData().String(), nil
}
