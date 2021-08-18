// +build !windows

package bootstrap

func DefaultBootstrapParamsConfig() *BootstrapParamsConfig {
	return buildDefaultBootstrapParamsConfig("/dev/null")
}
