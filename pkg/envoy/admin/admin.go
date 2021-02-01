package admin

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/kumahq/kuma/pkg/core"

	"github.com/kumahq/kuma/pkg/core/resources/apis/system"

	"github.com/pkg/errors"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/tokens/builtin/issuer"
)

const envoyAdminClientTokenPrefix = "envoy-admin-client-token"

var log = core.Log.WithName("envoy-admin")

type EnvoyAdminClient interface {
	GenerateAPIToken(dataplane *mesh_core.DataplaneResource) (string, error)
	PostQuit(dataplane *mesh_core.DataplaneResource) error
}

type envoyAdminClient struct {
	rm         manager.ResourceManager
	cfg        kuma_cp.Config
	httpClient *http.Client
	scheme     string
}

func NewEnvoyAdminClient(rm manager.ResourceManager, cfg kuma_cp.Config) EnvoyAdminClient {
	return &envoyAdminClient{
		rm:  rm,
		cfg: cfg,
		httpClient: &http.Client{
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout: 3 * time.Second,
				}).Dial,
				TLSHandshakeTimeout: 3 * time.Second,
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
			Timeout: 5 * time.Second,
		},
		scheme: "https",
	}
}

const (
	quitquitquit = "quitquitquit"
)

func (a *envoyAdminClient) GenerateAPIToken(dataplane *mesh_core.DataplaneResource) (string, error) {
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

func (a *envoyAdminClient) getSigningKeyString(mesh string) (string, error) {
	signingKey := system.NewSecretResource()
	key := issuer.SigningKeyResourceKey(envoyAdminClientTokenPrefix, mesh)
	err := a.rm.Get(context.Background(), signingKey, core_store.GetBy(key))
	if err == nil {
		return signingKey.Spec.GetData().String(), nil
	}
	if !core_store.IsResourceNotFound(err) {
		return "", errors.Wrap(err, "could not retrieve a resource")
	}

	// Key not found, create it
	signingKey, err = issuer.CreateSigningKey()
	if err != nil {
		return "", errors.Wrap(err, "could not create a signing key")
	}
	if err := a.rm.Create(context.Background(), signingKey, core_store.CreateBy(key)); err != nil {
		return "", errors.Wrap(err, "could not create a resource")
	}
	return signingKey.Spec.GetData().String(), nil
}

func (a *envoyAdminClient) adminAddress(dataplane *mesh_core.DataplaneResource) string {
	ip := dataplane.GetIP()
	portUint := a.cfg.Runtime.Kubernetes.Injector.SidecarContainer.AdminPort

	return fmt.Sprintf("%s:%d", ip, portUint)
}

func (a *envoyAdminClient) PostQuit(dataplane *mesh_core.DataplaneResource) error {
	token, err := a.GenerateAPIToken(dataplane)
	if err != nil {
		return err
	}

	body := bytes.NewReader([]byte(""))
	url := fmt.Sprintf("%s://%s/%s", a.scheme, a.adminAddress(dataplane), quitquitquit)

	request, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}

	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	log.V(1).Info("Sendig", "Method", request.Method, "URL", request.URL, "Header", request.Header)

	// Envoy will not send back any response, so do we not check the response
	response, err := a.httpClient.Do(request)
	if err != nil {
		return errors.Wrapf(err, "unable to send POST to %s", quitquitquit)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return errors.Errorf("Envoy response [%d %s] [%s]", response.StatusCode, response.Status, response.Body)
	}

	return nil
}
