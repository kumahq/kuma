package transparentproxy

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/config"
)

var _ config.Config = TransparentProxyConfig{}

type Params struct {
	Flags     []string
	EchoStdin string
	Files     map[string]string
}

type ParamsMap map[string]Params

func (f *ParamsMap) Decode(value string) error {
	result := map[string]Params{}

	if err := json.Unmarshal([]byte(value), &result); err != nil {
		return err
	}

	*f = result

	return nil
}

type TransparentProxyConfig struct {
	config.BaseConfig

	KumactlLinuxBin     string            `json:"kumactlLinuxBin,omitempty" envconfig:"KUMACTL_LINUX_BIN"`
	DockerImagesToTest  map[string]string `json:"dockerImagesToTest,omitempty" envconfig:"DOCKER_IMAGES_TO_TEST"`
	InstallParamsToTest *ParamsMap        `json:"additionalParamsToTest,omitempty" envconfig:"ADDITIONAL_PARAMS_TO_TEST"`
	IPV6                bool              `json:"ipv6,omitempty" envconfig:"IPV6"`
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
		"Ubuntu 18.04":      "ubuntu:bionic-20230530",
		"Debian 12":         "debian:bookworm-20240612",
		"Debian 11":         "debian:bullseye-20240612",
		"Debian 10":         "debian:buster-20240612",
		"RHEL 9":            "redhat/ubi9:9.4-1123",
		"RHEL 8":            "redhat/ubi8:8.10-901.1717584420",
		"Alpine 3":          "alpine:3.20.1",
		"Amazon Linux 2023": "amazonlinux:2023.4.20240611.0",
		"Amazon Linux 2":    "amazonlinux:2.0.20240610.1",
		"Fedora 41":         "fedora:41",
		"Fedora 40":         "fedora:40",
		"Fedora 39":         "fedora:39",
		"Fedora 38":         "fedora:38",
	},
	InstallParamsToTest: &ParamsMap{
		"default": Params{},
		"with-multiple-flags": Params{
			Flags: []string{
				"--redirect-all-dns-traffic",
				"--vnet", "docker0:172.17.0.0/16",
				"--vnet", "br+:172.18.0.0/16",
				"--vnet", "iface:::1/64",
				"--exclude-outbound-ports-for-uids", "53,3000-5000:106-108",
				"--exclude-outbound-ips", "10.0.0.1,192.168.0.0/24,fe80::1",
				"--exclude-outbound-ips", "fd00::/8",
				"--exclude-outbound-ports", "1,22,333",
				"--exclude-inbound-ports", "4444,55555",
				"--exclude-inbound-ips", "192.168.0.1,172.32.16.8/16,fe80::/10",
				"--exclude-inbound-ips", "a81b:a033:6399:73c7:72b6:aa8c:6f22:7098",
			},
		},
		"with-multiple-config-flags": Params{
			Flags: []string{
				"--config", "/1.yaml",
				"--config", "/2.yaml",
				"--config", "-",
			},
			EchoStdin: `
redirect:
  vnet:
    networks:
    - docker2:10.0.0.0/8
    - docker0:172.17.0.0/16
    - br+:172.18.0.0/16
    - iface:::1/64
  dns:
    enabled: true
    captureAll: true
  inbound:
    enabled: true
    excludePorts:
    - 4444
    - 55555
    excludePortsForIPs:
    - 192.168.0.1,172.32.16.8/16,fe80::/10
    - a81b:a033:6399:73c7:72b6:aa8c:6f22:7098
  outbound:
    enabled: true
    excludePortsForUIDs:
    - 1
    - 2-3:4-5
    excludePorts:
    - 6
    - 77
    - 888
    excludePortsForIPs:
    - 10.0.0.1,192.168.0.0/24,fe80::1
    - fd00::/8
`,
			Files: map[string]string{
				"/1.yaml": `
redirect:
  dns:
    enabled: false
    port: 7777
`,
				"/2.yaml": `
redirect:
  dns:
    enabled: true
    port: 8888
  outbound:
    excludePortsForIPs:
    - 10.11.12.13
    - 172.14.15.16/32
  inbound:
    excludePorts:
    - 10000
    - 20000
`,
			},
		},
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
