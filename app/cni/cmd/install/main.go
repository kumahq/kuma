package main

import (
	"bytes"
	"encoding/base64"
	"io/fs"
	"io/ioutil"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/natefinch/atomic"
	"github.com/pkg/errors"
	"k8s.io/utils/env"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/util/files"
)

const (
	kumaCniBinaryPath  = "/opt/cni/bin/kuma-cni"
	primaryBinDir      = "/host/opt/cni/bin"
	secondaryBinDir    = "/host/secondary-bin-dir"
	serviceAccountPath = "/var/run/secrets/kubernetes.io/serviceaccount"
)

var (
	log = core.NewLoggerWithRotation(2, "/tmp/install-cni.log", 100, 0, 0).WithName("install-cni")
)

func removeBinFiles() {
	log.V(1).Info("removing existing binaries")
	err := os.Remove("/host/opt/cni/bin/kuma-cni")
	if err != nil {
		log.V(1).Error(err, "couldn't remove cni bin file")
	}
}

func cleanup(ic *InstallerConfig) {
	removeBinFiles()
	err := revertConfig(ic.MountedCniNetDir, ic.CniConfName, ic.ChainedCniPlugin)
	if err != nil {
		log.Error(err, "could not revert config")
	}
	removeKubeconfig(ic.MountedCniNetDir, ic.KubeconfigName)
}

func removeKubeconfig(mountedCniNetDir, kubeconfigName string) {
	kubeconfigPath := mountedCniNetDir + "/" + kubeconfigName
	if files.FileExists(kubeconfigPath) {
		err := os.Remove(kubeconfigPath)
		if err != nil {
			log.V(1).Error(err, "couldn't remove cni conf file")
		}
	}
}

func revertConfig(mountedCniNetDir, cniConfName string, chainedCniPlugin bool) error {
	configPath := mountedCniNetDir + "/" + cniConfName

	if !files.FileExists(configPath) {
		log.Info("no need to revert config - file does not exist", "configPath", configPath)
		return nil
	}

	if chainedCniPlugin {
		contents, err := ioutil.ReadFile(configPath)
		if err != nil {
			return errors.Wrap(err, "couldn't read cni conf file")
		}
		newContents, err := revertConfigContents(contents)
		if err != nil {
			return errors.Wrap(err, "could not revert config contents")
		}
		err = atomic.WriteFile(configPath, bytes.NewReader(newContents))
		if err != nil {
			return errors.Wrap(err, "could not write new conf")
		}
	} else {
		err := os.Remove(configPath)
		if err != nil {
			return errors.Wrap(err, "couldn't remove cni conf file")
		}
	}

	return nil
}

func install(ic *InstallerConfig) error {
	err := copyBinaries()
	if err != nil {
		return err
	}

	err = prepareKubeconfig(ic.MountedCniNetDir, ic.KubeconfigName, serviceAccountPath)
	if err != nil {
		return err
	}

	err = prepareKumaCniConfig(ic, serviceAccountPath)
	if err != nil {
		return err
	}


	return nil
}

func prepareKumaCniConfig(ic *InstallerConfig, serviceAccountPath string) error {
	rawConfig := env.GetString("CNI_NETWORK_CONFIG", "")
	kubeconfigFilePath := ic.HostCniNetDir + "/" + ic.KubeconfigName

	config := strings.Replace(rawConfig, "__KUBECONFIG_FILEPATH__", kubeconfigFilePath, 1)
	log.V(1).Info("config after replace", "config", config)

	serviceAccountToken, err := ioutil.ReadFile(serviceAccountPath + "/token")
	if err != nil {
		return err
	}
	config = strings.Replace(config, "__SERVICEACCOUNT_TOKEN__", string(serviceAccountToken), 1)

	if ic.ChainedCniPlugin {
		err := setupChainedPlugin(ic.MountedCniNetDir, ic.CniConfName, config)
		if err != nil {
			return errors.Wrap(err, "unable to setup kuma cni as chained plugin")
		}
	}

	return nil
}

func setupChainedPlugin(mountedCniNetDir, cniConfName, kumaCniConfig string) error {
	resolvedName := cniConfName
	extension := filepath.Ext(cniConfName)
	if !files.FileExists(mountedCniNetDir+"/"+cniConfName) && extension == ".conf" && files.FileExists(mountedCniNetDir+"/"+cniConfName+"list") {
		resolvedName = cniConfName + "list"
	}

	if files.FileExists(mountedCniNetDir + "/" + resolvedName) {
		hostCniConfig, err := ioutil.ReadFile(mountedCniNetDir + "/" + resolvedName)
		if err != nil {
			return err
		}

		marshaled, err := transformJsonConfig(kumaCniConfig, hostCniConfig)
		if err != nil {
			return err
		}
		log.V(1).Info("resulting config", "config", string(marshaled))

		err = atomic.WriteFile(mountedCniNetDir+"/"+resolvedName, bytes.NewReader(marshaled))
		if err != nil {
			return err
		}

		return nil
	}
	return nil
}

