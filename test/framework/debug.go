package framework

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"slices"
	"strings"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"
	"go.uber.org/multierr"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kumahq/kuma/v3/test/framework/kumactl"
	"github.com/kumahq/kuma/v3/test/framework/report"
	"github.com/kumahq/kuma/v3/test/framework/utils"
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
	assertNoXdsNacks(cluster)
	assertNoXdsChurn(cluster, logs)
}

// assertNoXdsChurn parses the CP logs and fails if any proxy's xDS config
// was regenerated with a repeated, byte-identical content hash. That
// signals non-deterministic xDS generation (e.g. unordered map iteration),
// which forces Envoy to rebuild and warm the affected resources repeatedly
// even though nothing about the proxy's configuration actually changed.
func assertNoXdsChurn(cluster Cluster, logs map[string]string) {
	ginkgo.GinkgoHelper()
	for k, log := range logs {
		reports := utils.DetectXdsChurn(log)
		Expect(reports).To(BeEmpty(), "CP %s has xDS churn in logs %s: %v", cluster.Name(), k, reports)
	}
}

// nackCounter is a CP counter family that records a rejected config, the
// label pair that marks a NACK series inside that family, and how many NACKs
// are tolerated before the suite fails.
type nackCounter struct {
	family string
	// label and value select the NACK series. Empty label means every
	// series in the family is a NACK by construction.
	label string
	value string
	// skipErrorType exempts series carrying this error_type. Used for
	// "user", which marks a NACK the user caused rather than a CP defect.
	skipErrorType string
	// maxTolerated is the highest value that still passes.
	maxTolerated float64
}

// nackCounters covers both the proxy-facing xDS servers and the KDS streams
// between zone and global. The two sides are separate metric families because
// they come from separate servers:
//   - "xds"/"delta_xds" prefixes: pkg/xds/server/components.go, a NACK means
//     Envoy rejected a config this CP generated.
//   - "kds_delta" prefix: pkg/kds/server/components.go, a NACK means the peer
//     CP rejected a resource this CP sent over KDS.
//   - kds_nack_total: pkg/kds/mux/zone_sync.go, a NACK this CP sent back
//     because a resource the peer sent failed validation. Every series is a
//     NACK, and the family is absent until the first one.
//
// KDS tolerates nothing. A KDS NACK is not a transient eventual-consistency
// blip that resolves itself: the sender keeps resending the same rejected
// resource on every resync, so the stream never converges and the two CPs
// disagree about state for as long as the resource exists. The proxy-facing
// thresholds are left as they were.
//
// The exception is error_type="user", which KDS sets when the peer sent a
// resource that conflicts with one the user already created locally. The CP
// skips it and NACKs on purpose (pkg/kds/store/sync.go), and
// test/e2e_env/multizone/sync asserts exactly that, so it is not a defect.
var nackCounters = []nackCounter{
	{family: "xds_requests_received", label: "confirmation", value: "NACK", maxTolerated: 2},
	{family: "delta_xds_requests_received", label: "confirmation", value: "NACK", maxTolerated: 2},
	{family: "kds_delta_requests_received", label: "confirmation", value: "NACK", skipErrorType: "user", maxTolerated: 0},
	{family: "kds_nack_total", maxTolerated: 0},
}

// assertNoXdsNacks scrapes the CP /metrics endpoint and fails if any xDS or
// KDS request has been answered with a NACK. A non-zero counter means the CP
// produced a config that Envoy or a peer control plane rejected.
func assertNoXdsNacks(cluster Cluster) {
	ginkgo.GinkgoHelper()
	raw, err := cluster.GetKuma().GetMetrics()
	Expect(err).ToNot(HaveOccurred(), "failed to scrape metrics from CP %s", cluster.Name())

	nacks, err := findNacks(raw)
	Expect(err).ToNot(HaveOccurred(), "failed to parse metrics from CP %s", cluster.Name())
	Expect(nacks).To(BeEmpty(), "CP %s reported NACK(s): %s", cluster.Name(), strings.Join(nacks, ", "))
}

// findNacks returns one description per NACK counter that exceeds its
// tolerance in a Prometheus text exposition.
func findNacks(raw string) ([]string, error) {
	parser := expfmt.NewTextParser(model.UTF8Validation)
	families, err := parser.TextToMetricFamilies(strings.NewReader(raw))
	if err != nil {
		return nil, err
	}

	var nacks []string
	for _, counter := range nackCounters {
		family, ok := families[counter.family]
		if !ok {
			continue
		}
		for _, metric := range family.GetMetric() {
			if counter.label != "" && !hasLabel(metric.GetLabel(), counter.label, counter.value) {
				continue
			}
			if counter.skipErrorType != "" && hasLabel(metric.GetLabel(), "error_type", counter.skipErrorType) {
				continue
			}
			value := metric.GetCounter().GetValue()
			if value <= counter.maxTolerated {
				continue
			}
			nacks = append(nacks, fmt.Sprintf("%s%s = %v (tolerated %v)",
				counter.family, formatLabels(metric.GetLabel()), value, counter.maxTolerated))
		}
	}
	return nacks, nil
}

