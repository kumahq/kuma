package mesh_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	. "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

var _ = Describe("VirtualOutbound_Helpers", func() {
	type portTestCase struct {
		in        *mesh_proto.VirtualOutbound_Conf
		givenTags map[string]string
		thenPort  uint32
		thenErr   string
	}
	DescribeTable("EvalPort",
		func(tc portTestCase) {
			virtualOutbound := NewVirtualOutboundResource()
			virtualOutbound.Spec.Conf = tc.in

			// when
			out, err := virtualOutbound.EvalPort(tc.givenTags)
			// then
			if tc.thenErr == "" {
				Expect(err).ToNot(HaveOccurred())
				Expect(out).To(Equal(tc.thenPort))
			} else {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(tc.thenErr))
				Expect(out).To(Equal(uint32(0)))
			}
		},
		Entry("not int out", portTestCase{
			in: &mesh_proto.VirtualOutbound_Conf{
				Port: "{{.port}}",
				Parameters: []*mesh_proto.VirtualOutbound_Conf_TemplateParameter{
					{Name: "port"},
				},
			},
			givenTags: map[string]string{"port": "hello"},
			thenErr:   "evaluation of template with parameters didn't evaluate to a parsable number result='hello'",
		}),
		Entry("complex template", portTestCase{
			in: &mesh_proto.VirtualOutbound_Conf{
				Port: "{{.port}}{{.offset}}",
				Parameters: []*mesh_proto.VirtualOutbound_Conf_TemplateParameter{
					{Name: "port"},
					{Name: "offset"},
				},
			},
			givenTags: map[string]string{"port": "80", "offset": "81"},
			thenPort:  8081,
		}),
		Entry("out of bound", portTestCase{
			in: &mesh_proto.VirtualOutbound_Conf{
				Port: "{{.port}}",
				Parameters: []*mesh_proto.VirtualOutbound_Conf_TemplateParameter{
					{Name: "port"},
				},
			},
			givenTags: map[string]string{"port": "80000", "offset": "81"},
			thenErr:   "a port outside of the range [1..65535] result='80000'",
		}),
	)

	type hostTestCase struct {
		in        *mesh_proto.VirtualOutbound_Conf
		givenTags map[string]string
		thenHost  string
		thenErr   string
	}
	DescribeTable("EvalHost()",
		func(tc hostTestCase) {
			virtualOutbound := NewVirtualOutboundResource()
			virtualOutbound.Spec.Conf = tc.in

			// when
			out, err := virtualOutbound.EvalHost(tc.givenTags)
			// then
			if tc.thenErr == "" {
				Expect(err).ToNot(HaveOccurred())
				Expect(out).To(Equal(tc.thenHost))
			} else {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(tc.thenErr))
				Expect(out).To(BeEmpty())
			}
		},
		Entry("invalid host out", hostTestCase{
			in: &mesh_proto.VirtualOutbound_Conf{
				Host: "{{.host}}",
				Parameters: []*mesh_proto.VirtualOutbound_Conf_TemplateParameter{
					{Name: "host", TagKey: "kuma.io/service"},
				},
			},
			givenTags: map[string]string{"kuma.io/service": "foo()bar"},
			thenErr:   "evaluation of template with parameters didn't return a valid dns name result='foo()bar'",
		}),
		Entry("simple eval", hostTestCase{
			in: &mesh_proto.VirtualOutbound_Conf{
				Host: "{{.host}}.{{.instance}}",
				Parameters: []*mesh_proto.VirtualOutbound_Conf_TemplateParameter{
					{Name: "host", TagKey: "kuma.io/service"},
					{Name: "instance"},
				},
			},
			givenTags: map[string]string{"kuma.io/service": "foo-bar", "instance": "2"},
			thenHost:  "foo-bar.2",
		}),
	)
})