func prepareKubeconfig(mountedCniNetDir, kubeconfigName, serviceAccountPath string) error {
	kubeconfigPath := mountedCniNetDir + "/" + kubeconfigName
	serviceAccountTokenPath := serviceAccountPath + "/token"
	serviceAccountToken, err := ioutil.ReadFile(serviceAccountTokenPath)
	if err != nil {
		return err
	}

	if files.FileExists(serviceAccountTokenPath) {
		kubernetesServiceHost := env.GetString("KUBERNETES_SERVICE_HOST", "")
		if kubernetesServiceHost == "" {
			return errors.New("KUBERNETES_SERVICE_HOST env variable not set")
		}

		kubernetesServicePort := env.GetString("KUBERNETES_SERVICE_PORT", "")
		if kubernetesServicePort == "" {
			return errors.New("KUBERNETES_SERVICE_PORT env variable not set")
		}

		kubeCaFile := env.GetString("KUBE_CA_FILE", serviceAccountPath+"/ca.crt")
		kubeCa, err := ioutil.ReadFile(kubeCaFile)
		if err != nil {
			return err
		}
		kubernetesServiceProtocol := env.GetString("KUBERNETES_SERVICE_PROTOCOL", "https")
		caData := base64.StdEncoding.EncodeToString(kubeCa)

		kubeconfig := kubeconfigTemplate(kubernetesServiceProtocol, kubernetesServiceHost, kubernetesServicePort, string(serviceAccountToken), caData)
		log.Info("writing kubeconfig", "path", kubeconfigPath)
		err = atomic.WriteFile(kubeconfigPath, strings.NewReader(kubeconfig))
		if err != nil {
			return err
		}
	}

	return nil
}

func kubeconfigTemplate(protocol, host, port, token, caData string) string {
	safeHost := host
	if govalidator.IsIPv6(host) {
		if !(strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]")) {
			safeHost = "[" + host + "]"
		}
	}

	serverUrl := url.URL{
		Scheme: protocol,
		Host:   safeHost + ":" + port,
	}

	return `# Kubeconfig file for kuma CNI plugin.
apiVersion: v1
kind: Config
clusters:
- name: local
  cluster:
    server: ` + serverUrl.String() + `
    certificate-authority-data: ` + caData + `
users:
- name: kuma-cni
  user:
    token: ` + token + `
contexts:
- name: kuma-cni-context
  context:
    cluster: local
    user: kuma-cni
current-context: kuma-cni-context`
}

func copyBinaries() error {
	dirs := []string{primaryBinDir, secondaryBinDir}
	writtenOnce := false
	allErrors := errors.New("combined errors for copying binaries")
	for _, dir := range dirs {
		err := tryWritingToDir(dir)
		if err != nil {
			allErrors = errors.Wrapf(err, "could not write to dir %v", dir)
			log.Error(err, "writing to dir failed", "dir", dir)
			continue
		}

		log.Info("wrote kuma CNI binaries", "dir", dir)
		writtenOnce = true
	}

	if !writtenOnce {
		return allErrors
	}
	return nil
}

func tryWritingToDir(dir string) error {
	if err := isDirWriteable(dir); err != nil {
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
	err = atomic.WriteFile(destination, file)
	if err != nil {
		return errors.Wrap(err, "can't atomically write kuma-cni file")
	}

	err = os.Chmod(destination, stat.Mode()|0111)
	if err != nil {
		return errors.Wrap(err, "can't chmod kuma-cni file")
	}

	if err != nil {
		return errors.Wrap(err, "can't atomically write cni file")
	}

	return nil
}

// isDirWriteable checks if dir is writable by writing and removing a file
// to dir. It returns true if dir is writable.
func isDirWriteable(dir string) error {
	f := filepath.Join(dir, ".touch")
	perm := 0600
	if err := ioutil.WriteFile(f, []byte(""), fs.FileMode(perm)); err != nil {
		return err
	}
	return os.Remove(f)
}

func setupSignalCleanup(ic *InstallerConfig) {
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		_ = <-osSignals
		cleanup(ic)
	}()
}

func main() {
	ic, err := loadInstallerConfig()
	if err != nil {
		log.Error(err, "error occurred during config loading")
		os.Exit(1)
	}
	err = install(ic)
	defer cleanup(ic)
	setupSignalCleanup(ic)

	if err != nil {
		log.Error(err, "error occurred during cni installation")
		os.Exit(1)
	}

	for ic.ShouldSleep {
		time.Sleep(time.Duration(ic.CfgCheckInterval) * time.Second)
		err := checkInstall(ic.MountedCniNetDir+"/"+ic.CniConfName, ic.ChainedCniPlugin)
		if err != nil {
			log.Error(err, "checking installation failed - exiting")
			os.Exit(1)
		}
	}
}