func hasLabel(labels []*io_prometheus_client.LabelPair, name string, value string) bool {
	for _, label := range labels {
		if label.GetName() == name && label.GetValue() == value {
			return true
		}
	}
	return false
}

func formatLabels(labels []*io_prometheus_client.LabelPair) string {
	parts := make([]string, 0, len(labels))
	for _, label := range labels {
		parts = append(parts, fmt.Sprintf("%s=%q", label.GetName(), label.GetValue()))
	}
	return "{" + strings.Join(parts, ",") + "}"
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
		inspectDataplane(&kumactlOpts, cluster, mesh),
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
	out, err := k8s.RunKubectlAndGetOutputContextE(cluster.GetTesting(), context.Background(), &defaultKubeOptions, "get", "pods", "-A")
	if err != nil {
		errs = multierr.Combine(errs, fmt.Errorf("failed to get pods, %w", err))
	} else {
		report.AddFileToReportEntry(path.Join(cluster.Name(), "pods.txt"), out)
	}

	Logf("debug nodes and print resource usage of cluster %q", cluster.Name())
	nodes, err := k8s.GetNodesContextE(cluster.GetTesting(), context.Background(), &defaultKubeOptions)
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
			out, err := k8s.RunKubectlAndGetOutputContextE(cluster.GetTesting(), context.Background(), &defaultKubeOptions, "describe", "node", node.Name)
			if err != nil {
				errs = multierr.Combine(errs, fmt.Errorf("failed to describe node %s, %w", node.Name, err))
			} else {
				report.AddFileToReportEntry(path.Join(cluster.Name(), "k8s", fmt.Sprintf("node-%s.txt", node.Name)), out)
			}
		}
	}
	Logf("printing debug information of cluster %q for mesh %q and namespaces %q", cluster.Name(), mesh, namespaces)
	for _, namespace := range namespaces {
		if err := debugKubeNamespace(cluster, namespace); err != nil {
			errs = multierr.Combine(errs, fmt.Errorf("failed to debug namespace %s, %w", namespace, err))
		}
	}
	return errs
}

func debugKubeNamespace(cluster Cluster, namespace string) error {
	Logf("debug namespace %q of cluster %q", namespace, cluster.Name())
	var errs error
	kubeOptions := *cluster.GetKubectlOptions(namespace) // copy to not override fields globally
	kubeOptions.Logger = logger.Discard                  // to not print on stdout
	out, err := k8s.RunKubectlAndGetOutputContextE(cluster.GetTesting(), context.Background(), &kubeOptions, "get", "all,kuma", "-oyaml")
	if err != nil {
		errs = multierr.Append(errs, fmt.Errorf("kubectl get for namespace %s failed with error: %w", namespace, err))
	}

	// Ignore it if we don't have Gateway API resources installed
	gatewayAPIOut, err := k8s.RunKubectlAndGetOutputContextE(cluster.GetTesting(), context.Background(), &kubeOptions, "get", "gateway-api", "-oyaml")
	if err == nil {
		out += gatewayAPIOut
	} else {
		Logf("Gateway API CRDs not installed in cluster %q", cluster.Name())
	}
	// Per namespace, like the events below: this function runs once for every
	// namespace and a shared path would leave only the last one's manifests.
	report.AddFileToReportEntry(path.Join(cluster.Name(), "k8s", namespace, "manifests.yaml"), out)

	events, err := k8s.RunKubectlAndGetOutputContextE(cluster.GetTesting(), context.Background(), &kubeOptions, "get", "events", "--sort-by=.lastTimestamp", "-owide")
	if err != nil {
		errs = multierr.Append(errs, fmt.Errorf("failed to get events for namespace %s, %w", namespace, err))
	} else {
		report.AddFileToReportEntry(path.Join(cluster.Name(), "k8s", namespace, "events.txt"), events)
	}

	deployments, err := k8s.ListDeploymentsContextE(cluster.GetTesting(), context.Background(), &kubeOptions, kube_meta.ListOptions{})
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

func inspectDataplane(kumactlOpts *kumactl.KumactlOptions, cluster Cluster, mesh string) error {
	var errs error
	dpListJson, err := kumactlOpts.RunKumactlAndGetOutput("get", "dataplanes", "--mesh", mesh, "-ojson")
	if err != nil {
		return fmt.Errorf("failed to retrieve dataplanes, %w", err)
	}
	dpResp := dataplaneListResponse{}
	if jsonErr := json.Unmarshal([]byte(dpListJson), &dpResp); jsonErr != nil {
		return fmt.Errorf("failed to unmarshall dataplanes, %w", jsonErr)
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
			dpName := dpObj.Name
			args := []string{"inspect", "dataplane", dpName, "--type", inspectType}
			if inspectType == "get" {
				args = []string{"get", "dataplane", dpName, "-oyaml"}
			}
			args = append(args, "--mesh", mesh)
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
