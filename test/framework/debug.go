package framework

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path"
	"slices"
	"strings"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kumahq/kuma/v2/test/framework/kumactl"
	"github.com/kumahq/kuma/v2/test/framework/report"
	"github.com/kumahq/kuma/v2/test/framework/utils"
)

func ControlPlaneAssertions(cluster Cluster) {
	ginkgo.GinkgoHelper()
	defer ginkgo.GinkgoRecover() // Ensures that Ginkgo can recover from any failures
	logs := cluster.GetKumaCPLogs()
	for k, log := range logs {
		Expect(utils.HasPanicInCpLogs(log)).To(BeFalse(), fmt.Sprintf("CP %s has panic in logs %s", cluster.Name(), k))
	}
	switch cluster.(type) {
	case *UniversalCluster:
		// CP does not recover restart on universal. If it crashed, we can just check if the process is still running.
		out, _, _ := cluster.Exec("", "", AppModeCP, "ps", "aux")
		Expect(out).To(ContainSubstring("kuma-cp run"), "CP %s is not running", cluster.Name())
	case *K8sCluster:
		restartCount := RestartCount(cluster.GetKuma().(*K8sControlPlane).GetKumaCPPods())
		Expect(restartCount).To(BeZero(), "CP %s has restarted %d times", cluster.Name(), restartCount)
	default:
		ginkgo.Fail("unknown cluster")
	}
}

// DebugUniversal prints state of the cluster. Useful in case of failure.
// Ideas what we can add
// * XDS / Stats / Clusters of all DPPs (ideally in form of command that we can use on prod as well)
// * CP logs (although we print this already on failure)
func DebugUniversal(cluster Cluster, mesh string) {
	DumpState(cluster, mesh)
}

func DebugKube(cluster Cluster, mesh string, namespaces ...string) {
	DumpState(cluster, mesh, namespaces...)
}

// DumpState prints debug information of the cluster. Useful in case of failure.
// Ideally we should have Cluster keep an inventory of the namespaces and meshes it has so we don't have
// to pass them here.
// This way we'd be able to use ginkgo.ReportAfterEach
func DumpState(cluster Cluster, mesh string, namespaces ...string) {
	ginkgo.GinkgoHelper()
	switch ginkgo.CurrentSpecReport().State {
	case types.SpecStatePending, types.SpecStateSkipped:
		return
	default:
	}
	kumactlOpts := *cluster.GetKumactlOptions()
	kumactlOpts.Verbose = false
	var errs error

	debugCPLogs(cluster)
	errs = multierr.Combine(
		debugExport(cluster, &kumactlOpts),
		inspectDataplane(&kumactlOpts, cluster, mesh, dataplaneType),
		inspectDataplane(&kumactlOpts, cluster, mesh, zoneegressType),
		inspectDataplane(&kumactlOpts, cluster, mesh, zoneingressType),
	)
	switch cluster.(type) {
	case *K8sCluster:
		errs = multierr.Combine(errs, debugKube(cluster, mesh, namespaces...))
	case *UniversalCluster:
	}
	if errs != nil {
		Logf("[WARNING]: some debug commands failed %v", errs)
		report.AddFileToReportEntry("debug-errors.txt", fmt.Appendf(nil, "%v", errs))
	}
}

func DebugCPLogs(cluster Cluster) {
	ginkgo.GinkgoHelper()
	debugCPLogs(cluster)
}

func debugCPLogs(cluster Cluster) {
	logs := cluster.GetKumaCPLogs()
	for k, log := range logs {
		report.AddFileToReportEntry(path.Join(cluster.Name(), fmt.Sprintf("cp-logs-%s.log", k)), log)
	}
}

func debugExport(cluster Cluster, kumactlOpts *kumactl.KumactlOptions) error {
	var errs error

	Logf("saving export for %q", cluster.Name())

	out, err := kumactlOpts.RunKumactlAndGetOutput("export", "--profile", "all")
	if err != nil {
		wrappedErr := errors.Wrap(err, "failed to run 'kumactl export --profile all'")
		errs = multierr.Combine(err, wrappedErr)
		return errs
	}
	report.AddFileToReportEntry(path.Join(cluster.Name(), "kumactl-export.yaml"), []byte(out))
	return nil
}

