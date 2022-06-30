package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/itchyny/gojq"
	"github.com/natefinch/atomic"
	"go.uber.org/zap"
	"k8s.io/utils/env"

	"github.com/kumahq/kuma/pkg/core"
)

var (
	log = core.Log.WithName("cni-install")
)

func removeBinFiles() {
	// todo hook this up to cleanup
	log.V(1).Info("removing existing binaries")
	_ = os.Remove("/host/opt/cni/bin/kuma-cni")
}

func findCniConfFile(mountedCNINetDir string) string {
	files, _ := filepath.Glob(mountedCNINetDir + "/*.conf")
	file, found := lookForValidConfig(files)
	if found {
		return file
	}

	files, _ = filepath.Glob(mountedCNINetDir + "/*.conflist")
	file, found = lookForValidConfig(files)
	if found {
		return file
	}

	// probably should return an error
	return ""
}

func lookForValidConfig(files []string) (string, bool) {
	for _, file := range files {
		found := isFileAValidConfig(file)
		if found {
			return file, true
		}
	}
	return "", false
}

func isFileAValidConfig(file string) bool {
	var parsed map[string]interface{}
	contents, _ := ioutil.ReadFile(file)
	// this is probably going to be rewritten to not use `jq` at all
	err := json.Unmarshal(contents, &parsed)
	if err != nil {
		log.Error(err, "could not unmarshal config file")
		return false
	}
	query, _ := gojq.Parse(`has("type")`)
	iterator := query.Run(parsed)
	v, ok := iterator.Next()
	if !ok {
		return false
	}
	log.Info("checking file", zap.String("file", file))
	if v.(bool) == true {
		return true
	}
	return false
}

func check_install(mountedCNINetDir string) {
	// todo: implement
}

func cleanup() {
	removeBinFiles()
	// todo: implement reverting config
}

func install() error {
	hostCniNetDir := env.GetString("CNI_NET_DIR", "/etc/cni/net.d")
	kubecfgName := env.GetString("KUBECFG_FILE_NAME", "ZZZ-kuma-cni-kubeconfig")
	cfgCheckInterval, _ := env.GetInt("CFGCHECK_INTERVAL", 1)
	chainedCniPlugin, _ := env.GetBool("CHAINED_CNI_PLUGIN", true)
	mountedCniNetDir := env.GetString("MOUNTED_CNI_NET_DIR", "/host/etc/cni/net.d")
	serviceAccountPath := "/var/run/secrets/kubernetes.io/serviceaccount"
	cniConfName := env.GetString("CNI_CONF_NAME", findCniConfFile(mountedCniNetDir))
	if cniConfName == "" {
		cniConfName = "YYY-kuma-cni.conflist"
	}

	copyBinaries()
	err := prepareKubeconfig(mountedCniNetDir, kubecfgName, serviceAccountPath)
	if err != nil {
		return err
	}
	err = prepareKumaCniConfig(mountedCniNetDir, hostCniNetDir, kubecfgName, serviceAccountPath, cniConfName, chainedCniPlugin)
	if err != nil {
		return err
	}

	shouldSleep, _ := env.GetBool("SLEEP", true)

	for shouldSleep == true {
		time.Sleep(time.Duration(cfgCheckInterval) * time.Second)
		check_install(mountedCniNetDir)
	}

	return nil
}

func prepareKumaCniConfig(mountedCniNetDir, hostCniNetDir, kubecfgName, serviceAccountPath, cniConfName string, chainedCniPlugin bool) error {
	rawConfig := env.GetString("CNI_NETWORK_CONFIG", "")
	kubeconfigFilePath := hostCniNetDir + "/" + kubecfgName

	config := strings.Replace(rawConfig, "__KUBECONFIG_FILEPATH__", kubeconfigFilePath, 1)
	log.V(1).Info("config after replace", zap.String("config", config))

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
		log.V(1).Info("resulting config", zap.String("config", string(marshaled)))

		err = atomic.WriteFile(mountedCniNetDir+"/"+resolvedName, bytes.NewReader(marshaled))
		if err != nil {
			return err
		}

		return nil
	}
	return nil
}

