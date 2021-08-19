package admin

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/tokens/builtin/issuer"
)

type EnvoyAdminClient interface {
	GenerateAPIToken(dataplane *core_mesh.DataplaneResource) (string, error)
	PostQuit(dataplane *core_mesh.DataplaneResource) error
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
					// we do not want to explicitly verify the https
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

func (a *envoyAdminClient) GenerateAPIToken(dataplane *core_mesh.DataplaneResource) (string, error) {
	mesh := dataplane.Meta.GetMesh()
	key, err := a.getOrCreateSigningKey(mesh)
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

func (a *envoyAdminClient) getOrCreateSigningKey(mesh string) (string, error) {
	key, err := issuer.GetSigningKey(a.rm, issuer.DataplaneTokenPrefix, mesh)
	if err != nil {
		return "", errors.Wrap(err, "unable to retrieve the signing key")
	}

	return string(key), nil
}

func (a *envoyAdminClient) adminAddress(dataplane *core_mesh.DataplaneResource) string {
	ip := dataplane.GetIP()
	// TODO: this will work perfectly fine with K8s, but will fail for Universal
	// The real allocated admin port is part of the DP metadata, but it is attached to a particular CP,
	// so we can not reliably use that. A better approach would be to include the admin port
	// in the DataplaneInsights.
	portUint := a.cfg.Runtime.Kubernetes.Injector.SidecarContainer.AdminPort

	return net.JoinHostPort(ip, strconv.FormatUint(uint64(portUint), 10))
}

func (a *envoyAdminClient) PostQuit(dataplane *core_mesh.DataplaneResource) error {
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

	// Envoy will not send back any response, so do we not check the response
	response, err := a.httpClient.Do(request)
	if err != nil {
		return errors.Wrapf(err, "unable to send POST to %s", quitquitquit)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return errors.Errorf("envoy response [%d %s] [%s]", response.StatusCode, response.Status, response.Body)
	}

	return nil
}
