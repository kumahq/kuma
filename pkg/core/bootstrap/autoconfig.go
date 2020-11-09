package bootstrap

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/config/api-server/catalog"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	config_core "github.com/kumahq/kuma/pkg/config/core"
	dp_server "github.com/kumahq/kuma/pkg/config/dp-server"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/tls"
)

var autoconfigureLog = core.Log.WithName("bootstrap").WithName("auto-configure")

func autoconfigure(cfg *kuma_cp.Config) error {
	autoconfigureDpServerAuth(cfg)
	if err := autoconfigureTLS(cfg); err != nil {
		return err
	}
	autoconfigureServersTLS(cfg)
	autoconfigureCatalog(cfg)
	autoconfigBootstrapXdsParams(cfg)
	return nil
}

func autoconfigureDpServerAuth(cfg *kuma_cp.Config) {
	if cfg.DpServer.Auth.Type == "" {
		switch cfg.Environment {
		case config_core.KubernetesEnvironment:
			cfg.DpServer.Auth.Type = dp_server.DpServerAuthServiceAccountToken
		case config_core.UniversalEnvironment:
			cfg.DpServer.Auth.Type = dp_server.DpServerAuthDpToken
		}
	}
}

func autoconfigureServersTLS(cfg *kuma_cp.Config) {
	if cfg.Multizone.Global.KDS.TlsCertFile == "" {
		cfg.Multizone.Global.KDS.TlsCertFile = cfg.General.TlsCertFile
		cfg.Multizone.Global.KDS.TlsKeyFile = cfg.General.TlsKeyFile
	}
	if cfg.DpServer.TlsCertFile == "" {
		cfg.DpServer.TlsCertFile = cfg.General.TlsCertFile
		cfg.DpServer.TlsKeyFile = cfg.General.TlsKeyFile
	}
	if cfg.ApiServer.HTTPS.TlsCertFile == "" {
		cfg.ApiServer.HTTPS.TlsCertFile = cfg.General.TlsCertFile
		cfg.ApiServer.HTTPS.TlsKeyFile = cfg.General.TlsKeyFile
	}
}

func autoconfigureTLS(cfg *kuma_cp.Config) error {
	if cfg.General.TlsCertFile == "" {
		hosts := []string{
			cfg.General.AdvertisedHostname,
			"localhost",
		}
		cert, err := tls.NewSelfSignedCert("kuma-control-plane", tls.ServerCertType, hosts...)
		if err != nil {
			return errors.Wrap(err, "failed to auto-generate TLS certificate")
		}
		crtFile, keyFile, err := saveKeyPair(cert)
		if err != nil {
			return errors.Wrap(err, "failed to save auto-generated TLS certificate")
		}
		cfg.General.TlsCertFile = crtFile
		cfg.General.TlsKeyFile = keyFile
		autoconfigureLog.Info("auto-generated TLS certificate for Kuma Control Plane", "crtFile", crtFile, "keyFile", keyFile)
	}
	return nil
}

func autoconfigureCatalog(cfg *kuma_cp.Config) {
	bootstrapUrl := cfg.ApiServer.Catalog.Bootstrap.Url
	if len(bootstrapUrl) == 0 {
		bootstrapUrl = fmt.Sprintf("https://%s:%d", cfg.General.AdvertisedHostname, cfg.DpServer.Port)
	}
	madsUrl := cfg.ApiServer.Catalog.MonitoringAssignment.Url
	if len(madsUrl) == 0 {
		madsUrl = fmt.Sprintf("grpc://%s:%d", cfg.General.AdvertisedHostname, cfg.MonitoringAssignmentServer.GrpcPort)
	}
	cat := &catalog.CatalogConfig{
		ApiServer: catalog.ApiServerConfig{
			Url: fmt.Sprintf("http://%s:%d", cfg.General.AdvertisedHostname, cfg.ApiServer.HTTP.Port),
		},
		Bootstrap: catalog.BootstrapApiConfig{
			Url: bootstrapUrl,
		},
		MonitoringAssignment: catalog.MonitoringAssignmentApiConfig{
			Url: madsUrl,
		},
		Sds: catalog.SdsApiConfig{
			Url: cfg.ApiServer.Catalog.Sds.Url,
		},
	}
	if cfg.DpServer.Auth.Type == dp_server.DpServerAuthDpToken {
		cat.DataplaneToken.LocalUrl = fmt.Sprintf("http://localhost:%d", cfg.ApiServer.HTTP.Port)
	}
	cfg.ApiServer.Catalog = cat
}

func autoconfigBootstrapXdsParams(cfg *kuma_cp.Config) {
	if cfg.BootstrapServer.Params.XdsHost == "" {
		cfg.BootstrapServer.Params.XdsHost = cfg.General.AdvertisedHostname
	}
	if cfg.BootstrapServer.Params.XdsPort == 0 {
		cfg.BootstrapServer.Params.XdsPort = uint32(cfg.DpServer.Port)
	}
}

func saveKeyPair(pair tls.KeyPair) (string, string, error) {
	crtFile, err := ioutil.TempFile("", "*.crt")
	if err != nil {
		return "", "", errors.Wrap(err, "failed to create a temp file with TLS cert")
	}
	if err := ioutil.WriteFile(crtFile.Name(), pair.CertPEM, os.ModeTemporary); err != nil {
		return "", "", errors.Wrapf(err, "failed to save TLS cert into a temp file %q", crtFile.Name())
	}

	keyFile, err := ioutil.TempFile("", "*.key")
	if err != nil {
		return "", "", errors.Wrap(err, "failed to create a temp file with TLS key")
	}
	if err := ioutil.WriteFile(keyFile.Name(), pair.KeyPEM, os.ModeTemporary); err != nil {
		return "", "", errors.Wrapf(err, "failed to save TLS key into a temp file %q", keyFile.Name())
	}

	return crtFile.Name(), keyFile.Name(), nil
}
