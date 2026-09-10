package v1alpha1_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	config_proto "github.com/kumahq/kuma/v3/pkg/config/app/kumactl/v1alpha1"
)

// jsonpb wrote the camel cased spelling, accepted the underscored one as well and dropped
// every zero value. The file lives in the user's home directory rather than a database, so
// a config an older kumactl wrote has to keep loading with no migration step, and a config
// this kumactl writes has to keep loading in an older one.
var _ = Describe("kumactl config file compatibility", func() {
	It("round trips the bytes an older kumactl wrote", func() {
		stored := `{"controlPlanes":[{"name":"test","coordinates":{"apiServer":{"url":"https://localhost:5681","caCertFile":"/ca.pem","clientCertFile":"/crt.pem","clientKeyFile":"/key.pem","headers":[{"key":"x-token","value":"secret"}],"authType":"tokens","authConf":{"token":"abc"},"skipVerify":true}}}],"contexts":[{"name":"test","controlPlane":"test","defaults":{"mesh":"demo"}}],"currentContext":"test"}`

		cfg := &config_proto.Configuration{}
		Expect(json.Unmarshal([]byte(stored), cfg)).To(Succeed())

		rewritten, err := json.Marshal(cfg)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(rewritten)).To(Equal(stored))
	})

	It("reads the underscored spelling jsonpb also accepted", func() {
		stored := `
control_planes:
- name: test1
  coordinates:
    api_server:
      url: https://test1.internal:5681
      ca_cert_file: /ca.pem
      client_cert_file: /crt.pem
      client_key_file: /key.pem
      auth_type: tokens
      auth_conf:
        token: abc
      skip_verify: true
contexts:
- name: test1
  control_plane: test1
  defaults:
    mesh: demo
current_context: test1
`
		cfg := &config_proto.Configuration{}
		Expect(yaml.Unmarshal([]byte(stored), cfg)).To(Succeed())

		Expect(cfg.CurrentContext).To(Equal("test1"))
		Expect(cfg.ControlPlanes).To(HaveLen(1))
		Expect(cfg.Contexts).To(HaveLen(1))
		Expect(cfg.Contexts[0].ControlPlane).To(Equal("test1"))
		Expect(cfg.Contexts[0].Defaults.Mesh).To(Equal("demo"))

		api := cfg.ControlPlanes[0].Coordinates.ApiServer
		Expect(api.Url).To(Equal("https://test1.internal:5681"))
		Expect(api.CaCertFile).To(Equal("/ca.pem"))
		Expect(api.ClientCertFile).To(Equal("/crt.pem"))
		Expect(api.ClientKeyFile).To(Equal("/key.pem"))
		Expect(api.AuthType).To(Equal("tokens"))
		Expect(api.AuthConf).To(Equal(map[string]string{"token": "abc"}))
		Expect(api.SkipVerify).To(BeTrue())
	})

	It("keeps user supplied authConf keys verbatim when they spell a field of the enclosing message", func() {
		stored := `{"controlPlanes":[{"name":"test","coordinates":{"apiServer":{"authConf":{"auth_type":"a","client_key_file":"c","skip_verify":"b"}}}}]}`

		cfg := &config_proto.Configuration{}
		Expect(json.Unmarshal([]byte(stored), cfg)).To(Succeed())

		api := cfg.ControlPlanes[0].Coordinates.ApiServer
		Expect(api.AuthConf).To(Equal(map[string]string{"auth_type": "a", "skip_verify": "b", "client_key_file": "c"}))
		Expect(api.AuthType).To(BeEmpty())
		Expect(api.SkipVerify).To(BeFalse())
		Expect(api.ClientKeyFile).To(BeEmpty())

		rewritten, err := json.Marshal(cfg)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(rewritten)).To(Equal(stored))
	})

	It("drops the values jsonpb left out", func() {
		out, err := json.Marshal(&config_proto.Configuration{})
		Expect(err).ToNot(HaveOccurred())
		Expect(string(out)).To(Equal("{}"))
	})

	DescribeTable("reports what is wrong instead of panicking",
		func(stored string, expected string) {
			cfg := &config_proto.Configuration{}
			err := yaml.Unmarshal([]byte(stored), cfg)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(expected))
		},
		Entry("a scalar where the document belongs", "42\n", "configuration must be an object"),
		Entry("a scalar where the coordinates belong",
			"controlPlanes:\n- name: test\n  coordinates: 42\n", "coordinates must be an object"),
		Entry("a scalar where the api server belongs",
			"controlPlanes:\n- name: test\n  coordinates:\n    apiServer: 42\n", "apiServer must be an object"),
		Entry("a scalar where a context belongs", "contexts:\n- 42\n", "context must be an object"),
		Entry("a string where the control plane list belongs",
			"controlPlanes: not-a-list\n", "controlPlanes"),
	)

	It("keeps the getters usable on a nil receiver", func() {
		var cfg *config_proto.Configuration
		Expect(cfg.GetCurrentContext()).To(BeEmpty())
		Expect(cfg.GetControlPlanes()).To(BeNil())

		var cp *config_proto.ControlPlane
		Expect(cp.GetCoordinates().GetApiServer().GetUrl()).To(BeEmpty())

		var ctx *config_proto.Context
		Expect(ctx.GetDefaults().GetMesh()).To(BeEmpty())
	})
})