func transformJsonConfig(kumaCniConfig string, hostCniConfig []byte) ([]byte, error) {
	queryString := `if has("type") then
   .plugins = [.]
   | del(.plugins[0].cniVersion)
   | to_entries
   | map(select(.key=="plugins"))
   | from_entries
   | .plugins += [` + kumaCniConfig + `]
   | .name = "k8s-pod-network"
   | .cniVersion = "0.3.0"
else
  del(.plugins[]? | select(.type == "kuma-cni"))
  | .plugins += [` + kumaCniConfig + `]
end`
	// this is probably going to be rewritten to not use `jq` at all
	query, err := gojq.Parse(queryString)
	if err != nil {
		return nil, err
	}

	var parsed map[string]interface{}
	err = json.Unmarshal(hostCniConfig, &parsed)
	if err != nil {
		return nil, err
	}

	result := query.Run(parsed)
	modified, _ := result.Next()
	marshaled, err := json.MarshalIndent(modified, "", "  ")
	if err != nil {
		return nil, err
	}
	return marshaled, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func prepareKubeconfig(mountedCniNetDir, kubecfgName, serviceAccountPath string) error {
	serviceAccountTokenPath := serviceAccountPath + "/token"
	serviceAccountToken, err := ioutil.ReadFile(serviceAccountTokenPath)
	if err != nil {
		return err
	}

	if fileExists(serviceAccountTokenPath) {
		kubernetesServiceHost := env.GetString("KUBERNETES_SERVICE_HOST", "")
		if kubernetesServiceHost == "" {
			log.Info("KUBERNETES_SERVICE_HOST env variable not set")
			os.Exit(1)
		}

		kubernetesServicePort := env.GetString("KUBERNETES_SERVICE_PORT", "")
		if kubernetesServicePort == "" {
			log.Info("KUBERNETES_SERVICE_PORT env variable not set")
			os.Exit(1)
		}

		kubeCaFile := env.GetString("KUBE_CA_FILE", serviceAccountPath+"/ca.crt")
		kubeCa, err := ioutil.ReadFile(kubeCaFile)
		if err != nil {
			return err
		}
		kubernetesServiceProtocol := env.GetString("KUBERNETES_SERVICE_PROTOCOL", "https")
		tlsConfig := "certificate-authority-data: " + base64.StdEncoding.EncodeToString(kubeCa)

		kubeconfig := kubeconfigTemplate(kubernetesServiceProtocol, kubernetesServiceHost, kubernetesServicePort, string(serviceAccountToken), tlsConfig)
		err = atomic.WriteFile(mountedCniNetDir+"/"+kubecfgName, strings.NewReader(kubeconfig))
		if err != nil {
			return err
		}
	}

	return nil
}

func kubeconfigTemplate(protocol, host, port, token, tlsConfig string) string {
	return `
# Kubeconfig file for kuma CNI plugin.
apiVersion: v1
kind: Config
clusters:
- name: local
  cluster:
    server: ` + protocol + `://` + host + `:` + port + `
    ` + tlsConfig + `
users:
- name: kuma-cni
  user:
    token: ` + token + `
contexts:
- name: kuma-cni-context
  context:
    cluster: local
    user: kuma-cni
current-context: kuma-cni-context
`
}

func copyBinaries() {
	dirs := []string{"/host/opt/cni/bin", "/host/secondary-bin-dir"}
	// todo: if not written anywhere fail
	for _, dir := range dirs {
		if !isDirWriteable(dir) {
			log.Info("directory is not writeable", zap.String("dir", dir))
			continue
		}
		file, err := os.Open("/opt/cni/bin/kuma-cni")
		if err != nil {
			log.Error(err, "can't open kuma-cni file")
			os.Exit(1)
		}

		stat, err := os.Stat("/opt/cni/bin/kuma-cni")
		if err != nil {
			log.Error(err, "can't stat kuma-cni file")
			os.Exit(1)
		}

		log.V(1).Info("/opt/cni/bin/kuma-cni file permissions", zap.Int("permissions", int(stat.Mode())))
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
			log.Error(err, "can't atomically write cni file", zap.String("dir", dir))
			os.Exit(1)
		}

		log.Info("wrote kuma CNI binaries", zap.String("dir", dir))
	}
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

func main() {
	err := install()
	if err != nil {
		log.Error(err, "error occurred")
		return
	}
	cleanup()
}
