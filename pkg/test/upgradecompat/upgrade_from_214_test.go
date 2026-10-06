package upgradecompat_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/proto"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_config "github.com/kumahq/kuma/v3/pkg/config"
	kuma_cp "github.com/kumahq/kuma/v3/pkg/config/app/kuma-cp"
	kuma_dp "github.com/kumahq/kuma/v3/pkg/config/app/kuma-dp"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	tproxy_config "github.com/kumahq/kuma/v3/pkg/transparentproxy/config"
	tproxy_dp "github.com/kumahq/kuma/v3/pkg/transparentproxy/config/dataplane"
	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
)

// As a 2.14 control plane wrote it: the redirect ports live on the resource
// and nowhere else.
const dataplane214 = `
networking:
  address: 10.42.0.7
  inbound:
    - port: 8080
      tags:
        kuma.io/service: backend
  transparentProxying:
    redirectPortInbound: 15006
    redirectPortOutbound: 15001
    ipFamilyMode: DualStack
    directAccessServices:
      - "*"
`

const dataplane30 = `
networking:
  address: 10.42.0.8
  inbound:
    - port: 8080
      tags:
        kuma.io/service: backend
`

func parse(yaml string) *core_mesh.DataplaneResource {
	GinkgoHelper()
	dp := core_mesh.NewDataplaneResource()
	Expect(core_model.FromYAML([]byte(yaml), dp.Spec)).To(Succeed())
	return dp
}

var _ = Describe("upgrade from 2.14", func() {
	It("reads a Dataplane 2.14 left in the store", func() {
		dp := parse(dataplane214)

		Expect(dp.Validate()).To(Succeed())
		Expect(dp.Spec.GetNetworking().GetTransparentProxying().GetDirectAccessServices()).To(HaveLen(1))
	})

	// KDS carries the spec as binary protobuf, so a 2.14 zone sends field 6,
	// which 3.0 no longer defines.
	It("reads a KDS payload from a zone still on 2.14", func() {
		// TransparentProxying as 2.14 encodes it: 1 = 15006, 2 = 15001, 6 = DualStack.
		wire := []byte{
			0x08, 0x9e, 0x75,
			0x10, 0x99, 0x75,
			0x30, 0x01,
		}
		tp := &mesh_proto.Dataplane_Networking_TransparentProxying{}

		Expect(proto.Unmarshal(wire, tp)).To(Succeed())
		Expect(tp.GetRedirectPortInbound()).To(Equal(uint32(15006))) //nolint:staticcheck // deprecated on purpose
		Expect(util_proto.ToJSON(tp)).Error().ToNot(HaveOccurred())
	})

	// With the 2.14 default transparentProxy.configMap.enabled=false the
	// injected kuma-dp gets no --transparent-proxy-config and so reports no
	// transparent proxy of its own, while iptables keeps redirecting to these
	// ports.
	It("keeps transparent proxy for a sidecar injected by 2.14", func() {
		cfg := tproxy_dp.GetDataplaneConfig(parse(dataplane214), &core_xds.DataplaneMetadata{IPv6Enabled: true})

		Expect(cfg.Enabled()).To(BeTrue())
		Expect(cfg.Redirect.Inbound.Port).To(Equal(tproxy_config.Port(15006)))
		Expect(cfg.Redirect.Outbound.Port).To(Equal(tproxy_config.Port(15001)))
	})

	// The removed ipFamilyMode field is not consulted; what the proxy reports
	// about its machine is.
	It("does not redirect IPv6 for a proxy on a machine without it", func() {
		cfg := tproxy_dp.GetDataplaneConfig(parse(dataplane214), &core_xds.DataplaneMetadata{IPv6Enabled: false})

		Expect(cfg.IPFamilyMode).To(Equal(tproxy_config.IPFamilyModeIPv4))
		Expect(cfg.EnabledIPv6()).To(BeFalse())
	})

	It("prefers the proxy's metadata over the deprecated fields", func() {
		tproxy := tproxy_dp.DefaultDataplaneConfig()
		tproxy.Redirect.Inbound = tproxy_dp.NewDataplaneTrafficFlow(true, 25006)

		cfg := tproxy_dp.GetDataplaneConfig(parse(dataplane214), &core_xds.DataplaneMetadata{
			TransparentProxy: &tproxy,
			IPv6Enabled:      true,
		})

		Expect(cfg.Redirect.Inbound.Port).To(Equal(tproxy_config.Port(25006)))
	})

	// Zone proxies reach this with a typed nil Dataplane, which is not the same
	// as a nil interface.
	It("leaves transparent proxy off for a proxy with no Dataplane", func() {
		var dp *core_mesh.DataplaneResource

		cfg := tproxy_dp.GetDataplaneConfig(dp, &core_xds.DataplaneMetadata{})

		Expect(cfg.Enabled()).To(BeFalse())
	})

	It("leaves transparent proxy off for a proxy that reports none", func() {
		cfg := tproxy_dp.GetDataplaneConfig(parse(dataplane30), &core_xds.DataplaneMetadata{})

		Expect(cfg.Enabled()).To(BeFalse())
	})

	It("starts with config that still sets the removed options", func() {
		dir := GinkgoT().TempDir()

		cpConf := filepath.Join(dir, "kuma-cp.conf")
		Expect(os.WriteFile(cpConf, []byte(`
runtime:
  kubernetes:
    injector:
      otelPipeEnabled: false
`), 0o600)).To(Succeed())
		cpCfg := kuma_cp.DefaultConfig()
		Expect(core_config.NewLoader(&cpCfg).LoadFile(cpConf)).To(Succeed())

		dpConf := filepath.Join(dir, "kuma-dp.conf")
		Expect(os.WriteFile(dpConf, []byte(`
dataplaneRuntime:
  strictInboundPortsEnabled: false
  reusePortEnabled: false
`), 0o600)).To(Succeed())
		dpCfg := kuma_dp.DefaultConfig()
		Expect(core_config.NewLoader(&dpCfg).LoadFile(dpConf)).To(Succeed())

		for k, v := range map[string]string{
			"KUMA_DATAPLANE_RUNTIME_STRICT_INBOUND_PORTS_ENABLED": "false",
			"KUMA_DATAPLANE_RUNTIME_REUSE_PORT_ENABLED":           "false",
			"KUMA_RUNTIME_KUBERNETES_INJECTOR_OTEL_PIPE_ENABLED":  "false",
		} {
			GinkgoT().Setenv(k, v)
		}
		dpCfg2 := kuma_dp.DefaultConfig()
		Expect(core_config.NewLoader(&dpCfg2).WithEnvVarsLoading("").LoadFile("")).To(Succeed())
	})
})
