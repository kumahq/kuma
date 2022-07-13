package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/cni/pkg/version"

	"github.com/kumahq/kuma/pkg/core"
)

const (
	podRetrievalMaxRetries = 30
	podRetrievalInterval   = 1 * time.Second
)

var (
	log = core.NewLoggerWithRotation(2, "/tmp/kuma-cni.log", 100, 0, 0).WithName("kuma-cni")
)

// Kubernetes a K8s specific struct to hold config
type Kubernetes struct {
	Kubeconfig        string   `json:"kubeconfig"`
	ExcludeNamespaces []string `json:"exclude_namespaces"`
	CniBinDir         string   `json:"cni_bin_dir"`
}

type PluginConf struct {
	types.NetConf // You may wish to not nest this type

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

// parseConfig parses the supplied configuration (and prevResult) from stdin.
func parseConfig(stdin []byte) (*PluginConf, error) {
	conf := PluginConf{}

	if err := json.Unmarshal(stdin, &conf); err != nil {
		return nil, fmt.Errorf("failed to parse network configuration: %v", err)
	}

	// Parse previous result. Remove this if your plugin is not chained.
	if conf.RawPrevResult != nil {
		resultBytes, err := json.Marshal(conf.RawPrevResult)
		if err != nil {
			return nil, fmt.Errorf("could not serialize prevResult: %v", err)
		}
		res, err := version.NewResult(conf.CNIVersion, resultBytes)
		if err != nil {
			return nil, fmt.Errorf("could not parse prevResult: %v", err)
		}
		conf.RawPrevResult = nil
		conf.PrevResult, err = current.NewResultFromResult(res)
		if err != nil {
			return nil, fmt.Errorf("could not convert result to current version: %v", err)
		}
	}
	// End previous result parsing

	return &conf, nil
}

// cmdAdd is called for ADD requests
func cmdAdd(args *skel.CmdArgs) error {
	ctx := context.Background()
	conf, err := parseConfig(args.StdinData)
	if err != nil {
		log.Error(err, "error parsing kuma-cni cmdAdd config")
		return err
	}
	logPrevResult(conf)

	// Determine if running under k8s by checking the CNI args
	k8sArgs := K8sArgs{}
	if err := types.LoadArgs(args.Args, &k8sArgs); err != nil {
		log.Error(err, "error loading kuma-cni cmdAdd args")
		return err
	}
	logContainerInfo(args.Args, args.ContainerID, k8sArgs)

	if string(k8sArgs.K8S_POD_NAMESPACE) == "" || string(k8sArgs.K8S_POD_NAME) != "" {
		log.Info("no kubernetes data")
		return prepareResult(conf)
	}

	if shouldExcludePod(conf.Kubernetes.ExcludeNamespaces, k8sArgs.K8S_POD_NAMESPACE) {
		log.Info("pod excluded")
		return prepareResult(conf)
	}

	containers, initContainersMap, annotations, err := getPodInfoWithRetries(ctx, conf, k8sArgs)
	if err != nil {
		log.Info("error getting pod info")
		return err
	}

	if isInitContainerPresent(initContainersMap, k8sArgs) {
		log.Info("pod excluded - init container present")
		return prepareResult(conf)
	}

	if containers < 2 {
		log.Info("not enough containers in pod")
		return prepareResult(conf)
	}

	logAnnotations(args, k8sArgs, annotations)
	if isSidecarInjectedAnnotationAbsent(annotations) {
		log.Info("no sidecar injected annotation")
		return prepareResult(conf)
	}

	if intermediateConfig, configErr := NewIntermediateConfig(annotations); configErr != nil {
		log.Error(configErr, "pod intermediateConfig failed due to bad params")
	} else {
		if err := Inject(args.Netns, intermediateConfig); err != nil {
			log.Error(err, "could not inject rules into namespace")
			return err
		}
	}
	return prepareResult(conf)
}

func prepareResult(conf *PluginConf) error {
	var result *current.Result
	if conf.PrevResult == nil {
		result = &current.Result{
			CNIVersion: current.ImplementedSpecVersion,
		}
	} else {
		// Pass through the result for the next plugin
		result = conf.PrevResult
	}
	log.Info("Result: %v", "result", result)
	return types.PrintResult(result, conf.CNIVersion)
}

func logAnnotations(args *skel.CmdArgs, k8sArgs K8sArgs, annotations map[string]string) {
	log.V(1).Info("checking annotations prior to injecting redirect",
		"containerID", args.ContainerID,
		"netns", args.Netns,
		"pod", string(k8sArgs.K8S_POD_NAME),
		"namespace", string(k8sArgs.K8S_POD_NAMESPACE),
		"annotations", annotations)
}

func isSidecarInjectedAnnotationAbsent(annotations map[string]string) bool {
	excludePod := false
	val, ok := annotations["kuma.io/sidecar-injected"]
	if !ok || val != "true" {
		log.V(1).Info("pod excluded due to lack of 'kuma.io/sidecar-injected: true' annotation")
		excludePod = true
	}
	return excludePod
}

func isInitContainerPresent(initContainersMap map[string]struct{}, k8sArgs K8sArgs) bool {
	excludePod := false
	// Check if kuma-init container is present; in that case exclude pod
	if _, present := initContainersMap["kuma-init"]; present {
		log.V(1).Info("pod excluded due to being already injected with kuma-init container",
			"pod", string(k8sArgs.K8S_POD_NAME),
			"namespace", string(k8sArgs.K8S_POD_NAMESPACE))
		excludePod = true
	}
	return excludePod
}

func getPodInfoWithRetries(ctx context.Context, conf *PluginConf, k8sArgs K8sArgs) (int, map[string]struct{}, map[string]string, error) {
	client, err := newKubeClient(*conf)
	if err != nil {
		return 0, nil, nil, err
	}
	log.V(1).Info("created Kubernetes client", "client", client)

	var containerCount int
	var initContainersMap map[string]struct{}
	var annotations map[string]string
	var k8sErr error
	for attempt := 1; attempt <= podRetrievalMaxRetries; attempt++ {
		containerCount, initContainersMap, annotations, k8sErr = getKubePodInfo(ctx, client, string(k8sArgs.K8S_POD_NAME), string(k8sArgs.K8S_POD_NAMESPACE))
		log.V(1).Info("container count in a pod", "count", containerCount)
		if k8sErr == nil {
			break
		}
		log.Error(k8sErr, "waiting for pod metadata", "attempt", attempt)
		time.Sleep(podRetrievalInterval)
	}
	if k8sErr != nil {
		log.Error(k8sErr, "failed to get pod data")
		return 0, nil, nil, k8sErr
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

func logContainerInfo(args string, containerID string, k8sArgs K8sArgs) {
	log.Info("getting identifiers with arguments: %s", "arguments", args)
	log.Info("loaded k8s arguments: %v", "k8s args", k8sArgs)
	log.Info("container information",
		"containerID", containerID,
		"pod", string(k8sArgs.K8S_POD_NAME),
		"namespace", string(k8sArgs.K8S_POD_NAMESPACE))
}

func logPrevResult(conf *PluginConf) {
	var loggedPrevResult interface{}
	if conf.PrevResult == nil {
		loggedPrevResult = "none"
	} else {
		loggedPrevResult = conf.PrevResult
	}

	log.V(1).Info("cmdAdd config parsed",
		"version", conf.CNIVersion,
		"prevResult", loggedPrevResult)
}

func cmdGet(args *skel.CmdArgs) error {
	log.Info("cmdGet not implemented")
	// TODO: implement
	return nil
}

// cmdDel is called for DELETE requests
func cmdDel(args *skel.CmdArgs) error {
	log.Info("cmdDel not implemented")
	// TODO: implement
	return nil
}

func main() {
	// TODO: implement plugin version
	skel.PluginMain(cmdAdd, cmdGet, cmdDel, version.All, "kuma-cni")
}
