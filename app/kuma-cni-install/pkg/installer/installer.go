package installer

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/pkg/errors"
)

type Installer struct {
	cfg                *Config
	status             *atomic.Value
	saToken            string
	kubeconfigFilepath string
	cniConfigFilepath  string
}

func NewInstaller(cfg *Config, status *atomic.Value) *Installer {
	return &Installer{
		cfg:    cfg,
		status: status,
	}
}

func (in *Installer) Run(ctx context.Context) (err error) {
	for {
		if err = copyDir(in.cfg.CNIBinSourceDir, in.cfg.CNIBinDestinationDir); err != nil {
			return
		}

		if in.saToken, err = readServiceAccountToken(); err != nil {
			return
		}

		if in.kubeconfigFilepath, err = createKubeconfigFile(in.cfg, in.saToken); err != nil {
			return
		}

		if in.cniConfigFilepath, err = createCNIConfigFile(ctx, in.cfg, in.saToken); err != nil {
			return
		}

		if err = sleepCheckInstall(ctx, in.cfg, in.cniConfigFilepath, in.status); err != nil {
			return
		}
		// Invalid config; pod set to "NotReady"
		log.Println("Restarting...")
	}
}

// Cleanup remove Kuma CNI's config, kubeconfig file, and binaries.
func (in *Installer) Cleanup() error {
	log.Println("Cleaning up.")
	if len(in.cniConfigFilepath) > 0 && fileExists(in.cniConfigFilepath) {
		if in.cfg.ChainedCNIPlugin {
			log.Printf("Removing Kuma CNI config from CNI config file: %s", in.cniConfigFilepath)

			// Read JSON from CNI config file
			cniConfigMap, err := ReadCNIConfigMap(in.cniConfigFilepath)
			if err != nil {
				return err
			}
			// Find Kuma CNI and remove from plugin list
			plugins, err := GetPlugins(cniConfigMap)
			if err != nil {
				return errors.Wrap(err, in.cniConfigFilepath)
			}
			for i, rawPlugin := range plugins {
				plugin, err := GetPlugin(rawPlugin)
				if err != nil {
					return errors.Wrap(err, in.cniConfigFilepath)
				}
				if plugin["type"] == "kuma-cni" {
					cniConfigMap["plugins"] = append(plugins[:i], plugins[i+1:]...)
					break
				}
			}

			cniConfig, err := MarshalCNIConfig(cniConfigMap)
			if err != nil {
				return err
			}
			if err = fileAtomicWrite(in.cniConfigFilepath, cniConfig, os.FileMode(0o644)); err != nil {
				return err
			}
		} else {
			log.Printf("Removing Kuma CNI config file: %s", in.cniConfigFilepath)
			if err := os.Remove(in.cniConfigFilepath); err != nil {
				return err
			}
		}
	}

	if len(in.kubeconfigFilepath) > 0 && fileExists(in.kubeconfigFilepath) {
		log.Printf("Removing Kuma CNI kubeconfig file: %s", in.kubeconfigFilepath)
		if err := os.Remove(in.kubeconfigFilepath); err != nil {
			return err
		}
	}

	kumaCNIBin := filepath.Join(in.cfg.CNIBinDestinationDir, "kuma-cni")
	if fileExists(kumaCNIBin) {
		log.Printf("Removing binary: %s", kumaCNIBin)
		if err := os.Remove(kumaCNIBin); err != nil {
			return err
		}
	}

	return nil
}

func readServiceAccountToken() (string, error) {
	saToken := ServiceAccountPath + "/token"
	if !fileExists(saToken) {
		return "", fmt.Errorf("service account token file %s does not exist. Is this not running within a pod?", saToken)
	}

	token, err := ioutil.ReadFile(saToken)
	if err != nil {
		return "", err
	}

	return string(token), nil
}

// sleepCheckInstall verifies the configuration then blocks until an invalid configuration is detected, and return nil.
// If an error occurs or context is canceled, the function will return the error.
// Returning from this function will set the pod to "NotReady".
func sleepCheckInstall(ctx context.Context, cfg *Config, cniConfigFilepath string, isReady *atomic.Value) error {
	// Create file watcher before checking for installation
	// so that no file modifications are missed while and after checking
	watcher, fileModified, errChan, err := CreateFileWatcher(cfg.MountedCNINetDir)
	if err != nil {
		return err
	}
	defer func() {
		isReady.Store(false)
		_ = watcher.Close()
	}()

	for {
		if checkErr := checkInstall(cfg, cniConfigFilepath); checkErr != nil {
			// Pod set to "NotReady" due to invalid configuration
			log.Printf("Invalid configuration. %v", checkErr)
			return nil
		}
		// Check if file has been modified or if an error has occurred during checkInstall before setting isReady to true
		select {
		case <-fileModified:
			return nil
		case err := <-errChan:
			return err
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Valid configuration; set isReady to true and wait for modifications before checking again
			isReady.Store(true)
			if err = WaitForFileMod(ctx, fileModified, errChan); err != nil {
				// Pod set to "NotReady" before termination
				return err
			}
		}
	}
}

// checkInstall returns an error if an invalid CNI configuration is detected
func checkInstall(cfg *Config, cniConfigFilepath string) error {
	defaultCNIConfigFilename, err := getDefaultCNINetwork(cfg.MountedCNINetDir)
	if err != nil {
		return err
	}
	defaultCNIConfigFilepath := filepath.Join(cfg.MountedCNINetDir, defaultCNIConfigFilename)
	if defaultCNIConfigFilepath != cniConfigFilepath {
		if len(cfg.CNIConfName) > 0 {
			// Install was run with overridden CNI config file so don't error out on preempt check
			// Likely the only use for this is testing the script
			log.Printf("CNI config file %s preempted by %s", cniConfigFilepath, defaultCNIConfigFilepath)
		} else {
			return fmt.Errorf("CNI config file %s preempted by %s", cniConfigFilepath, defaultCNIConfigFilepath)
		}
	}

	if !fileExists(cniConfigFilepath) {
		return fmt.Errorf("CNI config file removed: %s", cniConfigFilepath)
	}

	if cfg.ChainedCNIPlugin {
		// Verify that Kuma CNI config exists in the CNI config plugin list
		cniConfigMap, err := ReadCNIConfigMap(cniConfigFilepath)
		if err != nil {
			return err
		}
		plugins, err := GetPlugins(cniConfigMap)
		if err != nil {
			return errors.Wrap(err, cniConfigFilepath)
		}
		for _, rawPlugin := range plugins {
			plugin, err := GetPlugin(rawPlugin)
			if err != nil {
				return errors.Wrap(err, cniConfigFilepath)
			}
			if plugin["type"] == "kuma-cni" {
				return nil
			}
		}

		return fmt.Errorf("kuma-cni CNI config removed from CNI config file: %s", cniConfigFilepath)
	}
	// Verify that Kuma CNI config exists as a standalone plugin
	cniConfigMap, err := ReadCNIConfigMap(cniConfigFilepath)
	if err != nil {
		return err
	}

	if cniConfigMap["type"] != "kuma-cni" {
		return fmt.Errorf("kuma-cni CNI config file modified: %s", cniConfigFilepath)
	}
	return nil
}
