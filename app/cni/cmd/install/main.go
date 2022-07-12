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
	return "", nil
}

func lookForValidConfig(files []string, checkerFn func(string) bool) (string, bool) {
	for _, file := range files {
		found := checkerFn(file)
		if found {
			return file, true
		}
	}
	return "", false
}

func cleanup(mountedCniNetDir, cniConfName, kubeconfigName string, chainedCniPlugin bool) {
	removeBinFiles()
	err := revertConfig(mountedCniNetDir, cniConfName, chainedCniPlugin)
	if err != nil {
		log.Error(err, "could not revert config")
	}
	removeKubeconfig(mountedCniNetDir, kubeconfigName)
}

func removeKubeconfig(mountedCniNetDir, kubeconfigName string) {
	kubeconfigPath := mountedCniNetDir + "/" + kubeconfigName
	if fileExists(kubeconfigPath) {
		err := os.Remove(kubeconfigPath)
		if err != nil {
			log.V(1).Error(err, "couldn't remove cni conf file")
		}
	}
}

func revertConfig(mountedCniNetDir, cniConfName string, chainedCniPlugin bool) error {
	configPath := mountedCniNetDir + "/" + cniConfName

	if fileExists(configPath) {
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
	}

	return nil
}

func install() error {
	hostCniNetDir := env.GetString("CNI_NET_DIR", "/etc/cni/net.d")
	kubeconfigName := env.GetString("KUBECFG_FILE_NAME", "ZZZ-kuma-cni-kubeconfig")
	cfgCheckInterval, _ := env.GetInt("CFGCHECK_INTERVAL", 1)
	chainedCniPlugin, _ := env.GetBool("CHAINED_CNI_PLUGIN", true)
	mountedCniNetDir := env.GetString("MOUNTED_CNI_NET_DIR", "/host/etc/cni/net.d")

	cniConfFile, err := findCniConfFile(mountedCniNetDir)
	if err != nil {
		return err
	}
	cniConfName := env.GetString("CNI_CONF_NAME", cniConfFile)
	if cniConfName == "" {
		cniConfName = "YYY-kuma-cni.conflist"
	}
	defer cleanup(mountedCniNetDir, cniConfName, kubeconfigName, chainedCniPlugin)
	setupSignalCleanup(mountedCniNetDir, cniConfName, kubeconfigName, chainedCniPlugin)

	err = copyBinaries()
	if err != nil {
		return err
	}

	err = prepareKubeconfig(mountedCniNetDir, kubeconfigName, serviceAccountPath)
	if err != nil {
		return err
	}

	err = prepareKumaCniConfig(mountedCniNetDir, hostCniNetDir, kubeconfigName, serviceAccountPath, cniConfName, chainedCniPlugin)
	if err != nil {
		return err
	}

	shouldSleep, err := env.GetBool("SLEEP", true)
	if err != nil {
		shouldSleep = true // sleep by default
	}

	for shouldSleep {
		time.Sleep(time.Duration(cfgCheckInterval) * time.Second)
		checkInstall(mountedCniNetDir+"/"+cniConfName, chainedCniPlugin)
	}

	return nil
}

func prepareKumaCniConfig(mountedCniNetDir, hostCniNetDir, kubeconfigName, serviceAccountPath, cniConfName string, chainedCniPlugin bool) error {
	rawConfig := env.GetString("CNI_NETWORK_CONFIG", "")
	kubeconfigFilePath := hostCniNetDir + "/" + kubeconfigName

	config := strings.Replace(rawConfig, "__KUBECONFIG_FILEPATH__", kubeconfigFilePath, 1)
	log.V(1).Info("config after replace", "config", config)

	serviceAccountToken, err := ioutil.ReadFile(serviceAccountPath + "/token")
	if err != nil {
		return err
	}
	config = strings.Replace(config, "__SERVICEACCOUNT_TOKEN__", string(serviceAccountToken), 1)

	if chainedCniPlugin {
		err := setupChainedPlugin(mountedCniNetDir, cniConfName, config)
		if err != nil {
			log.Error(err, "unable to setup kuma cni as chained plugin")
			return err
		}
	}

	return nil
}

func setupChainedPlugin(mountedCniNetDir, cniConfName, kumaCniConfig string) error {
	resolvedName := cniConfName
	extension := filepath.Ext(cniConfName)
	if !fileExists(mountedCniNetDir+"/"+cniConfName) && extension == ".conf" && fileExists(mountedCniNetDir+"/"+cniConfName+"list") {
		resolvedName = cniConfName + "list"
	}

	if fileExists(mountedCniNetDir + "/" + resolvedName) {
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

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func prepareKubeconfig(mountedCniNetDir, kubeconfigName, serviceAccountPath string) error {
	serviceAccountTokenPath := serviceAccountPath + "/token"
	serviceAccountToken, err := ioutil.ReadFile(serviceAccountTokenPath)
	if err != nil {
		return err
	}

	if fileExists(serviceAccountTokenPath) {
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
		err = atomic.WriteFile(mountedCniNetDir+"/"+kubeconfigName, strings.NewReader(kubeconfig))
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
	var err error
	for _, dir := range dirs {
		if !isDirWriteable(dir) {
			log.Info("directory is not writeable", "dir", dir)
			continue
		}
		file, err := os.Open(kumaCniBinaryPath)
		if err != nil {
			log.Error(err, "can't open kuma-cni file")
			continue
		}

		stat, err := os.Stat(kumaCniBinaryPath)
		if err != nil {
			log.Error(err, "can't stat kuma-cni file")
		}

		log.V(1).Info("cni binary file permissions", "permissions", int(stat.Mode()), "path", kumaCniBinaryPath)
		destination := dir + "/kuma-cni"
		err = atomic.WriteFile(destination, file)
		if err != nil {
			log.Error(err, "can't atomically write kuma-cni file")
		}

		err = os.Chmod(destination, stat.Mode()|0111)
		if err != nil {
			log.Error(err, "can't chmod kuma-cni file")
		}

		if err != nil {
			log.Error(err, "can't atomically write cni file", "dir", dir)
		}

		log.Info("wrote kuma CNI binaries", "dir", dir)
		writtenOnce = true
	}

	if !writtenOnce {
		return err
	}
	return nil
}

// isDirWriteable checks if dir is writable by writing and removing a file
// to dir. It returns nil if dir is writable.
func isDirWriteable(dir string) bool {
	f := filepath.Join(dir, ".touch")
	perm := 0600
	if err := ioutil.WriteFile(f, []byte(""), fs.FileMode(perm)); err != nil {
		return false
	}
	return os.Remove(f) == nil
}

func setupSignalCleanup(mountedCniNetDir, cniConfName, kubeconfigName string, chainedCniPlugin bool) {
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		_ = <-osSignals
		cleanup(mountedCniNetDir, cniConfName, kubeconfigName, chainedCniPlugin)
	}()
}

func main() {
	err := install()
	if err != nil {
		log.Error(err, "error occurred during cni installation")
		os.Exit(1)
	}
}
