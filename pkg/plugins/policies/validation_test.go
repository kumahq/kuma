package policies_test

import (
    "github.com/kumahq/kuma/pkg/plugins/policies"
    meshtrace_proto "github.com/kumahq/kuma/pkg/plugins/policies/meshtrace/api/v1alpha1"
    util_proto "github.com/kumahq/kuma/pkg/util/proto"
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
)

var _ = Describe("policies validation", func() {
    Describe("ValidateSchema()", func() {
        DescribeTable("valid MeshTrace should pass validation",
            func(resourceYaml string) {
                meshTrace := meshtrace_proto.NewMeshTraceResource()
                err := util_proto.FromYAML([]byte(resourceYaml), meshTrace.Spec)
                Expect(err).ToNot(HaveOccurred())

                verr := policies.ValidateSchema(meshTrace.Spec, "MeshTrace")

                Expect(verr).To(BeNil())
            },
            Entry("MeshTrace", `
targetRef:
  kind: MeshService
  name: backend
default:
  backends:
    - zipkin:
        url: http://jaeger-collector.mesh-observability:9411/api/v2/spans
        apiVersion: httpJson
  tags:
    - name: team
      literal: core
    - name: env
      header:
        name: x-env
        default: prod
    - name: version
      header:
        name: x-version
  sampling:
    overall: 80
    random: 60
    client: 40
`),
        )
        
    })
})
