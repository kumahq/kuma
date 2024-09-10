package cni

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path"
	"time"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-retry"

	"github.com/kumahq/kuma/app/cni/pkg/install"
	"github.com/kumahq/kuma/pkg/core"
	kuma_log "github.com/kumahq/kuma/pkg/log"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/util"
	kuma_version "github.com/kumahq/kuma/pkg/version"
)

const (
	podRetrievalMaxRetries = 30
	podRetrievalInterval   = 1 * time.Second
	defaultLogLocation     = "/tmp/kuma-cni.log"
	defaultLogLevel        = kuma_log.DebugLevel
	defaultLogName         = "kuma-cni"
	installCNIBinary       = "/install-cni"
)

var log = core.NewLoggerWithRotation(defaultLogLevel, defaultLogLocation, 100, 0, 0).WithName(defaultLogName)

// Kubernetes a K8s specific struct to hold config
type Kubernetes struct {
	Kubeconfig        string   `json:"kubeconfig"`
	ExcludeNamespaces []string `json:"exclude_namespaces"`
	CniBinDir         string   `json:"cni_bin_dir"`
}

type PluginConf struct {
	types.NetConf

	RawPrevResult *map[string]interface{} `json:"prevResult"`
	PrevResult    *current.Result         `json:"-"`

	// plugin-specific fields
	LogLevel   string     `json:"log_level"`
	Kubernetes Kubernetes `json:"kubernetes"`
}

// K8sArgs is the valid CNI_ARGS used for Kubernetes
// The field names need to match exact keys in kubelet args for unmarshalling
type K8sArgs struct {
	types.CommonArgs
	IP                         net.IP
	K8S_POD_NAME               types.UnmarshallableString // nolint: golint, stylecheck
	K8S_POD_NAMESPACE          types.UnmarshallableString // nolint: golint, stylecheck
	K8S_POD_INFRA_CONTAINER_ID types.UnmarshallableString // nolint: golint, stylecheck
}

// parseConfig parses the supplied configuration (and prevResult) from stdin
func parseConfig(stdin []byte) (*PluginConf, error) {
	var conf PluginConf

	if err := json.Unmarshal(stdin, &conf); err != nil {
		return nil, errors.Wrap(err, "failed to parse network configuration from stdin")
	}

	if conf.RawPrevResult != nil {
		resultBytes, err := json.Marshal(conf.RawPrevResult)
		if err != nil {
			return nil, errors.Wrap(err, "failed to serialize previous result")
		}

		res, err := version.NewResult(conf.CNIVersion, resultBytes)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse previous result")
		}

		conf.RawPrevResult = nil

		if conf.PrevResult, err = current.NewResultFromResult(res); err != nil {
			return nil, errors.Wrap(err, "failed to convert previous result to current version")
		}
	}

	return &conf, nil
}

func hijackMainCNIProcessStderr(l logr.Logger) (*os.File, error) {
	file, err := getCniProcessStderr()
	if err != nil {
		l.Error(err, "failed to hijack stderr of the CNI process, continuing to log to the default location: "+defaultLogLocation)
		return nil, errors.Wrap(err, "unable to open CNI process stderr file")
	}

	os.Stderr = file

	l.Info("successfully hijacked stderr of the CNI process; logs will be visible via 'kubectl logs'")

	return file, nil
}

func getCniProcessStderr() (*os.File, error) {
	pid, err := pidOf(installCNIBinary)
	if err != nil {
		return nil, err
	}

	return os.OpenFile(path.Join("/proc", pid, "fd", "2"), os.O_WRONLY, 0)
}

// cmdAdd is called for ADD requests
func cmdAdd(args *skel.CmdArgs) error {
	conf, err := add(context.Background(), args)
	if err != nil {
		log.Info("[WARNING]: pod excluded", "reason", err)
		return err
	}

	result := conf.PrevResult
	if conf.PrevResult == nil {
		result = &current.Result{CNIVersion: current.ImplementedSpecVersion}
	}

	log.Info("cmdAdd result", "result", result)

	return types.PrintResult(result, conf.CNIVersion)
}

