package admin

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/tokens/builtin/issuer"
)

type EnvoyAdmin interface {
	GenerateAPIToken(dataplane *mesh_core.DataplaneResource) (string, error)
	PostQuit(dataplane *mesh_core.DataplaneResource) error
}

type envoyAdmin struct {
	rm  manager.ResourceManager
	cfg kuma_cp.Config
}

func NewEnvoyAdmin(rm manager.ResourceManager, cfg kuma_cp.Config) EnvoyAdmin {
	return &envoyAdmin{
		rm:  rm,
		cfg: cfg,
	}
}

const (
	quitquitquit = "quitquitquit"
)

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

func (a *envoyAdmin) adminAddress(dataplane *mesh_core.DataplaneResource) string {
	ip := dataplane.GetIP()
	portUint := a.cfg.Runtime.Kubernetes.Injector.SidecarContainer.AdminPort

	return fmt.Sprintf("%s:%d", ip, portUint)
}

func (a *envoyAdmin) PostQuit(dataplane *mesh_core.DataplaneResource) error {
	request := &http.Request{
		Method: "POST",
		URL: &url.URL{
			Scheme: "http",
			Host:   a.adminAddress(dataplane),
			Path:   quitquitquit,
		},
		Header: nil,
	}

	// Envoy will not send back any response, so do we not check the response
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return errors.Wrapf(err, "unable to send POST to %s", quitquitquit)
	}
	defer response.Body.Close()

	return nil
}
