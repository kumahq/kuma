package bootstrap

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	config_core "github.com/kumahq/kuma/pkg/config/core"
	dp_server "github.com/kumahq/kuma/pkg/config/dp-server"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/tls"
	util_net "github.com/kumahq/kuma/pkg/util/net"
)

const (
	crtFileName = "kuma-cp.crt"
	keyFileName = "kuma-cp.key"
)

var autoconfigureLog = core.Log.WithName("bootstrap").WithName("auto-configure")

func autoconfigure(cfg *kuma_cp.Config) error {
	if err := autoconfigureGeneral(cfg); err != nil {
		return err
	}
	autoconfigureDpServerAuth(cfg)
	if err := autoconfigureTLS(cfg); err != nil {
		return errors.Wrap(err, "could not autogenerate TLS certificate")
	}
	autoconfigureServersTLS(cfg)
	autoconfigBootstrapXdsParams(cfg)
	if err := autoconfigureInterCp(cfg); err != nil {
		return errors.Wrap(err, "could not autoconfigure Inter CP config")
	}
	return nil
}

func autoconfigureGeneral(cfg *kuma_cp.Config) error {
	if cfg.General.WorkDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return errors.Errorf("failed to create a working directory inside $HOME: %v, "+
				"please pick a working directory by setting KUMA_GENERAL_WORK_DIR manually", err)
		}
		cfg.General.WorkDir = path.Join(home, ".kuma")
	}
	return nil
}

func autoconfigureDpServerAuth(cfg *kuma_cp.Config) {
	if cfg.DpServer.Authn.DpProxy.Type == "" {
		switch cfg.Environment {
		case config_core.KubernetesEnvironment:
			cfg.DpServer.Authn.DpProxy.Type = dp_server.DpServerAuthServiceAccountToken
		case config_core.UniversalEnvironment:
			cfg.DpServer.Authn.DpProxy.Type = dp_server.DpServerAuthDpToken
		}
	}
	if cfg.DpServer.Authn.ZoneProxy.Type == "" {
		switch cfg.Environment {
		case config_core.KubernetesEnvironment:
			cfg.DpServer.Authn.ZoneProxy.Type = dp_server.DpServerAuthServiceAccountToken
		case config_core.UniversalEnvironment:
			cfg.DpServer.Authn.ZoneProxy.Type = dp_server.DpServerAuthZoneToken
		}
	}
}

