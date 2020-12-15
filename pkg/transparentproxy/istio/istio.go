package istio

import (
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	uninstall "istio.io/istio/tools/istio-clean-iptables/pkg/cmd"
	install "istio.io/istio/tools/istio-iptables/pkg/cmd"
	"istio.io/istio/tools/istio-iptables/pkg/constants"
)

type IstioTransparentProxy struct {
	// output redirection
	stdout *os.File
	stderr *os.File
	reader *os.File
}

func NewIstioTransparentProxy() *IstioTransparentProxy {
	return &IstioTransparentProxy{}
}

func (tp *IstioTransparentProxy) Setup(dryRun bool, excludeInboundPorts string) (string, error) {

	viper.Set(constants.EnvoyPort, "15001")
	viper.Set(constants.InboundCapturePort, "15006")
	viper.Set(constants.ProxyUID, "5678")
	viper.Set(constants.ProxyGID, "5678")
	viper.Set(constants.InboundInterceptionMode, "REDIRECT")
	viper.Set(constants.InboundPorts, "*")
	viper.Set(constants.LocalExcludePorts, excludeInboundPorts)
	viper.Set(constants.ServiceCidr, "*")
	viper.Set(constants.DryRun, dryRun)
	viper.Set(constants.SkipRuleApply, false)
	viper.Set(constants.RunValidation, false)

	tp.redirectStdOutStdErr()
	defer func() {
		tp.restoreStdOutStderr()
	}()

	if err := install.GetCommand().Execute(); err != nil {
		return tp.getStdOutStdErr(), errors.Wrapf(err, "setting istio")
	}

	return tp.getStdOutStdErr(), nil
}

func (tp *IstioTransparentProxy) Cleanup(dryRun bool) (string, error) {

	viper.Set(constants.DryRun, dryRun)

	tp.redirectStdOutStdErr()
	defer func() {
		tp.restoreStdOutStderr()
	}()

	if err := uninstall.GetCommand().Execute(); err != nil {
		return tp.getStdOutStdErr(), errors.Wrapf(err, "setting istio")
	}

	return tp.getStdOutStdErr(), nil
}

func (tp *IstioTransparentProxy) redirectStdOutStdErr() {
	reader, writer, err := os.Pipe()

	if err != nil {
		panic(err)
	}
	tp.stdout = os.Stdout
	tp.stderr = os.Stderr
	tp.reader = reader

	os.Stdout = writer
	os.Stderr = writer
}

func (tp *IstioTransparentProxy) getStdOutStdErr() string {
	data := make([]byte, 1*1024*1024)

	_, _ = tp.reader.Read(data)

	return string(data)
}

func (tp *IstioTransparentProxy) restoreStdOutStderr() {
	os.Stdout = tp.stdout
	os.Stderr = tp.stderr
}
