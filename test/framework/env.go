package framework

import (
	"os"
	"strconv"
	"strings"
	"time"
)

func GetDefaultRetries() int {
	retries := DefaultRetries

	if r := os.Getenv("KUMA_DEFAULT_RETRIES"); r != "" {
		if r, err := strconv.Atoi(r); err != nil {
			retries = r
		}
	}

	return retries
}

func GetDefaultTimeout() time.Duration {
	timeout := DefaultTimeout

	if t := os.Getenv("KUMA_DEFAULT_TIMEOUT"); t != "" {
		if t, err := time.ParseDuration(t); err != nil {
			timeout = t
		}
	}

	return timeout
}

func GetGlobalImageRegistry() string {
	return os.Getenv("KUMA_GLOBAL_IMAGE_REGISTRY")
}

func HasGlobalImageRegistry() bool {
	return envIsPresent("KUMA_GLOBAL_IMAGE_REGISTRY")
}

func GetGlobalImageTag() string {
	return os.Getenv("KUMA_GLOBAL_IMAGE_TAG")
}

func HasGlobalImageTag() bool {
	return envIsPresent("KUMA_GLOBAL_IMAGE_TAG")
}

func GetCpImageRegistry() string {
	return os.Getenv("KUMA_CP_IMAGE_REPOSITORY")
}

func HasCpImageRegistry() bool {
	return envIsPresent("KUMA_CP_IMAGE_REPOSITORY")
}

func GetDpImageRegistry() string {
	return os.Getenv("KUMA_DP_IMAGE_REPOSITORY")
}

func HasDpImageRegistry() bool {
	return envIsPresent("KUMA_DP_IMAGE_REPOSITORY")
}

func GetDpInitImageRegistry() string {
	return os.Getenv("KUMA_DP_INIT_IMAGE_REPOSITORY")
}

func HasDpInitImageRegistry() bool {
	return envIsPresent("KUMA_DP_INIT_IMAGE_REPOSITORY")
}

func GetUniversalImage() string {
	if envIsPresent("KUMA_UNIVERSAL_IMAGE") {
		return os.Getenv("KUMA_UNIVERSAL_IMAGE")
	}

	return KumaUniversalImage
}

func GetApiVersion() string {
	return os.Getenv(envAPIVersion)
}

func HasApiVersion() bool {
	return envIsPresent(envAPIVersion)
}

func GetHelmChartPath() string {
	return os.Getenv("HELM_CHART_PATH")
}

func HasHelmChartPath() bool {
	return envIsPresent("HELM_CHART_PATH")
}

func GetCniConfName() string {
	return os.Getenv("KUMA_CNI_CONF_NAME")
}

func HasCniConfName() bool {
	return envIsPresent("KUMA_CNI_CONF_NAME")
}

func UseLoadBalancer() bool {
	return envBool("KUMA_USE_LOAD_BALANCER")
}

func UseHostnameInsteadOfIP() bool {
	return envBool("KUMA_USE_HOSTNAME_INSTEAD_OF_IP")
}

func IsIPv6() bool {
	return envBool(envIPv6)
}

func GetKumactlBin() string {
	return os.Getenv(envKUMACTLBIN)
}

func IsK8sClustersStarted() bool {
	return envIsPresent(envK8SCLUSTERS)
}

func envIsPresent(env string) bool {
	_, ok := os.LookupEnv(env)

	return ok
}

func envBool(env string) bool {
	value, found := os.LookupEnv(env)
	return found && strings.ToLower(value) == "true"
}
