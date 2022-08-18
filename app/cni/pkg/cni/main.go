package cni

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path"
	"strconv"
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
)

var (
	log = core.NewLoggerWithRotation(defaultLogLevel, defaultLogLocation, 100, 0, 0).WithName(defaultLogName)
)

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
	conf := PluginConf{}

	if err := json.Unmarshal(stdin, &conf); err != nil {
		return nil, errors.Wrapf(err, "could not parse network configuration")
	}

	if conf.RawPrevResult != nil {
		resultBytes, err := json.Marshal(conf.RawPrevResult)
		if err != nil {
			return nil, errors.Wrapf(err, "could not serialize prevResult")
		}
		res, err := version.NewResult(conf.CNIVersion, resultBytes)
		if err != nil {
			return nil, errors.Wrapf(err, "could not parse prevResult")
		}
		conf.RawPrevResult = nil
		conf.PrevResult, err = current.NewResultFromResult(res)
		if err != nil {
			return nil, errors.Wrapf(err, "could not convert result to current version")
		}
	}

	return &conf, nil
}

func hijackMainProcessStderr(logLevel string) (*os.File, error) {
	file, err := getCniProcessStderr()
	if err != nil {
		log.Error(err, "could not hijack main process file - continue logging to "+defaultLogLocation)
		return nil, err
	}
	log.V(0).Info("successfully hijacked stderr of cni process - logs will be available in 'kubectl logs'")
	os.Stderr = file
	if err := install.SetLogLevel(&log, logLevel, defaultLogName); err != nil {
		return file, errors.Wrap(err, "wrong set the right log level")
	}

	return file, err
}

func getCniProcessStderr() (*os.File, error) {
	pids, err := pidOf("/install-cni")
	if err != nil {
		return nil, err
	}
	if len(pids) != 1 {
		return nil, errors.New("more than one process '/install-cni' running on a node, this should not happen")
	}

	file, err := os.OpenFile(path.Join("/proc", strconv.Itoa(pids[0]), "fd", "2"), os.O_WRONLY, 0)
	return file, err
}

// cmdAdd is called for ADD requests
func cmdAdd(args *skel.CmdArgs) error {
	ctx := context.Background()
	conf, err := parseConfig(args.StdinData)
	if err != nil {
		return errors.Wrap(err, "error parsing kuma-cni cmdAdd config")
	}

	mainProcessStderr, err := hijackMainProcessStderr(conf.LogLevel)
	if mainProcessStderr != nil {
		defer mainProcessStderr.Close()
	}
	if err != nil {
		return err
	}
	logPrevResult(conf)

	// Determine if running under k8s by checking the CNI args
	k8sArgs := K8sArgs{}
	if err := types.LoadArgs(args.Args, &k8sArgs); err != nil {
		return errors.Wrap(err, "error loading kuma-cni cmdAdd args")
	}
	logger := log.WithValues(
		"pod", string(k8sArgs.K8S_POD_NAME),
		"namespace", string(k8sArgs.K8S_POD_NAMESPACE),
		"podInfraContainerId", string(k8sArgs.K8S_POD_INFRA_CONTAINER_ID),
		"ip", string(k8sArgs.IP),
		"containerId", args.ContainerID,
		"args", args.Args,
	)

	if string(k8sArgs.K8S_POD_NAMESPACE) == "" || string(k8sArgs.K8S_POD_NAME) == "" {
		logger.Info("pod excluded - no kubernetes data")
		return prepareResult(conf, logger)
	}

	if shouldExcludePod(conf.Kubernetes.ExcludeNamespaces, k8sArgs.K8S_POD_NAMESPACE) {
		logger.Info(`pod excluded - is in the namespace excluded by "exclude_namespaces"`)
		return prepareResult(conf, logger)
	}

	containerCount, initContainersMap, annotations, err := getPodInfoWithRetries(ctx, conf, k8sArgs)
	if err != nil {
		return errors.Wrap(err, "pod excluded - error getting pod info")
	}

	if isInitContainerPresent(initContainersMap) {
		logger.Info("pod excluded - already injected with kuma-init container")
		return prepareResult(conf, logger)
	}

	if containerCount < 2 {
		logger.Info("pod excluded - not enough containers in pod. Kuma-sidecar container required")
		return prepareResult(conf, logger)
	}

	logger.V(1).Info("checking annotations prior to injecting redirect",
		"netns", args.Netns,
		"annotations", annotations)
	if excludeByMissingSidecarInjectedAnnotation(annotations) {
		logger.Info("pod excluded due to lack of 'kuma.io/sidecar-injected: true' annotation")
		return prepareResult(conf, logger)
	}

	if intermediateConfig, configErr := NewIntermediateConfig(annotations); configErr != nil {
		return errors.Wrap(configErr, "pod excluded - pod intermediateConfig failed due to bad params")
	} else {
		if err := Inject(args.Netns, logger, intermediateConfig); err != nil {
			return errors.Wrap(err, "pod excluded - could not inject rules into namespace")
		}
	}
	logger.Info("successfully injected iptables rules")
	return prepareResult(conf, logger)
}

func prepareResult(conf *PluginConf, logger logr.Logger) error {
	var result *current.Result
	if conf.PrevResult == nil {
		result = &current.Result{
			CNIVersion: current.ImplementedSpecVersion,
		}
	} else {
		// Pass through the result for the next plugin
		result = conf.PrevResult
	}
	logger.Info("result", "result", result)
	return types.PrintResult(result, conf.CNIVersion)
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
	client, err := newKubeClient(*conf)
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

func cmdCheck(args *skel.CmdArgs) error {
	return nil
}

// cmdDel is called for DELETE requests
func cmdDel(args *skel.CmdArgs) error {
	return nil
}

func Run() {
	skel.PluginMain(cmdAdd, cmdCheck, cmdDel, version.All, fmt.Sprintf("kuma-cni %v", kuma_version.Build.Version))
}
