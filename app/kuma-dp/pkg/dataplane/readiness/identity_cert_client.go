package readiness

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	envoy_admin "github.com/envoyproxy/go-control-plane/envoy/admin/v3"
	envoy_tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/golang/protobuf/jsonpb"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/version"
	"io"
	"net/http"
	"time"
)

const (
	// default timeout for the readiness/liveness probe is 3 seconds, so we use a shorter timeout here (set by SidecarReadinessProbe.TimeoutSeconds)
	defaultTimeout   = 2 * time.Second
	identityCertPath = "/config_dump?resource=dynamic_active_secrets&name_regex=identity_cert:secret:.*"
	uninitialized    = "uninitialized"
)

type IdentityCertClient struct {
	EnvoyAdminAddress string
	EnvoyAdminPort    uint32
	HttpClient        *http.Client
}

func (c *IdentityCertClient) CheckIfReady() (bool, error) {
	if c.EnvoyAdminPort == 0 {
		return true, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	r, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("http://%s:%d%s", c.EnvoyAdminAddress, c.EnvoyAdminPort, identityCertPath), http.NoBody)
	if err != nil {
		return false, fmt.Errorf("failed to build a request: %w", err)
	}
	r.Header.Set("User-Agent", fmt.Sprintf("kuma-dp/%s", version.Build.Version))
	response, err := c.HttpClient.Do(r)
	if err != nil {
		return false, fmt.Errorf("could not request envoy: %w", err)
	}

	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return false, fmt.Errorf("envoy responded a failure %s(%d): %s", response.Status, response.StatusCode, response.Body)
	}

	resp, err := io.ReadAll(response.Body)
	if err != nil {
		return false, fmt.Errorf("could not read envoy response: %w", err)
	}
	return certSecretReady(resp)
}

func certSecretReady(resp []byte) (bool, error) {
	if len(resp) <= 2 {
		// "{}", meaning no identity cert is found because the mesh has not enabled mTLS.
		return true, nil
	}

	configDump := &envoy_admin.ConfigDump{}
	err := jsonpb.Unmarshal(bytes.NewReader(resp), configDump)
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal envoy config dump: %w", err)
	}

	switch len(configDump.Configs) {
	case 0:
		// no identity cert is found, can't check at this time, don't fail the readiness
		return true, nil
	default:
		return false, fmt.Errorf("more than one identity cert secrets are found from config dump")
	case 1:
		secret := &envoy_admin.SecretsConfigDump_DynamicSecret{VersionInfo: uninitialized}
		err = util_proto.UnmarshalAnyTo(configDump.Configs[0], secret)
		if err != nil {
			return false, fmt.Errorf("failed to unmarshal dynamic secret: %w", err)
		}

		if secret.GetVersionInfo() == "" || secret.GetVersionInfo() == uninitialized || secret.GetSecret() == nil {
			return false, nil
		}

		tlsSecret := &envoy_tls.Secret{}
		err = util_proto.UnmarshalAnyTo(secret.GetSecret(), tlsSecret)
		if err != nil {
			return false, fmt.Errorf("failed to unmarshal tls secret: %w", err)
		}

		pemBytes := tlsSecret.GetTlsCertificate().GetCertificateChain().GetInlineBytes()
		pemBlock, _ := pem.Decode(pemBytes)
		if pemBlock == nil {
			return false, fmt.Errorf("invalid PEM block: %w", err)
		}
		var x509Cert *x509.Certificate
		x509Cert, err = x509.ParseCertificate(pemBlock.Bytes)
		return x509Cert != nil && x509Cert.NotAfter.After(time.Now().UTC()), err
	}
}
