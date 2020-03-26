package vault

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/go-retryablehttp"
	vault "github.com/hashicorp/vault/api"
	"golang.org/x/net/http2"
)

type Config struct {
	Address      string
	AgentAddress string
	Token        string
	Namespace    string
	Tls          TLSConfig
}

// Those values are from vault.DefaultConfig. We didn't use it method so we don't set the config with default Vault ENV VARs, but we use Kuma config instead.
func (v *Config) toVaultConfig() *vault.Config {
	cfg := &vault.Config{
		Address:      v.Address,
		AgentAddress: v.AgentAddress,
		HttpClient:   cleanhttp.DefaultPooledClient(),
		Timeout:      time.Second * 60,
		MaxRetries:   2,
		Backoff:      retryablehttp.LinearJitterBackoff,
	}

	transport := cfg.HttpClient.Transport.(*http.Transport)
	transport.TLSHandshakeTimeout = 10 * time.Second
	transport.TLSClientConfig = &tls.Config{
		MinVersion: tls.VersionTLS12,
	}
	if err := http2.ConfigureTransport(transport); err != nil {
		cfg.Error = err
		return cfg
	}

	cfg.HttpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	cfg.Error = cfg.ConfigureTLS(v.Tls.toVaultConfig())
	return cfg
}

type TLSConfig struct {
	CaCertPath     string
	CaCertDir      string
	ClientCertPath string
	ClientKeyPath  string
	SkipVerify     bool
	ServerName     string
}

func (v *TLSConfig) toVaultConfig() *vault.TLSConfig {
	return &vault.TLSConfig{
		CACert:        v.CaCertPath,
		CAPath:        v.CaCertDir,
		ClientCert:    v.ClientCertPath,
		ClientKey:     v.ClientKeyPath,
		TLSServerName: v.ServerName,
		Insecure:      v.SkipVerify,
	}
}

func NewVaultClient(config Config) (*vault.Client, error) {
	client, err := vault.NewClient(config.toVaultConfig())
	if err != nil {
		return nil, err
	}
	client.SetToken(config.Token)
	client.SetNamespace(config.Namespace)
	return client, err
}