func debugKube(cluster Cluster, mesh string, namespaces ...string) error {
	Logf("%s", "Kube state of cluster: "+cluster.Name())
	if !slices.Contains(namespaces, Config.KumaNamespace) {
		namespaces = append(namespaces, Config.KumaNamespace)
	}
	defaultKubeOptions := *cluster.GetKubectlOptions("default") // copy to not override fields globally
	defaultKubeOptions.Logger = logger.Discard
	var errs error
	out, err := k8s.RunKubectlAndGetOutputE(cluster.GetTesting(), &defaultKubeOptions, "get", "pods", "-A")
	if err != nil {
		errs = multierr.Combine(errs, fmt.Errorf("failed to get pods, %w", err))
	} else {
		report.AddFileToReportEntry(path.Join(cluster.Name(), "pods.txt"), out)
	}

	Logf("debug nodes and print resource usage of cluster %q", cluster.Name())
	nodes, err := k8s.GetNodesE(cluster.GetTesting(), &defaultKubeOptions)
	if err != nil {
		Logf("get nodes from cluster %q failed with error: %s", cluster.Name(), err.Error())
		errs = multierr.Combine(errs, fmt.Errorf("failed to get nodes, %w", err))
	} else {
		nodesJson, err := json.Marshal(nodes)
		if err != nil {
			errs = multierr.Combine(errs, fmt.Errorf("failed marshaling nodes %w", err))
		} else {
			report.AddFileToReportEntry(path.Join(cluster.Name(), "k8s", "nodes.json"), nodesJson)
		}
		for _, node := range nodes {
			out, err := k8s.RunKubectlAndGetOutputE(cluster.GetTesting(), &defaultKubeOptions, "describe", "node", node.Name)
			if err != nil {
				errs = multierr.Combine(errs, fmt.Errorf("failed to describe node %s, %w", node.Name, err))
			} else {
				report.AddFileToReportEntry(path.Join(cluster.Name(), "k8s", fmt.Sprintf("node-%s.txt", node.Name)), out)
			}
		}
	}
	eventsOut, err := k8s.RunKubectlAndGetOutputE(cluster.GetTesting(), &defaultKubeOptions, "get", "events", "-A", "--sort-by=.lastTimestamp")
	if err != nil {
		errs = multierr.Combine(errs, fmt.Errorf("failed to get events: %w", err))
	} else {
		report.AddFileToReportEntry(path.Join(cluster.Name(), "k8s", "events.txt"), eventsOut)
	}

	switch Config.K8sType {
	case K3dK8sType, K3dCalicoK8sType:
		if err := debugDockerNodeLogs(cluster, "k3d-"+cluster.Name()); err != nil {
			errs = multierr.Combine(errs, err)
		}
	case KindK8sType:
		if err := debugDockerNodeLogs(cluster, cluster.Name()); err != nil {
			errs = multierr.Combine(errs, err)
		}
	}

	topNodes, err := k8s.RunKubectlAndGetOutputE(cluster.GetTesting(), &defaultKubeOptions, "top", "nodes")
	if err != nil {
		Logf("kubectl top nodes not available for cluster %q: %s", cluster.Name(), err)
	} else {
		report.AddFileToReportEntry(path.Join(cluster.Name(), "k8s", "top-nodes.txt"), topNodes)
	}

	topPods, err := k8s.RunKubectlAndGetOutputE(cluster.GetTesting(), &defaultKubeOptions, "top", "pods", "-A", "--containers")
	if err != nil {
		Logf("kubectl top pods not available for cluster %q: %s", cluster.Name(), err)
	} else {
		report.AddFileToReportEntry(path.Join(cluster.Name(), "k8s", "top-pods.txt"), topPods)
	}

	Logf("printing debug information of cluster %q for mesh %q and namespaces %q", cluster.Name(), mesh, namespaces)
	for _, namespace := range namespaces {
		if err := debugKubeNamespace(cluster, namespace); err != nil {
			errs = multierr.Combine(errs, fmt.Errorf("failed to debug namespace %s, %w", namespace, err))
		}
	}
	return errs
}

