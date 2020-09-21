package bootstrap

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/config/api-server/catalog"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/tls"
)

var autoconfigureLog = core.Log.WithName("bootstrap").WithName("auto-configure")

func autoconfigure(cfg *kuma_cp.Config) error {
	autoconfigureAdminServer(cfg)
	autoconfigureCatalog(cfg)
	autoconfigBootstrapXdsParams(cfg)
	if err := autoconfigureKds(cfg); err != nil {
		return err
	}
	return autoconfigureSds(cfg)
}

func autoconfigureCatalog(cfg *kuma_cp.Config) {
	bootstrapUrl := cfg.ApiServer.Catalog.Bootstrap.Url
	if len(bootstrapUrl) == 0 {
		bootstrapUrl = fmt.Sprintf("http://%s:%d", cfg.General.AdvertisedHostname, cfg.BootstrapServer.Port)
	}
	madsUrl := cfg.ApiServer.Catalog.MonitoringAssignment.Url
	if len(madsUrl) == 0 {
		madsUrl = fmt.Sprintf("grpc://%s:%d", cfg.General.AdvertisedHostname, cfg.MonitoringAssignmentServer.GrpcPort)
	}
	cat := &catalog.CatalogConfig{
		ApiServer: catalog.ApiServerConfig{
			Url: fmt.Sprintf("http://%s:%d", cfg.General.AdvertisedHostname, cfg.ApiServer.Port),
		},
		Bootstrap: catalog.BootstrapApiConfig{
			Url: bootstrapUrl,
		},
		Admin: catalog.AdminApiConfig{
			LocalUrl: fmt.Sprintf("http://localhost:%d", cfg.AdminServer.Local.Port),
		},
		MonitoringAssignment: catalog.MonitoringAssignmentApiConfig{
			Url: madsUrl,
		},
		Sds: catalog.SdsApiConfig{
			Url: cfg.ApiServer.Catalog.Sds.Url,
		},
	}
	if cfg.AdminServer.Public.Enabled {
		cat.Admin.PublicUrl = fmt.Sprintf("https://%s:%d", cfg.General.AdvertisedHostname, cfg.AdminServer.Public.Port)
	}
	if cfg.AdminServer.Apis.DataplaneToken.Enabled {
		cat.DataplaneToken.LocalUrl = fmt.Sprintf("http://localhost:%d", cfg.AdminServer.Local.Port)
		if cfg.AdminServer.Public.Enabled {
			cat.DataplaneToken.PublicUrl = fmt.Sprintf("https://%s:%d", cfg.General.AdvertisedHostname, cfg.AdminServer.Public.Port)
		}
	}
	cfg.ApiServer.Catalog = cat
}

func autoconfigureSds(cfg *kuma_cp.Config) error {
	// to improve UX, we want to auto-generate TLS cert for SDS if possible
	if cfg.Environment == config_core.UniversalEnvironment {
		if cfg.SdsServer.TlsCertFile == "" {
			var sdsHost = ""
			if cfg.ApiServer.Catalog.Sds.Url != "" {
				u, err := url.Parse(cfg.ApiServer.Catalog.Sds.Url)
				if err != nil {
					return errors.Wrap(err, "sds url is malformed")
				}
				sdsHost = strings.Split(u.Host, ":")[0]
			}
			if len(sdsHost) == 0 {
				sdsHost = cfg.BootstrapServer.Params.XdsHost
			}
			hosts := []string{
				sdsHost,
				"localhost",
			}
			// notice that Envoy's SDS client (Google gRPC) does require DNS SAN in a X509 cert of an SDS server
			sdsCert, err := tls.NewSelfSignedCert("kuma-sds", tls.ServerCertType, hosts...)
			if err != nil {
				return errors.Wrap(err, "failed to auto-generate TLS certificate for SDS server")
			}
			crtFile, keyFile, err := saveKeyPair(sdsCert)
			if err != nil {
				return errors.Wrap(err, "failed to save auto-generated TLS certificate for SDS server")
			}
			cfg.SdsServer.TlsCertFile = crtFile
			cfg.SdsServer.TlsKeyFile = keyFile

			autoconfigureLog.Info("auto-generated TLS certificate for SDS server", "crtFile", crtFile, "keyFile", keyFile)
		}
	}
	return nil
}

func autoconfigureKds(cfg *kuma_cp.Config) error {
	// to improve UX, we want to auto-generate TLS cert for KDS if possible
	if cfg.Environment == config_core.UniversalEnvironment {
		if cfg.Multicluster.Global.KDS.TlsCertFile == "" {
			hosts := []string{}
			kdsCert, err := tls.NewSelfSignedCert("kuma-kds", tls.ServerCertType, hosts...)
			if err != nil {
				return errors.Wrap(err, "failed to auto-generate TLS certificate for KDS server")
			}
			crtFile, keyFile, err := saveKeyPair(kdsCert)
			if err != nil {
				return errors.Wrap(err, "failed to save auto-generated TLS certificate for KDS server")
			}
			cfg.Multicluster.Global.KDS.TlsCertFile = crtFile
			cfg.Multicluster.Global.KDS.TlsKeyFile = keyFile

			autoconfigureLog.Info("auto-generated TLS certificate for KDS server", "crtFile", crtFile, "keyFile", keyFile)
		}
	}
	return nil
}

func autoconfigureAdminServer(cfg *kuma_cp.Config) {
	if cfg.AdminServer.Public.Enabled && cfg.AdminServer.Public.Port == 0 {
		cfg.AdminServer.Public.Port = cfg.AdminServer.Local.Port
	}
}

func autoconfigBootstrapXdsParams(cfg *kuma_cp.Config) {
	if cfg.BootstrapServer.Params.XdsHost == "" {
		cfg.BootstrapServer.Params.XdsHost = cfg.General.AdvertisedHostname
	}
	if cfg.BootstrapServer.Params.XdsPort == 0 {
		cfg.BootstrapServer.Params.XdsPort = uint32(cfg.XdsServer.GrpcPort)
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
