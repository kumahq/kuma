package transparentproxy

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/config"
)

var _ config.Config = TransparentProxyConfig{}

type TransparentProxyConfig struct {
	config.BaseConfig

	KumactlLinuxBin    string            `json:"kumactlLinuxBin,omitempty" envconfig:"KUMACTL_LINUX_BIN"`
	DockerImagesToTest map[string]string `json:"dockerImagesToTest,omitempty" envconfig:"DOCKER_IMAGES_TO_TEST"`
	InstallFlagsToTest []string          `json:"additionalFlagsToTest,omitempty" envconfig:"ADDITIONAL_FLAGS_TO_TEST"`
	IPV6               bool              `json:"ipv6,omitempty" envconfig:"IPV6"`
}

func (c TransparentProxyConfig) Validate() error {
	if Config.KumactlLinuxBin != "" {
		_, err := os.Stat(Config.KumactlLinuxBin)
		if os.IsNotExist(err) {
			return errors.Wrapf(
				err,
				"unable to find kumactl for linux at: %s",
				Config.KumactlLinuxBin,
			)
		}

		return err
	}

	return nil
}

func (c TransparentProxyConfig) AutoConfigure() error {
	absoluteKumactlPath, err := filepath.Abs(Config.KumactlLinuxBin)
	if err != nil {
		return err
	}

	Config.KumactlLinuxBin = absoluteKumactlPath

	return nil
}

var Config TransparentProxyConfig

var defaultConfig = TransparentProxyConfig{
	KumactlLinuxBin: fmt.Sprintf(
		"../../../build/artifacts-linux-%s/kumactl/kumactl",
		runtime.GOARCH,
	),
	DockerImagesToTest: map[string]string{
		"Ubuntu 24.04":      "ubuntu:noble-20240605",
		"Ubuntu 22.04":      "ubuntu:jammy-20240530",
		"Ubuntu 20.04":      "ubuntu:focal-20240530",
		"Debian 12":         "debian:bookworm-20240612",
		"Debian 11":         "debian:bullseye-20240612",
		"Debian 10":         "debian:buster-20240612",
		"RHEL 9":            "redhat/ubi9:9.4-1123",
		"Alpine 3":          "alpine:3.20.1",
		"Amazon Linux 2023": "amazonlinux:2023.4.20240611.0",
		"Amazon Linux 2":    "amazonlinux:2.0.20240610.1",
		// Skipping RHEL 8 as our transparent proxy currently relies on
		// iptables-nft or iptables-legacy binaries. RHEL 8 only provides the
		// base iptables binary. Unpause these entries to include RHEL 8 once
		// out transparent proxy is fixed to support base iptables binaries.
		// "RHEL 8":            "redhat/ubi8:8.10-901.1717584420",
	},
	InstallFlagsToTest: []string{
		"--redirect-all-dns-traffic",
	},
	IPV6: false,
}

func InitConfig() {
	Config = defaultConfig

	if err := config.Load(os.Getenv("TPROXY_TESTS_CONFIG_FILE"), &Config); err != nil {
		panic(err)
	}

	if err := Config.AutoConfigure(); err != nil {
		panic(err)
	}
}
