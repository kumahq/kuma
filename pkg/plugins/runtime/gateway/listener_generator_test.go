package gateway_test

import (
	"context"
	"path"
	"sync"

	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/test/matchers"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
	xds_server "github.com/kumahq/kuma/pkg/xds/server/v3"
)

var _ = Describe("Gateway Listener", func() {
	var rt runtime.Runtime

	Do := func(gateway string) (cache.ResourceSnapshot, error) {
		serverCtx := xds_server.NewXdsContext()
		statsCallbacks, err := util_xds.NewStatsCallbacks(rt.Metrics(), "xds", util_xds.NoopVersionExtractor)
		if err != nil {
			return nil, err
		}
		reconciler := xds_server.DefaultReconciler(rt, serverCtx, statsCallbacks, &sync.Mutex{})

		Expect(StoreInlineFixture(rt, []byte(gateway))).To(Succeed())

		// Unmarshal the gateway YAML again so that we can figure
		// out which mesh it's in.
		r, err := rest.YAML.UnmarshalCore([]byte(gateway))
		Expect(err).To(Succeed())

		// We expect there to be a Dataplane fixture named
		// "default" in the current mesh.
		xdsCtx, proxy := MakeGeneratorContext(rt,
			core_model.ResourceKey{Mesh: r.GetMeta().GetMesh(), Name: "default"})

		Expect(proxy.Dataplane.Spec.IsBuiltinGateway()).To(BeTrue())

		if _, err := reconciler.Reconcile(context.Background(), *xdsCtx, proxy); err != nil {
			return nil, err
		}

		return serverCtx.Cache().GetSnapshot(proxy.Id.String())
	}

	BeforeEach(func() {
		var err error

		rt, err = BuildRuntime()
		Expect(err).To(Succeed(), "build runtime instance")

		Expect(StoreNamedFixture(rt, "mesh-default.yaml")).To(Succeed())
		Expect(StoreNamedFixture(rt, "serviceinsight-default.yaml")).To(Succeed())
		Expect(StoreNamedFixture(rt, "dataplane-default.yaml")).To(Succeed())

		Expect(StoreNamedFixture(rt, "mesh-tracing.yaml")).To(Succeed())
		Expect(StoreNamedFixture(rt, "serviceinsight-tracing.yaml")).To(Succeed())
		Expect(StoreNamedFixture(rt, "dataplane-tracing.yaml")).To(Succeed())
		Expect(StoreNamedFixture(rt, "traffictrace.yaml")).To(Succeed())

		Expect(StoreNamedFixture(rt, "mesh-logging.yaml")).To(Succeed())
		Expect(StoreNamedFixture(rt, "serviceinsight-logging.yaml")).To(Succeed())
		Expect(StoreNamedFixture(rt, "dataplane-logging.yaml")).To(Succeed())
		Expect(StoreNamedFixture(rt, "trafficlog.yaml")).To(Succeed())

		// This is an arbitrary certificate and key generated with kumactl.
		Expect(StoreInlineFixture(rt, []byte(`
type: Secret
mesh: default
name: server-certificate
data: LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFb3dJQkFBS0NBUUVBM3ZWM1cvNXJtYzlYL2Nsb1l2cEFnb0hxMG5GWTZWSWRaU0VDeVdMNnprUG5LNHhPCjFaRDN4SXpQTTZPRGgxYjFJRDFIR29rWE5QT21sVnQ1dlg0YXg3aEdPVTVqNE82cFJNSGdGTVNuMXVrV2ExYi8KTnNIN0xZQi9MeW9MTkpCVnd1NThkSzBwczAwbHA5RDZiUm9yRCtJS08va1ZIeURHb09xblVFYks5bXV6Z3I2dApmcVlnYVlGcDBqellURnJCalBMYTRWSG1OeSsxUWRHTnNzbU8weTVrajhadUhidHFKeWUvV0FzTGxOTXpBZXRhClFrYmNEcDR4MkEvZWQvZ0w5blVTYmpTQXAzQTVUZjlNMHJlVnY1UUF1RzVEQWY0SDVCZUdXVUlTWkFRY0RqSVgKb0sxclUvaEpzRTZ4NXorU3Frdm5BL25BYXo4M210OWVCb1BhYXdJREFRQUJBb0lCQUFteS9mMUhoU0RNejFRZwpCZVdBWTN3SjhOQTAxQnhhVVNNTUc1WHRNMkh6dkVPOXQ5UThtVHE0c1c3YXB5Y2xGa2JQdzU4WTVhU05FT3NnCmJweGF0d21ITDY3Z2hTSE00Qm80b09ubVlETE9Nd1o2WTJIYmNIVGJTUzBoRkJtMlNiVFFNU1BXUUtFbk13TW8KNlN3RDNtbXplS1NCUW5UM05RemRDR2hLbkJ1NklPTVk4RkJjRmFJcnNPRkFVRWVTblZybjZFOUVwYjNxYko4SwpSVUkxOG1QZ0d5V3R5ZXMxTkcyZWxtM2g0V3UrNjdjc3d1TU5xRzY0bHVJdHB6NkZMTGZLbEhTNXU1d0RsbTJICktLaFlpTW10em9IRnQ5UnRlQndBb2RGRHA5eDYvaTYvZ3RUaFNLcWc5RE80VWVQdThqK0w0ekVoSjZqTWNhbzAKZnFFcUdMa0NnWUVBNFBsdndGUHY2eitFTjYwSWVqeU9QS0Q5cGR4VjZCWVNpN0NtTE1tOVpIdW5lbHVJME1UaApYa1ZINlZabktYZ3dhQjlZcGVCM25zVW11WlJhSlMwMWRCRFMyWmRtL0RLYUN3azRwS3lUN3FOY3lKQUM1cGRGCjR3Y1RPUytrOFBtNHpvMHhkbUUwQU1LSUpLZUx3ZlU1OWRpemFTUkxJVXNkNFpiSlpLU1dSNGNDZ1lFQS9iVGYKb3NxWTc0MHVaZDhQRGo3ekl3QWpRR2hNb0Q2N0QyVkZOazRpMUQ4UENnS1hRNlRsTnFPOGMvLysvdmdQT0pxRQpKeEIvZGFTYXkwRU93clJvc0duZlFsUm9CV3dFN0ZtbEpibGg5UWdNZGhBbnNWa1gwSzlWMllTMkZPdjM2azVNCkFRV2pIdmhmLzBLNGptaExScEs2ZFB1VG5PS0YzTk9zVFhXZUJ2MENnWUFZQUlTN3NEallrRjQ2MG1zbEgzRE4KWngrb29tbEg2Wkx3OUZmR1QzKzFTTHdGZ2Q2RzUzcGo1R0JYdExBczdIVzlwaHAvR0FPckhMMlU3dzd2Q0hPNwpmbEFBaHZhbDBZQTl6UzRONDV1a3lpa0wvTkZTYUxFOEYzVWxsTCswTmZCUm1SNjkwb0VKMDdkU3NjMW5WQkpxCitFT3I1QU5mK2ZPbUxjQXV6S0I3NFFLQmdHK1RXajdueHJhamFtSlc1UElvOFJqVmVLdGNzMFpPRUVwSENWZEcKcWI2YU5PejhFclluRUw4azV6NUV1VW84b2NVTS8wMkd6ZWRaQ3RLVXUvOFpCR21CUmpTUGxtZThCN1pCL29WRwpzRFBvNUVJUC9NVGNIOE1oT1NvK1dTMStVVHQwVDZ5clkvKzh6OHNjOXJsNldKQ2krdWx6c29sdWZkeU9JdHExCi9WZXBBb0dCQUtWbVdwOU5BL1Eva2VDN2lCUWx5Q2pVSEVGZHBkWXhqTkxvYlcvcjIxb1RtYmNVRHJCa011aysKaG84NXI0a3MzdDFrT0RaMm9pNkFDVkM4QmFIMWloUTI5K1ZxQURxR3JPMlRXd0ttSEpSaG9UcmF6eXVkOVRWSApTSzR1ODArM016Zzlwb1V2ODBsd3R1Sjh4RFdwUWNSSXAxQ2Zxd2t2SUEzOTNKQ3c3VkNLCi0tLS0tRU5EIFJTQSBQUklWQVRFIEtFWS0tLS0tCi0tLS0tQkVHSU4gQ0VSVElGSUNBVEUtLS0tLQpNSUlESmpDQ0FnNmdBd0lCQWdJUkFJK0hxeDlIYUZScTh5aWxYZktrUVJJd0RRWUpLb1pJaHZjTkFRRUxCUUF3CkZqRVVNQklHQTFVRUF4TUxaWGhoYlhCc1pTNWpiMjB3SGhjTk1qRXhNVEF6TURRek1ERTNXaGNOTXpFeE1UQXgKTURRek1ERTNXakFXTVJRd0VnWURWUVFERXd0bGVHRnRjR3hsTG1OdmJUQ0NBU0l3RFFZSktvWklodmNOQVFFQgpCUUFEZ2dFUEFEQ0NBUW9DZ2dFQkFONzFkMXYrYTVuUFYvM0phR0w2UUlLQjZ0SnhXT2xTSFdVaEFzbGkrczVECjV5dU1UdFdROThTTXp6T2pnNGRXOVNBOVJ4cUpGelR6cHBWYmViMStHc2U0UmpsT1krRHVxVVRCNEJURXA5YnAKRm10Vy96YkIreTJBZnk4cUN6U1FWY0x1ZkhTdEtiTk5KYWZRK20wYUt3L2lDanY1RlI4Z3hxRHFwMUJHeXZacgpzNEsrclg2bUlHbUJhZEk4MkV4YXdZenkydUZSNWpjdnRVSFJqYkxKanRNdVpJL0diaDI3YWljbnYxZ0xDNVRUCk13SHJXa0pHM0E2ZU1kZ1AzbmY0Qy9aMUVtNDBnS2R3T1UzL1ROSzNsYitVQUxodVF3SCtCK1FYaGxsQ0VtUUUKSEE0eUY2Q3RhMVA0U2JCT3NlYy9rcXBMNXdQNXdHcy9ONXJmWGdhRDJtc0NBd0VBQWFOdk1HMHdEZ1lEVlIwUApBUUgvQkFRREFnS2tNQk1HQTFVZEpRUU1NQW9HQ0NzR0FRVUZCd01CTUE4R0ExVWRFd0VCL3dRRk1BTUJBZjh3CkhRWURWUjBPQkJZRUZJc2MySU1lQ0lvcm1CL3A1elVkQnZkNXFVR2RNQllHQTFVZEVRUVBNQTJDQzJWNFlXMXcKYkdVdVkyOXRNQTBHQ1NxR1NJYjNEUUVCQ3dVQUE0SUJBUUE2ZThlSkhaUmhER2lORzlvSWtjZGlydmRXNHQ3RwpBcGFXQUluWEo1bGJwMEdPRkNQdEtTc0lCc3F5TnFjWWhHd3o2OVVULzBsNzIrbS9OQ2t0Qlp6Q3ZSMGppRllVCnNzblpYM3E0QllubWUyMEZmN284azFTSDRYUTNpSU1lUUlwT2lFbW9pSHBhQm1EczgxVGpyT3ZoSTJXeE83S3QKblZpVGZWS2V5clFZSnRqK3BkVjJKeFJxemJHYjg5M2wzVXRFblVJYlZrU2pTaHpPUUk5K1BuRE40ZStLUEZDZQpvdmlTQllNVjhUUUhOSGxvNXF2Z2RTRWU2OEJHRUF1TDlkRkc2S0JmZmdCTzh0M1UvVWFINGdpYUlHN042RndKClpNU0diWGNuTE5XSk9XUk1veW1TcGs4YTkvaFhYS1ZZSHBwYmtBZUZHUndobTdYRFhzQ3dxNVdkCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
`))).To(Succeed())
	})

	DescribeTable("Generate Envoy xDS resources",
		func(golden string, gateway string) {
			snap, err := Do(gateway)
			Expect(err).To(Succeed())

			out, err := yaml.Marshal(MakeProtoSnapshot(snap))
			Expect(err).To(Succeed())

			Expect(out).To(matchers.MatchGoldenYAML(path.Join("testdata", golden)))
		},
		Entry("should generate a single listener",
			"01-gateway-listener.yaml", `
type: MeshGateway
mesh: default
name: edge-gateway
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  listeners:
  - port: 8080
    protocol: HTTP
    tags:
      port: http/8080
`),
		Entry("should generate a multiple listeners",
			"02-gateway-listener.yaml", `
type: MeshGateway
mesh: default
name: edge-gateway
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  listeners:
  - port: 8080
    protocol: HTTP
    tags:
      port: http/8080
  - port: 9090
    protocol: HTTP
    tags:
      port: http/9090
`),
		Entry("should generate listener tracing",
			"03-gateway-listener.yaml", `
type: MeshGateway
mesh: tracing
name: tracing-gateway
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  listeners:
  - port: 8080
    protocol: HTTP
    tags:
      port: http/8080
`),

		Entry("should generate listener logging",
			"04-gateway-listener.yaml", `
type: MeshGateway
mesh: logging
name: logging-gateway
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  listeners:
  - port: 8080
    protocol: HTTP
    tags:
      port: http/8080
`),

		Entry("should order HTTPS wildcard hostnames last",
			"05-gateway-listener.yaml", `
type: MeshGateway
mesh: default
name: default-gateway
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  listeners:
  - port: 443
    protocol: HTTPS
    hostname: bar.example.com
    tls:
      mode: TERMINATE
      certificates:
      - secret: server-certificate
    tags:
      name: example.com
  - port: 443
    protocol: HTTPS
    hostname: "*.example.com"
    tls:
      mode: TERMINATE
      certificates:
      - secret: server-certificate
    tags:
      name: example.com
  - port: 443
    protocol: HTTPS
    hostname: foo.example.com
    tls:
      mode: TERMINATE
      certificates:
      - secret: server-certificate
    tags:
      name: foo.example.com
  - port: 443
    protocol: HTTPS
    tls:
      mode: TERMINATE
      certificates:
      - secret: server-certificate
    tags:
      name: any-hostname
`),

		Entry("should generate TCP listener",
			"tcp-listener.yaml", `
type: MeshGateway
mesh: default
name: default-gateway
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  listeners:
  - port: 443
    protocol: TCP
    hostname: bar.example.com
    tags:
      name: example.com
`),

		Entry("should add connection limits",
			"connection-limited-listener.yaml", `
type: MeshGateway
mesh: default
name: default-gateway
selectors:
- match:
    kuma.io/service: gateway-default
conf:
  listeners:
  - port: 443
    protocol: TCP
    hostname: bar.example.com
    resources:
      connectionLimit: 10000
`),
	)
})