func add(ctx context.Context, args *skel.CmdArgs) (*PluginConf, error) {
	conf, err := parseConfig(args.StdinData)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse kuma-cni configuration in cmdAdd")
	}

	stderr, err := hijackMainCNIProcessStderr(log)
	if err != nil {
		return nil, errors.Wrap(err, "failed to hijack main process stderr")
	}
	defer stderr.Close()

	if err := install.SetLogLevel(&log, conf.LogLevel, defaultLogName); err != nil {
		return nil, errors.Wrap(err, "failed to set log level")
	}

	log.V(1).Info("cmdAdd config parsed", "version", conf.CNIVersion, "prevResult", conf.PrevResult)

	// Determine if running under k8s by checking the CNI args
	var k8sArgs K8sArgs
	if err := types.LoadArgs(args.Args, &k8sArgs); err != nil {
		return nil, errors.Wrap(err, "error loading kuma-cni cmdAdd args")
	}

	logger := log.WithValues(
		"pod", k8sArgs.K8S_POD_NAME,
		"namespace", k8sArgs.K8S_POD_NAMESPACE,
		"infraContainerId", k8sArgs.K8S_POD_INFRA_CONTAINER_ID,
		"ip", k8sArgs.IP,
		"containerId", args.ContainerID,
		"args", args.Args,
		"netns", args.Netns,
	)

	if string(k8sArgs.K8S_POD_NAMESPACE) == "" || string(k8sArgs.K8S_POD_NAME) == "" {
		logger.Info("pod excluded - no kubernetes data")
		return conf, nil
	}

	if shouldExcludePod(conf.Kubernetes.ExcludeNamespaces, k8sArgs.K8S_POD_NAMESPACE) {
		logger.Info("pod excluded - is in the namespace excluded by 'exclude_namespaces'")
		return conf, nil
	}

	containerCount, initContainersMap, annotations, err := getPodInfoWithRetries(ctx, conf, k8sArgs)
	if err != nil {
		return nil, errors.Wrap(err, "error getting pod info")
	}

	if isInitContainerPresent(initContainersMap) {
		logger.Info("pod excluded - already injected with kuma-init container")
		return conf, nil
	}

	if _, sidecarInInitContainers := initContainersMap[util.KumaSidecarContainerName]; containerCount < 2 && !sidecarInInitContainers {
		logger.Info("pod excluded - not enough containers in pod. Kuma-sidecar container required")
		return conf, nil
	}

	logger.V(1).Info("checking annotations prior to injecting redirect",
		"netns", args.Netns,
		"annotations", annotations)
	if excludeByMissingSidecarInjectedAnnotation(annotations) {
		logger.Info("pod excluded due to lack of 'kuma.io/sidecar-injected: true' annotation")
		return conf, nil
	}

	if intermediateConfig, configErr := NewIntermediateConfig(annotations); configErr != nil {
		return nil, errors.Wrap(configErr, "pod intermediateConfig failed due to bad params")
	} else {
		if err := Inject(ctx, args.Netns, intermediateConfig, logger); err != nil {
			return nil, errors.Wrap(err, "could not inject rules into namespace")
		}
	}

	logger.Info("successfully injected iptables rules")

	return conf, nil
}

func excludeByMissingSidecarInjectedAnnotation(annotations map[string]string) bool {
	excludePod := false
	val, ok := annotations[metadata.KumaSidecarInjectedAnnotation]
	if !ok || val != "true" {
		excludePod = true
	}
	return excludePod
}

func isInitContainerPresent(initContainersMap map[string]struct{}) bool {
	excludePod := false
	// Check if kuma-init container is present; in that case exclude pod
	if _, present := initContainersMap[util.KumaInitContainerName]; present {
		excludePod = true
	}
	return excludePod
}

func getPodInfoWithRetries(ctx context.Context, conf *PluginConf, k8sArgs K8sArgs) (int, map[string]struct{}, map[string]string, error) {
	client, err := newKubeClient(log, *conf)
	if err != nil {
		return 0, nil, nil, errors.Wrap(err, "could not create kube client")
	}

	var containerCount int
	var initContainersMap map[string]struct{}
	var annotations map[string]string
	var k8sErr error

	backoff := retry.WithMaxRetries(podRetrievalMaxRetries, retry.NewConstant(podRetrievalInterval))
	err = retry.Do(ctx, backoff, func(ctx context.Context) error {
		containerCount, initContainersMap, annotations, k8sErr = getKubePodInfo(ctx, client, string(k8sArgs.K8S_POD_NAME), string(k8sArgs.K8S_POD_NAMESPACE))
		if k8sErr != nil {
			log.Error(k8sErr, "error getting pod info", "retries", podRetrievalMaxRetries)
			return retry.RetryableError(k8sErr)
		}
		log.V(1).Info("pod container info",
			"count count", containerCount,
			"initContainers", initContainersMap,
			"annotations", annotations,
		)
		return nil
	})
	if err != nil {
		return 0, nil, nil, errors.Wrap(err, "failed to get pod data")
	}

	return containerCount, initContainersMap, annotations, nil
}

func shouldExcludePod(excludedNamespaces []string, podNamespace types.UnmarshallableString) bool {
	excludePod := false
	for _, excludeNs := range excludedNamespaces {
		if string(podNamespace) == excludeNs {
			excludePod = true
			break
		}
	}
	return excludePod
}

func logPrevResult(conf *PluginConf) {
	var loggedPrevResult interface{}
	if conf.PrevResult == nil {
		loggedPrevResult = "none"
	} else {
		loggedPrevResult = conf.PrevResult
	}

	log.V(1).Info("cmdAdd config parsed", "version", conf.CNIVersion, "prevResult", loggedPrevResult)
}

func cmdCheck(*skel.CmdArgs) error {
	return nil
}

// cmdDel is called for DELETE requests
func cmdDel(*skel.CmdArgs) error {
	return nil
}

func Run() {
	skel.PluginMainFuncs(
		skel.CNIFuncs{
			Add:   cmdAdd,
			Del:   cmdDel,
			Check: cmdCheck,
		},
		version.All,
		fmt.Sprintf("kuma-cni %v", kuma_version.Build.Version),
	)
}
