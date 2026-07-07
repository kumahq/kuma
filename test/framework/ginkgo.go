package framework

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/v3/test/framework/report"
	"github.com/kumahq/kuma/v3/test/framework/versions"
)

var suiteFailed bool

const redactedReproValue = "[redacted]"

func ShouldSkipCleanup() bool {
	suiteConfig, _ := ginkgo.GinkgoConfiguration()

	return (suiteFailed || ginkgo.CurrentSpecReport().Failed()) && suiteConfig.FailFast
}

func doIfNoSkipCleanup(fn func()) func() {
	return func() {
		if ShouldSkipCleanup() {
			return
		}

		fn()
	}
}

func AfterEachFailure(fn func()) bool {
	return ginkgo.JustAfterEach(func() {
		if !ginkgo.CurrentSpecReport().Failed() {
			return
		}
		fn()
	})
}

func E2EAfterEach(fn func()) bool {
	return ginkgo.AfterEach(doIfNoSkipCleanup(fn))
}

func E2EAfterAll(fn func()) bool {
	return ginkgo.AfterAll(doIfNoSkipCleanup(fn))
}

func E2EAfterSuite(fn func()) bool {
	return ginkgo.AfterSuite(doIfNoSkipCleanup(fn))
}

func E2ESynchronizedBeforeSuite(process1Body any, allProcessBody any, args ...any) bool {
	ginkgo.AfterEach(func() {
		if ginkgo.CurrentSpecReport().Failed() {
			suiteFailed = true
			addReproManifestToReport()
		}
	})
	return ginkgo.SynchronizedBeforeSuite(process1Body, allProcessBody, args...)
}

func E2EBeforeSuite(fn func()) bool {
	ginkgo.AfterEach(func() {
		if ginkgo.CurrentSpecReport().Failed() {
			suiteFailed = true
			addReproManifestToReport()
		}
	})

	return ginkgo.BeforeSuite(func() {
		fn()
	})
}

func E2EDeferCleanup(args ...any) {
	callback := reflect.ValueOf(args[0])
	if callback.Kind() != reflect.Func || callback.Type().NumOut() > 1 {
		ginkgo.Fail(fmt.Sprintf(
			"first argument in E2EDeferCleanup must be a function and is %T instead",
			args[0],
		))
	}

	fn := func(args []any) error {
		if ShouldSkipCleanup() {
			return nil
		}

		var callArgs []reflect.Value
		for _, arg := range args {
			callArgs = append(callArgs, reflect.ValueOf(arg))
		}

		out := callback.Call(callArgs)
		if len(out) > 0 && !out[len(out)-1].IsNil() {
			return out[len(out)-1].Interface().(error)
		}

		return nil
	}

	ginkgo.DeferCleanup(fn, args[1:])
}

func SupportedVersionEntries() []ginkgo.TableEntry {
	ginkgo.GinkgoHelper()
	var res []ginkgo.TableEntry
	for _, v := range versions.UpgradableVersionsFromBuild(Config.SupportedVersions()) {
		res = append(res, ginkgo.Entry(nil, v))
	}
	return res
}

func addReproManifestToReport() {
	ginkgo.GinkgoHelper()

	spec := ginkgo.CurrentSpecReport()
	suiteConfig, reporterConfig := ginkgo.GinkgoConfiguration()

	manifest := map[string]any{
		"spec": map[string]any{
			"fullText":        spec.FullText(),
			"labels":          spec.Labels(),
			"leafNodeType":    spec.LeafNodeType.String(),
			"leafNodeText":    spec.LeafNodeText,
			"location":        spec.LeafNodeLocation.String(),
			"state":           spec.State.String(),
			"startTime":       spec.StartTime,
			"endTime":         spec.EndTime,
			"runTime":         spec.RunTime.String(),
			"parallelProcess": spec.ParallelProcess,
		},
		"ginkgo": map[string]any{
			"randomSeed":            suiteConfig.RandomSeed,
			"focusStrings":          suiteConfig.FocusStrings,
			"skipStrings":           suiteConfig.SkipStrings,
			"labelFilter":           suiteConfig.LabelFilter,
			"semVerFilter":          suiteConfig.SemVerFilter,
			"failFast":              suiteConfig.FailFast,
			"flakeAttempts":         suiteConfig.FlakeAttempts,
			"mustPassRepeatedly":    suiteConfig.MustPassRepeatedly,
			"outputInterceptorMode": suiteConfig.OutputInterceptorMode,
			"parallelProcess":       suiteConfig.ParallelProcess,
			"parallelTotal":         suiteConfig.ParallelTotal,
			"githubOutput":          reporterConfig.GithubOutput,
			"jsonReport":            reporterConfig.JSONReport,
			"junitReport":           reporterConfig.JUnitReport,
		},
		"runtime": map[string]any{
			"goos":       runtime.GOOS,
			"goarch":     runtime.GOARCH,
			"goVersion":  runtime.Version(),
			"numCPU":     runtime.NumCPU(),
			"workingDir": getwd(),
		},
		"env":   filteredReproEnv(),
		"tools": toolVersions(),
	}

	if Config != nil {
		manifest["frameworkConfig"] = map[string]any{
			"kumaImageRegistry":                 Config.KumaImageRegistry,
			"kumaImageTag":                      Config.KumaImageTag,
			"kumaNamespace":                     Config.KumaNamespace,
			"k8sType":                           Config.K8sType,
			"ipv6":                              Config.IPV6,
			"arch":                              Config.Arch,
			"os":                                Config.OS,
			"defaultClusterStartupRetries":      Config.DefaultClusterStartupRetries,
			"defaultClusterStartupTimeout":      Config.DefaultClusterStartupTimeout.String(),
			"kumaExperimentalSidecarContainers": Config.KumaExperimentalSidecarContainers,
			"dumpDir":                           Config.DumpDir,
			"dumpOnSuccess":                     Config.DumpOnSuccess,
		}
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		report.AddFileToReportEntry("repro-error.txt", fmt.Sprintf("failed to marshal repro manifest: %v", err))
		return
	}
	report.AddFileToReportEntry("repro.json", data)
}