func autoconfigureServersTLS(cfg *kuma_cp.Config) {
	if cfg.Multizone.Global.KDS.TlsCertFile == "" {
		cfg.Multizone.Global.KDS.TlsCertFile = cfg.General.TlsCertFile
		cfg.Multizone.Global.KDS.TlsKeyFile = cfg.General.TlsKeyFile
	}
	if cfg.Diagnostics.TlsCertFile == "" {
		cfg.Diagnostics.TlsCertFile = cfg.General.TlsCertFile
		cfg.Diagnostics.TlsKeyFile = cfg.General.TlsKeyFile
	}
	if cfg.DpServer.TlsCertFile == "" {
		cfg.DpServer.TlsCertFile = cfg.General.TlsCertFile
		cfg.DpServer.TlsKeyFile = cfg.General.TlsKeyFile
	}
	if cfg.ApiServer.HTTPS.TlsCertFile == "" {
		cfg.ApiServer.HTTPS.TlsCertFile = cfg.General.TlsCertFile
		cfg.ApiServer.HTTPS.TlsKeyFile = cfg.General.TlsKeyFile
	}
	if cfg.MonitoringAssignmentServer.TlsCertFile == "" {
		cfg.MonitoringAssignmentServer.TlsCertFile = cfg.General.TlsCertFile
		cfg.MonitoringAssignmentServer.TlsKeyFile = cfg.General.TlsKeyFile
	}
	if cfg.General.TlsMinVersion != "" {
		cfg.Diagnostics.TlsMinVersion = cfg.General.TlsMinVersion
		cfg.Multizone.Global.KDS.TlsMinVersion = cfg.General.TlsMinVersion
		cfg.DpServer.TlsMinVersion = cfg.General.TlsMinVersion
		cfg.ApiServer.HTTPS.TlsMinVersion = cfg.General.TlsMinVersion
		cfg.MonitoringAssignmentServer.TlsMinVersion = cfg.General.TlsMinVersion
		cfg.InterCp.Server.TlsMinVersion = cfg.General.TlsMinVersion
	}
	if cfg.General.TlsMaxVersion != "" {
		cfg.Diagnostics.TlsMaxVersion = cfg.General.TlsMaxVersion
		cfg.Multizone.Global.KDS.TlsMaxVersion = cfg.General.TlsMaxVersion
		cfg.DpServer.TlsMaxVersion = cfg.General.TlsMaxVersion
		cfg.ApiServer.HTTPS.TlsMaxVersion = cfg.General.TlsMaxVersion
		cfg.MonitoringAssignmentServer.TlsMaxVersion = cfg.General.TlsMaxVersion
		cfg.InterCp.Server.TlsMaxVersion = cfg.General.TlsMaxVersion
	}
	if len(cfg.General.TlsCipherSuites) > 0 {
		cfg.Diagnostics.TlsCipherSuites = cfg.General.TlsCipherSuites
		cfg.Multizone.Global.KDS.TlsCipherSuites = cfg.General.TlsCipherSuites
		cfg.DpServer.TlsCipherSuites = cfg.General.TlsCipherSuites
		cfg.ApiServer.HTTPS.TlsCipherSuites = cfg.General.TlsCipherSuites
		cfg.MonitoringAssignmentServer.TlsCipherSuites = cfg.General.TlsCipherSuites
		cfg.InterCp.Server.TlsCipherSuites = cfg.General.TlsCipherSuites
	}
}

func autoconfigureTLS(cfg *kuma_cp.Config) error {
	if cfg.General.TlsCertFile != "" {
		return nil
	}
	autoconfigureLog.Info(fmt.Sprintf("directory %v will be used as a working directory, "+
		"it could be changed using KUMA_GENERAL_WORK_DIR environment variable", cfg.General.WorkDir))

	if crtFile, keyFile, err := tryReadKeyPair(workDir(cfg.General.WorkDir)); err == nil {
		cfg.General.TlsCertFile = crtFile
		cfg.General.TlsKeyFile = keyFile
		autoconfigureLog.Info("Kuma detected TLS cert and key in the working directory")
		return nil
	}

	ips, err := util_net.GetAllIPs()
	if err != nil {
		return errors.Wrap(err, "could not list all IPs of the machine")
	}
	hostname, err := os.Hostname()
	if err != nil {
		return errors.Wrap(err, "could not get a hostname of the machine")
	}
	hosts := append([]string{hostname, "localhost"}, ips...)
	cert, err := tls.NewSelfSignedCert(tls.ServerCertType, tls.DefaultKeyType, hosts...)
	if err != nil {
		return errors.Wrap(err, "failed to auto-generate TLS certificate")
	}
	crtFile, keyFile, err := saveKeyPair(cert, workDir(cfg.General.WorkDir))
	if err != nil {
		return errors.Errorf("failed to save auto-generated TLS cert and key into a working directory: %v, "+
			"working directory could be changed using KUMA_GENERAL_WORK_DIR environment variable", err)
	}
	cfg.General.TlsCertFile = crtFile
	cfg.General.TlsKeyFile = keyFile
	autoconfigureLog.Info("TLS certificate autogenerated. Autogenerated certificates are not synchronized between CP instances. It is only valid if the data plane proxy connects to the CP by one of the following address "+strings.Join(hosts, ", ")+
		". It is recommended to generate your own certificate based on yours trusted CA. You can also generate your own self-signed certificates using 'kumactl generate tls-certificate --type=server --hostname=<hostname>' and configure them using KUMA_GENERAL_TLS_CERT_FILE and KUMA_GENERAL_TLS_KEY_FILE", "crtFile", crtFile, "keyFile", keyFile)
	return nil
}