func debugDockerNodeLogs(cluster Cluster, nameFilter string) error {
	Logf("collecting docker node logs for cluster %q (filter: %s)", cluster.Name(), nameFilter)
	ctx := context.Background()
	listOut, err := exec.CommandContext(ctx, "docker", "ps", "-a", "--filter", "name="+nameFilter, "--format", "{{.Names}}").Output()
	if err != nil {
		return fmt.Errorf("failed to list docker containers for cluster %q: %w", cluster.Name(), err)
	}
	var errs error
	for containerName := range strings.FieldsSeq(string(listOut)) {
		logOut, err := exec.CommandContext(ctx, "docker", "logs", "--timestamps", containerName).CombinedOutput()
		if err != nil {
			errs = multierr.Combine(errs, fmt.Errorf("failed to get docker logs for container %q: %w", containerName, err))
		} else {
			report.AddFileToReportEntry(path.Join(cluster.Name(), "docker", containerName+".log"), logOut)
		}
		if err := debugDockerNodeStats(ctx, cluster, containerName); err != nil {
			errs = multierr.Combine(errs, err)
		}
	}
	return errs
}

// debugDockerNodeStats collects CPU pressure and load info from inside a node container.
// This is used to confirm CPU starvation when metrics-server is unavailable.
func debugDockerNodeStats(ctx context.Context, cluster Cluster, containerName string) error {
	// sh script that collects:
	// - load average and CPU count (quick starvation signal)
	// - cgroup v2 CPU throttling for all kubepods (nr_throttled / throttled_usec)
	// - cgroup v1 fallback
	script := `
echo "=== uptime ==="
uptime
echo "=== nproc ==="
nproc
echo "=== /proc/loadavg ==="
cat /proc/loadavg
echo "=== /proc/stat (cpu lines) ==="
grep '^cpu' /proc/stat
echo "=== cgroup v2 cpu throttling (kubepods) ==="
find /sys/fs/cgroup/kubepods.slice -name 'cpu.stat' 2>/dev/null \
  | xargs grep -h '' 2>/dev/null \
  | grep -E 'nr_throttled|throttled_usec|nr_periods' \
  | sort | uniq -c | sort -rn
echo "=== cgroup v1 cpu throttling (kubepods) ==="
find /sys/fs/cgroup/cpu,cpuacct/kubepods -name 'cpu.stat' 2>/dev/null \
  | xargs grep -h '' 2>/dev/null \
  | grep -E 'nr_throttled|throttled_time|nr_periods' \
  | sort | uniq -c | sort -rn
`
	out, err := exec.CommandContext(ctx, "docker", "exec", containerName, "sh", "-c", script).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to collect node stats from container %q: %w", containerName, err)
	}
	report.AddFileToReportEntry(path.Join(cluster.Name(), "docker", containerName+"-stats.txt"), out)
	return nil
}

func debugKubeNamespace(cluster Cluster, namespace string) error {
	Logf("debug namespace %q of cluster %q", namespace, cluster.Name())
	var errs error
	kubeOptions := *cluster.GetKubectlOptions(namespace) // copy to not override fields globally
	kubeOptions.Logger = logger.Discard                  // to not print on stdout
	out, err := k8s.RunKubectlAndGetOutputE(cluster.GetTesting(), &kubeOptions, "get", "all,kuma", "-oyaml")
	if err != nil {
		errs = multierr.Append(errs, fmt.Errorf("kubectl get for namespace %s failed with error: %w", namespace, err))
	}

	// Ignore it if we don't have Gateway API resources installed
	gatewayAPIOut, err := k8s.RunKubectlAndGetOutputE(cluster.GetTesting(), &kubeOptions, "get", "gateway-api", "-oyaml")
	if err == nil {
		out += gatewayAPIOut
	} else {
		Logf("Gateway API CRDs not installed in cluster %q", cluster.Name())
	}
	report.AddFileToReportEntry(path.Join(cluster.Name(), "k8s", "manifests.yaml"), out)

	pods, err := k8s.ListPodsE(cluster.GetTesting(), &kubeOptions, kube_meta.ListOptions{})
	if err != nil {
		errs = multierr.Append(errs, fmt.Errorf("failed to list pods in namespace %s: %w", namespace, err))
	} else {
		for i := range pods {
			podDescribe, err := k8s.RunKubectlAndGetOutputE(cluster.GetTesting(), &kubeOptions, "describe", "pod", pods[i].Name)
			if err != nil {
				errs = multierr.Append(errs, fmt.Errorf("failed to describe pod %s in namespace %s: %w", pods[i].Name, namespace, err))
			} else {
				report.AddFileToReportEntry(path.Join(cluster.Name(), "k8s", namespace, fmt.Sprintf("pod-%s-describe.txt", pods[i].Name)), podDescribe)
			}
		}
	}

	deployments, err := k8s.ListDeploymentsE(cluster.GetTesting(), &kubeOptions, kube_meta.ListOptions{})
	if err != nil {
		errs = multierr.Append(errs, fmt.Errorf("failed to list deployments in namespace %s, %w", namespace, err))
	} else {
		for _, deployment := range deployments {
			deployDetails := ExtractDeploymentDetails(cluster.GetTesting(), &kubeOptions, deployment.Name)

			for _, pod := range deployDetails.Pods {
				for container, log := range pod.Logs {
					if log == "" {
						continue
					}

					report.AddFileToReportEntry(path.Join(cluster.Name(), "k8s", deployment.Namespace, fmt.Sprintf("pod-%s-%s.log", pod.Name, container)), log)
				}
			}

			for _, pod := range deployDetails.Pods {
				pod.Logs = map[string]string{}
			}
			deployDetailsJson := MarshalObjectDetails(deployDetails)
			report.AddFileToReportEntry(path.Join(cluster.Name(), "k8s", deployment.Namespace, fmt.Sprintf("deployment-%s.json", deployment.Name)), deployDetailsJson)
		}
	}
	return errs
}

