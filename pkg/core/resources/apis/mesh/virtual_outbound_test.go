package mesh_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"

	. "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

var _ = Describe("VirtualOutbound", func() {

	Describe("Templates()", func() {
		type testCase struct {
			conf         *mesh_proto.VirtualOutbound_Conf
			expectedHost string
			expectedPort uint32
			expectedErr  string
			tags         map[string]string
		}
		DescribeTable("Host",
			func(tt testCase) {
				// setup
				voutbound := NewVirtualOutboundResource()
				voutbound.Spec = &mesh_proto.VirtualOutbound{
					Conf: tt.conf,
				}

				// when
				o, err := voutbound.EvalHost(tt.tags)
				// then
				if tt.expectedErr != "" {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(tt.expectedErr))
				} else {
					Expect(err).ToNot(HaveOccurred())
					Expect(o).To(Equal(tt.expectedHost))
				}
			},
			Entry("not a template", testCase{
				conf:         &mesh_proto.VirtualOutbound_Conf{Host: "foo.mesh"},
				expectedHost: "foo.mesh",
			}),
			Entry("with key", testCase{
				conf:         &mesh_proto.VirtualOutbound_Conf{Host: "{{.foo}}.mesh", Parameters: map[string]string{"foo": "my-tag"}},
				tags:         map[string]string{"my-tag": "bar"},
				expectedHost: "bar.mesh",
			}),
			Entry("with multiple keys", testCase{
				conf:         &mesh_proto.VirtualOutbound_Conf{Host: "{{.foo}}.{{.bar}}", Parameters: map[string]string{"foo": "my-tag", "bar": "my-tag"}},
				tags:         map[string]string{"my-tag": "bar"},
				expectedHost: "bar.bar",
			}),
			Entry("missing key", testCase{
				conf:        &mesh_proto.VirtualOutbound_Conf{Host: "{{.foo}}.mesh", Parameters: map[string]string{"foo": "my-tag"}},
				expectedErr: "evaluation of template with parameters didn't return a valid dns name result='<no value>.mesh'",
			}),
		)
		DescribeTable("Port",
			func(tt testCase) {
				// setup
				voutbound := NewVirtualOutboundResource()
				voutbound.Spec = &mesh_proto.VirtualOutbound{
					Conf: tt.conf,
				}

				// when
				o, err := voutbound.EvalPort(tt.tags)
				// then
				if tt.expectedErr != "" {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(tt.expectedErr))
				} else {
					Expect(err).ToNot(HaveOccurred())
					Expect(o).To(Equal(tt.expectedPort))
				}
			},
			Entry("not a template", testCase{
				conf:         &mesh_proto.VirtualOutbound_Conf{Port: "80"},
				expectedPort: 80,
			}),
			Entry("simple", testCase{
				conf:         &mesh_proto.VirtualOutbound_Conf{Port: "{{.port}}", Parameters: map[string]string{"port": "port"}},
				tags:         map[string]string{"port": "80"},
				expectedPort: 80,
			}),
			Entry("don't output int", testCase{
				conf:        &mesh_proto.VirtualOutbound_Conf{Port: "{{.foo}}", Parameters: map[string]string{"foo": "my-tag"}},
				tags:        map[string]string{"my-tag": "bar"},
				expectedErr: "evaluation of template with parameters didn't evaluate to a parsable number result='bar'",
			}),
			Entry("output too large", testCase{
				conf:        &mesh_proto.VirtualOutbound_Conf{Port: "{{.foo}}", Parameters: map[string]string{"foo": "my-tag"}},
				tags:        map[string]string{"my-tag": "999999"},
				expectedErr: "evaluation of template returned a port outside of the range [1..65535] result='999999'",
			}),
			Entry("missing key", testCase{
				conf:        &mesh_proto.VirtualOutbound_Conf{Port: "{{.foo}}", Parameters: map[string]string{"foo": "my-tag"}},
				expectedErr: "evaluation of template with parameters didn't evaluate to a parsable number result='<no value>'",
			}),
		)
	})
})
