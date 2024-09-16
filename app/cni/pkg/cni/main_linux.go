package cni

import (
	"bufio"
	"bytes"
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

	"github.com/kumahq/kuma/app/cni/pkg/install"
	"github.com/kumahq/kuma/pkg/core"
	kuma_log "github.com/kumahq/kuma/pkg/log"
	k8s_metadata "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	tproxy_k8s "github.com/kumahq/kuma/pkg/transparentproxy/kubernetes"
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
		return nil, errors.Wrap(err, "failed to load CNI arguments")
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

	annotations, ok, err := getAndValidatePodAnnotations(ctx, logger, conf, k8sArgs)
	if err != nil && !ok {
		return nil, errors.Wrap(err, "failed to get pod annotations")
	} else if err != nil {
		logger.Info("pod excluded", "reason", err)
		return conf, nil
	}

	if v := annotations[k8s_metadata.KumaTrafficTransparentProxyConfig]; v != "" {
		var logBuffer bytes.Buffer
		logWriter := bufio.NewWriter(&logBuffer)
		defer logWriter.Flush()

		cfg, err := tproxy_k8s.ConfigFromAnnotations(annotations)
		if err != nil {
			return nil, errors.Wrap(err, "failed to retrieve transparent proxy configuration")
		}

		if err := injectIptables(ctx, args.Netns, cfg.WithStdout(logWriter)); err != nil {
			return nil, errors.Wrap(err, "failed to inject iptables rules")
		}

		logger.V(1).Info("generated iptables rules", "output", logBuffer.String())
	} else if intermediate, err := NewIntermediateConfig(annotations); err != nil {
		return nil, errors.Wrap(err, "pod intermediate config failed due to bad params")
	} else if err := legacyInjectIptables(ctx, args.Netns, intermediate, logger); err != nil {
		return nil, errors.Wrap(err, "could not inject rules into namespace")
	}

	logger.Info("successfully injected iptables rules")

	return conf, nil
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