type dpType string

const (
	dataplaneType   dpType = "dataplane"
	zoneegressType  dpType = "zoneegress"
	zoneingressType dpType = "zoneingress"
)

func inspectDataplane(kumactlOpts *kumactl.KumactlOptions, cluster Cluster, mesh string, dpType dpType) error {
	var errs error
	var args []string
	switch dpType {
	case dataplaneType:
		args = []string{"get", "dataplanes", "--mesh", mesh, "-ojson"}
	case zoneegressType:
		args = []string{"get", "zoneegresses", "-ojson"}
	case zoneingressType:
		args = []string{"get", "zone-ingresses", "-ojson"}
	default:
		panic("unknown dp type " + string(dpType))
	}
	dpListJson, err := kumactlOpts.RunKumactlAndGetOutput(args...)
	if err != nil {
		return fmt.Errorf("failed to retrieve dps of type %q, %w", dpType, err)
	}
	dpResp := dataplaneListResponse{}
	if jsonErr := json.Unmarshal([]byte(dpListJson), &dpResp); jsonErr != nil {
		return fmt.Errorf("failed to unmarshall dps of type %q, %w", dpType, err)
	}

	for _, dpObj := range dpResp.Items {
		for inspectType, fileExtension := range map[string]string{
			"get":         ".yaml",
			"config-dump": ".json",
			"config":      ".json",
			"policies":    ".txt",
			"stats":       ".txt",
			"clusters":    ".txt",
		} {
			// zoneingress and zoneegress do not have policies nor config
			if dpType != dataplaneType && slices.Contains([]string{"policies", "config"}, inspectType) {
				continue
			}

			dpName := dpObj.Name
			args := []string{"inspect", string(dpType), dpName, "--type", inspectType}
			if inspectType == "get" {
				if dpType == zoneingressType {
					args = []string{"get", "zone-ingress", dpName, "-oyaml"}
				} else {
					args = []string{"get", string(dpType), dpName, "-oyaml"}
				}
			}
			if dpType == dataplaneType {
				args = append(args, "--mesh", mesh)
			}
			inspectResp, err := kumactlOpts.RunKumactlAndGetOutput(args...)

			if err != nil {
				errs = multierr.Combine(errs, fmt.Errorf("failed to inspect %s of dp %q from cluster %q for mesh %q, %w", inspectType, dpName, cluster.Name(), mesh, err))
			} else {
				inspectFilePath := fmt.Sprintf("%s-%s-%s%s", mesh, dpName, inspectType, fileExtension)
				report.AddFileToReportEntry(path.Join(cluster.Name(), "dps", inspectFilePath), inspectResp)
			}
		}
	}
	return errs
}

type dataplaneResponse struct {
	Mesh string `json:"mesh"`
	Name string `json:"name"`
}

type dataplaneListResponse struct {
	Total int                 `json:"total"`
	Items []dataplaneResponse `json:"items"`
}