func autoconfigBootstrapXdsParams(cfg *kuma_cp.Config) {
	if cfg.BootstrapServer.Params.XdsPort == 0 {
		cfg.BootstrapServer.Params.XdsPort = uint32(cfg.DpServer.Port)
	}
}

type workDir string

func (w workDir) Open(name string) (*os.File, error) {
	if err := os.MkdirAll(string(w), 0o700); err != nil && !os.IsExist(err) {
		return nil, err
	}
	return os.OpenFile(path.Join(string(w), name), os.O_RDWR|os.O_CREATE, 0o600)
}

func tryReadKeyPair(dir workDir) (string, string, error) {
	crtFile, err := dir.Open(crtFileName)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to open a file with TLS cert")
	}
	defer func() {
		if err := crtFile.Close(); err != nil {
			autoconfigureLog.Error(err, "failed to close TLS cert file")
		}
	}()

	certPEM, err := os.ReadFile(crtFile.Name())
	if err != nil {
		return "", "", err
	}
	if len(certPEM) == 0 {
		return "", "", errors.New("file with TLS cert is empty")
	}
	keyFile, err := dir.Open(keyFileName)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to open a file with TLS key")
	}
	defer func() {
		if err := keyFile.Close(); err != nil {
			autoconfigureLog.Error(err, "failed to close TLS key file")
		}
	}()
	keyPEM, err := os.ReadFile(keyFile.Name())
	if err != nil {
		return "", "", err
	}
	if len(keyPEM) == 0 {
		return "", "", errors.New("file with TLS cert is empty")
	}
	return crtFile.Name(), keyFile.Name(), nil
}

func saveKeyPair(pair tls.KeyPair, dir workDir) (string, string, error) {
	crtFile, err := dir.Open(crtFileName)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to create a file with TLS cert")
	}
	defer func() {
		if err := crtFile.Close(); err != nil {
			autoconfigureLog.Error(err, "failed to close TLS cert file")
		}
	}()
	if err := os.WriteFile(crtFile.Name(), pair.CertPEM, os.ModeTemporary); err != nil {
		return "", "", errors.Wrapf(err, "failed to save TLS cert into a file %q", crtFile.Name())
	}

	keyFile, err := dir.Open(keyFileName)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to create a file with TLS key")
	}
	defer func() {
		if err := keyFile.Close(); err != nil {
			autoconfigureLog.Error(err, "failed to close TLS key file")
		}
	}()
	if err := os.WriteFile(keyFile.Name(), pair.KeyPEM, os.ModeTemporary); err != nil {
		return "", "", errors.Wrapf(err, "failed to save TLS key into a file %q", keyFile.Name())
	}

	return crtFile.Name(), keyFile.Name(), nil
}

func autoconfigureInterCp(cfg *kuma_cp.Config) error {
	if cfg.InterCp.Catalog.InstanceAddress != "" {
		return nil
	}
	ips, err := util_net.GetAllIPs(util_net.NonLoopback)
	if err != nil {
		return errors.Wrap(err, "could not list all IPs of the machine")
	}
	if len(ips) == 0 {
		return errors.New("there is 0 non-loopback interfaces on the machine. Set KUMA_INTER_CP_CATALOG_INSTANCE_ADDRESS explicitly.")
	}
	if len(ips) > 1 {
		log.Info("there are multiple non-loopback interfaces on the machine. It is recommended to set KUMA_INTER_CP_CATALOG_INSTANCE_ADDRESS explicitly to set IP on which other control plane instances in the cluster can communicate with this instance.")
	}
	cfg.InterCp.Catalog.InstanceAddress = ips[0]
	return nil
}
