package main

import (
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/config"
)

const (
	defaultKumaCniConfName = "YYY-kuma-cni.conflist"
)

type InstallerConfig struct {
	HostCniNetDir    string `envconfig:"cni_net_dir" default:"/etc/cni/net.d"`
	KubeconfigName   string `envconfig:"kubecfg_file_name" default:"ZZZ-kuma-cni-kubeconfig"`
	CfgCheckInterval int    `envconfig:"cfgcheck_interval" default:"1"`
	ChainedCniPlugin bool   `envconfig:"chained_cni_plugin" default:"true"`
	MountedCniNetDir string `envconfig:"mounted_cni_net_dir" default:"/host/etc/cni/net.d"`
	ShouldSleep      bool   `envconfig:"sleep" default:"true"`
	CniConfName      string `envconfig:"cni_conf_name" default:""`
}

func (i InstallerConfig) Sanitize() {
}

func (i InstallerConfig) Validate() error {
	if i.CfgCheckInterval <= 0 {
		return errors.New("CFGCHECK_INTERVAL env variable needs to be greater than 0")
	}

	// should I check that dirs exist?

	return nil
}

func lookForValidConfig(files []string, checkerFn func(string) error) (string, bool) {
	for _, file := range files {
		err := checkerFn(file)
		if err != nil {
			log.Error(err, "error occurred testing config file", "file", file)
		} else {
			return file, true
		}
	}
	return "", false
}

func findCniConfFile(mountedCNINetDir string) (string, error) {
	files, err := filepath.Glob(mountedCNINetDir + "/*.conf")
	if err != nil {
		return "", err
	}

	file, found := lookForValidConfig(files, isValidConfFile)
	if found {
		return file, nil
	}

	files, err = filepath.Glob(mountedCNINetDir + "/*.conflist")
	if err != nil {
		return "", err
	}
	file, found = lookForValidConfig(files, isValidConflistFile)
	if found {
		return file, nil
	}

	// use default
	return "", errors.New("cni conf file not found - use default")
}

func loadInstallerConfig() (*InstallerConfig, error) {
	var installerConfig InstallerConfig
	err := config.Load("", &installerConfig)
	if err != nil {
		return nil, err
	}

	if installerConfig.CniConfName == "" {
		cniConfFile, err := findCniConfFile(installerConfig.MountedCniNetDir)
		if err != nil {
			log.Error(err, "could not find cni conf file using default")
			installerConfig.CniConfName = defaultKumaCniConfName
		}
		installerConfig.CniConfName = cniConfFile
	}

	return &installerConfig, nil
}