func filteredReproEnv() map[string]string {
	return filteredReproEnvFrom(os.Environ())
}

func filteredReproEnvFrom(environ []string) map[string]string {
	env := map[string]string{}
	for _, pair := range environ {
		key, value, ok := strings.Cut(pair, "=")
		if !ok || !isReproEnvKey(key) {
			continue
		}
		env[key] = reproEnvValue(key, value)
	}
	return env
}

func isReproEnvKey(key string) bool {
	if slices.Contains([]string{
		"ARCH",
		"CI",
		"CI_K3S_VERSION",
		"DOCKER_NETWORK",
		"GINKGO_E2E_TEST_FLAGS",
		"GINKGO_OPTS",
		"IPV6",
		"K8S_CLUSTER_TOOL",
		"K8SCLUSTERS",
		"K3D_NETWORK_CNI",
		"K3S_VERSION",
		"KIND",
		"KIND_EXPERIMENTAL_DOCKER_NETWORK",
		"OS",
		"PORT_PREFIX",
	}, key) {
		return true
	}
	return strings.HasPrefix(key, "GITHUB_") ||
		strings.HasPrefix(key, "KUMA_") ||
		strings.HasPrefix(key, "KUBECONFIG") ||
		strings.HasPrefix(key, "KIND_")
}

func reproEnvValue(key, value string) string {
	if isSafeReproEnvValueKey(key) {
		return value
	}
	return redactedReproValue
}

func isSafeReproEnvValueKey(key string) bool {
	if slices.Contains([]string{
		"ARCH",
		"CI",
		"CI_K3S_VERSION",
		"DOCKER_NETWORK",
		"GINKGO_E2E_TEST_FLAGS",
		"GINKGO_OPTS",
		"IPV6",
		"K8S_CLUSTER_TOOL",
		"K8SCLUSTERS",
		"K3D_NETWORK_CNI",
		"K3S_VERSION",
		"KIND",
		"KIND_EXPERIMENTAL_DOCKER_NETWORK",
		"KUBECONFIG",
		"KUMA_CNI_IMAGE_REPOSITORY",
		"KUMA_CP_IMAGE_REPOSITORY",
		"KUMA_DEBUG",
		"KUMA_DEFAULT_RETRIES",
		"KUMA_DEFAULT_TIMEOUT",
		"KUMA_DP_IMAGE_REPOSITORY",
		"KUMA_DUMP_DIR",
		"KUMA_EXPERIMENTAL_SIDECAR_CONTAINERS",
		"KUMA_GLOBAL_IMAGE_REGISTRY",
		"KUMA_GLOBAL_IMAGE_TAG",
		"KUMA_INIT_IMAGE_REPOSITORY",
		"KUMA_K8S_TYPE",
		"KUMA_USE_HOSTNAME_INSTEAD_OF_ID",
		"KUMA_USE_LOAD_BALANCER",
		"KUMA_ZONE_EGRESS_APP",
		"KUMA_ZONE_INGRESS_APP",
		"OS",
		"PORT_PREFIX",
	}, key) {
		return true
	}

	for _, prefix := range []string{
		"GITHUB_ACTION",
		"GITHUB_ACTOR",
		"GITHUB_API_URL",
		"GITHUB_BASE_REF",
		"GITHUB_ENV",
		"GITHUB_EVENT_NAME",
		"GITHUB_GRAPHQL_URL",
		"GITHUB_HEAD_REF",
		"GITHUB_JOB",
		"GITHUB_OUTPUT",
		"GITHUB_PATH",
		"GITHUB_REF",
		"GITHUB_REPOSITORY",
		"GITHUB_RETENTION_DAYS",
		"GITHUB_RUN_",
		"GITHUB_SERVER_URL",
		"GITHUB_SHA",
		"GITHUB_STEP_SUMMARY",
		"GITHUB_TRIGGERING_ACTOR",
		"GITHUB_WORKFLOW",
		"GITHUB_WORKSPACE",
		"KUBECONFIG",
	} {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}

	return false
}

func toolVersions() map[string]string {
	tools := map[string][]string{
		"docker":  {"version", "--format", "{{.Client.Version}} / {{.Server.Version}}"},
		"kind":    {"version"},
		"k3d":     {"version"},
		"kubectl": {"version", "--client=true"},
		"go":      {"version"},
	}

	versions := map[string]string{}
	for tool, args := range tools {
		versions[tool] = commandOutput(2*time.Second, tool, args...)
	}
	return versions
}

func commandOutput(timeout time.Duration, name string, args ...string) string {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		return ctx.Err().Error()
	}
	if err != nil {
		return fmt.Sprintf("%v: %s", err, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out))
}

func getwd() string {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Sprintf("unknown: %v", err)
	}
	return wd
}
