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
	return envBool("KUMA_GLOBAL_IMAGE_REGISTRY")
}

func GetGlobalImageTag() string {
	return os.Getenv("KUMA_GLOBAL_IMAGE_TAG")
}

func HasGlobalImageTag() bool {
	return envBool("KUMA_GLOBAL_IMAGE_TAG")
}

func GetCpImageRegistry() string {
	return os.Getenv("KUMA_CP_IMAGE_REPOSITORY")
}

func HasCpImageRegistry() bool {
	return envBool("KUMA_CP_IMAGE_REPOSITORY")
}

func GetDpImageRegistry() string {
	return os.Getenv("KUMA_DP_IMAGE_REPOSITORY")
}

func HasDpImageRegistry() bool {
	return envBool("KUMA_DP_IMAGE_REPOSITORY")
}

func GetDpInitImageRegistry() string {
	return os.Getenv("KUMA_DP_INIT_IMAGE_REPOSITORY")
}

func HasDpInitImageRegistry() bool {
	return envBool("KUMA_DP_INIT_IMAGE_REPOSITORY")
}

func GetUniversalImage() string {
	if envBool("KUMA_UNIVERSAL_IMAGE") {
		return os.Getenv("KUMA_UNIVERSAL_IMAGE")
	}

	return "kuma-universal"
}

func GetApiVersion() string {
	return os.Getenv(envAPIVersion)
}

func HasApiVersion() bool {
	return envBool(envAPIVersion)
}

func GetHelmChartPath() string {
	return os.Getenv("HELM_CHART_PATH")
}

func HasHelmChartPath() bool {
	return envBool("HELM_CHART_PATH")
}

func GetCniConfName() string {
	return os.Getenv("KUMA_CNI_CONF_NAME")
}

func HasCniConfName() bool {
	return envBool("KUMA_CNI_CONF_NAME")
}

func UseLoadBalancer() bool {
	return envBool("KUMA_USE_LOAD_BALANCER")
}

func IsInEKS() bool {
	return envBool("KUMA_IN_EKS")
}

func IsIPv6() bool {
	return envBool(envIPv6)
}

func GetKumactlBin() string {
	return os.Getenv(envKUMACTLBIN)
}

func IsK8sClustersStarted() bool {
	_, found := os.LookupEnv(envK8SCLUSTERS)
	return found
}

func envBool(env string) bool {
	value, found := os.LookupEnv(env)
	return found && strings.ToLower(value) == "true"
}
