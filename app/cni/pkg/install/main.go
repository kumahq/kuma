package install

import (
	"bytes"
	"context"
	std_errors "errors"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/natefinch/atomic"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-retry"

	kuma_log "github.com/kumahq/kuma/pkg/log"
	"github.com/kumahq/kuma/pkg/util/files"
)

const (
	kumaCniBinaryPath = "/opt/cni/bin/kuma-cni"
	primaryBinDir     = "/host/opt/cni/bin"
	secondaryBinDir   = "/host/secondary-bin-dir"
	saPath            = "/var/run/secrets/kubernetes.io/serviceaccount"
	saToken           = saPath + "/token"
	saCACrt           = saPath + "/ca.crt"
	readyFilePath     = "/tmp/ready"
	defaultLogName    = "install-cni"
)

var log = CreateNewLogger(defaultLogName, kuma_log.DebugLevel)

func removeBinFiles() error {
	return os.Remove("/host/opt/cni/bin/kuma-cni")
}

func cleanup(ic *InstallerConfig) {
	log.Info("starting cleanup")
	if err := removeBinFiles(); err != nil {
		log.Error(err, "could not remove cni bin file")
	} else {
		log.V(1).Info("removed existing binaries")
	}
	if err := revertConfig(ic.MountedCniNetDir + "/" + ic.CniConfName, ic.ChainedCniPlugin); err != nil {
		log.Error(err, "could not revert config")
	} else {
		log.V(1).Info("reverted config")
	}
	if err := removeKubeconfig(ic.MountedCniNetDir, ic.KubeconfigName); err != nil {
		log.Error(err, "could not remove kubeconfig")
	} else {
		log.V(1).Info("removed kubeconfig")
	}
	if err := os.Remove(readyFilePath); err != nil {
		log.Error(err, "couldn't remove ready file")
	} else {
		log.V(1).Info("removed ready file")
	}
	log.Info("finished cleanup")
}

func removeKubeconfig(mountedCniNetDir, kubeconfigName string) error {
	kubeconfigPath := mountedCniNetDir + "/" + kubeconfigName
	if files.FileExists(kubeconfigPath) {
		err := os.Remove(kubeconfigPath)
		if err != nil {
			return errors.Wrap(err, "couldn't remove kubeconfig file")
		}
	}
	return nil
}

func revertConfig(configPath string, chained bool) error {
	if !files.FileExists(configPath) {
		log.Info("no need to revert config - file does not exist", "configPath", configPath)
		return nil
	}

	if chained {
		contents, err := os.ReadFile(configPath)
		if err != nil {
			return errors.Wrap(err, "could not read cni conf file")
		}

		newContents, err := revertConfigContents(contents)
		if err != nil {
			return errors.Wrap(err, "could not revert config contents")
		}

		if err := atomic.WriteFile(configPath, bytes.NewReader(newContents)); err != nil {
			return errors.Wrap(err, "could not write new conf")
		}

		return nil
	}

	if err := os.Remove(configPath); err != nil {
		return errors.Wrap(err, "could not remove cni conf file")
	}

	return nil
}

func install(ic *InstallerConfig) error {
	if err := copyBinaries(); err != nil {
		return errors.Wrap(err, "could not copy binary files")
	}

	if err := prepareKubeconfig(ic, saToken, saCACrt); err != nil {
		return errors.Wrap(err, "could not prepare kubeconfig")
	}

	if err := prepareKumaCniConfig(ic, saToken); err != nil {
		return errors.Wrap(err, "could not prepare kuma cni config")
	}

	return nil
}

func setupChainedPlugin(mountedCniNetDir, cniConfName, kumaCniConfig string) error {
	resolvedName := cniConfName
	extension := filepath.Ext(cniConfName)
	if !files.FileExists(mountedCniNetDir+"/"+cniConfName) && extension == ".conf" && files.FileExists(mountedCniNetDir+"/"+cniConfName+"list") {
		resolvedName = cniConfName + "list"
	}

	cniConfPath := path.Join(mountedCniNetDir, resolvedName)
	backoff := retry.WithMaxDuration(5*time.Minute, retry.NewConstant(time.Second))
	err := retry.Do(context.Background(), backoff, func(ctx context.Context) error {
		if !files.FileExists(cniConfPath) {
			err := errors.Errorf("CNI config '%s' not found.", cniConfPath)
			log.Error(err, "error chaining Kuma CNI config, will retry...")
			return retry.RetryableError(err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	hostCniConfig, err := os.ReadFile(cniConfPath)
	if err != nil {
		return err
	}

	marshaled, err := transformJsonConfig(kumaCniConfig, hostCniConfig)
	if err != nil {
		return err
	}
	log.V(1).Info("resulting config", "config", string(marshaled))

	log.Info("chaining Kuma CNI config. Updating CNI config file", "file", mountedCniNetDir+"/"+resolvedName)
	err = atomic.WriteFile(mountedCniNetDir+"/"+resolvedName, bytes.NewReader(marshaled))
	return err
}

func copyBinaries() error {
	var errs error

	for _, dir := range []string{primaryBinDir, secondaryBinDir} {
		err := tryWritingToDir(dir)
		if err == nil {
			log.Info("successfully wrote kuma CNI binaries", "directory", dir)
			return nil
		}

		errs = std_errors.Join(
			errs,
			errors.Wrapf(err, "failed to write binaries to directory %s", dir),
		)

		log.Info("failed to write binaries", "directory", dir)
	}

	return errs
}

func tryWritingToDir(dir string) error {
	if err := files.IsDirWriteable(dir); err != nil {
		return errors.Wrap(err, "directory is not writeable")
	}
	file, err := os.Open(kumaCniBinaryPath)
	if err != nil {
		return errors.Wrap(err, "can't open kuma-cni file")
	}

	stat, err := os.Stat(kumaCniBinaryPath)
	if err != nil {
		return errors.Wrap(err, "can't stat kuma-cni file")
	}
	log.V(1).Info("cni binary file permissions", "permissions", int(stat.Mode()), "path", kumaCniBinaryPath)

	destination := dir + "/kuma-cni"
	if err := atomic.WriteFile(destination, file); err != nil {
		return errors.Wrap(err, "can't atomically write kuma-cni file")
	}

	if err := os.Chmod(destination, stat.Mode()|0o111); err != nil {
		return errors.Wrap(err, "can't chmod kuma-cni file")
	}

	return nil
}

func Run() {
	installerConfig, err := loadInstallerConfig()
	if err != nil {
		log.Error(err, "error occurred during config loading")
		os.Exit(1)
	}

	if err := SetLogLevel(&log, installerConfig.CniLogLevel, defaultLogName); err != nil {
		log.Error(err, "error occurred during setting the log level")
		os.Exit(2)
	}

	if err := install(installerConfig); err != nil {
		log.Error(err, "error occurred during cni installation")
		os.Exit(3)
	}

	if err := atomic.WriteFile(readyFilePath, strings.NewReader("")); err != nil {
		log.Error(err, "unable to mark as ready")
		os.Exit(4)
	}

	if err := runLoop(installerConfig); err != nil {
		log.Error(err, "checking installation failed - exiting")
		os.Exit(5)
	}
}

func runLoop(ic *InstallerConfig) error {
	defer cleanup(ic)
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	if !ic.ShouldSleep {
		return nil
	}

	for {
		select {
		case <-osSignals:
			return nil
		case <-time.After(time.Duration(ic.CfgCheckInterval) * time.Second):
			if err := checkInstall(ic.MountedCniNetDir+"/"+ic.CniConfName, ic.ChainedCniPlugin); err != nil {
				return err
			}
		}
	}
}
