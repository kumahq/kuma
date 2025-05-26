package externalservices

import (
	"encoding/base64"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"google.golang.org/protobuf/types/known/wrapperspb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/universal"
)

func Policy() {
	meshName := "external-service-base"
	meshNameNoDefaults := "external-service-no-default-policy"
	meshDefaulMtlsOn := func(meshName string) InstallFunc {
		return YamlUniversal(fmt.Sprintf(`
type: Mesh
name: "%s"
mtls:
  enabledBackend: ca-1
  backends:
  - name: ca-1
    type: builtin
networking:
  outbound:
    passthrough: false
`, meshName))
	}

	externalService := func(service, address, meshName string, tls bool, caCert []byte) *core_mesh.ExternalServiceResource {
		res := &core_mesh.ExternalServiceResource{
			Meta: &test_model.ResourceMeta{
				Mesh: meshName,
				Name: fmt.Sprintf("%s-%s", service, meshName),
			},
			Spec: &mesh_proto.ExternalService{
				Tags: map[string]string{
					mesh_proto.ServiceTag:  service,
					mesh_proto.ProtocolTag: core_mesh.ProtocolHTTP,
				},
				Networking: &mesh_proto.ExternalService_Networking{
					Address: address,
					Tls: &mesh_proto.ExternalService_Networking_TLS{
						Enabled: tls,
					},
				},
			},
		}
		if tls {
			res.Spec.Networking.Tls.CaCert = &system_proto.DataSource{
				Type: &system_proto.DataSource_Inline{
					Inline: &wrapperspb.BytesValue{
						Value: caCert,
					},
				},
			}
		}
		return res
	}

	esHttpName := "es-http"
	esHttpsName := "es-https"
	esHttp2Name := "es-http-2"

	var esHttpHostPort string
	var esHttp2HostPort string
	var esHttpsHostPort string

	var esHttpContainerName string
	var esHttpsContainerName string
	var esHttp2ContainerName string

	BeforeAll(func() {
		esHttpContainerName = fmt.Sprintf("%s_%s_%s", universal.Cluster.Name(), meshName, esHttpName)
		esHttpsContainerName = fmt.Sprintf("%s_%s_%s", universal.Cluster.Name(), meshName, esHttpsName)
		esHttp2ContainerName = fmt.Sprintf("%s_%s_%s", universal.Cluster.Name(), meshName, esHttp2Name)

		esHttpHostPort = fmt.Sprintf("%s:%d", esHttpContainerName, 80)
		esHttpsHostPort = fmt.Sprintf("%s:%d", esHttpsContainerName, 443)
		esHttp2HostPort = fmt.Sprintf("%s:%d", esHttp2ContainerName, 81)

		err := NewClusterSetup().
			Install(meshDefaulMtlsOn(meshName)).
			Install(TrafficPermissionUniversal(meshName)).
			Install(TrafficRouteUniversal(meshName)).
			Install(meshDefaulMtlsOn(meshNameNoDefaults)).
			Install(TestServerExternalServiceUniversal(esHttpName, 80, false, WithDockerContainerName(esHttpContainerName))).
			Install(TestServerExternalServiceUniversal(esHttpsName, 443, true, WithDockerContainerName(esHttpsContainerName))).
			Install(TestServerExternalServiceUniversal(esHttp2Name, 81, false, WithDockerContainerName(esHttp2ContainerName))).
			Install(DemoClientUniversal("demo-client", meshName, WithTransparentProxy(true))).
			Install(DemoClientUniversal("demo-client-no-defaults", meshNameNoDefaults, WithTransparentProxy(true))).
			Setup(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugUniversal(universal.Cluster, meshName)
	})

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(universal.Cluster.DeleteApp(esHttpName)).To(Succeed())
		Expect(universal.Cluster.DeleteApp(esHttpsName)).To(Succeed())
		Expect(universal.Cluster.DeleteApp(esHttp2Name)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	checkSuccessfulRequest := func(url, clientName string, matcher types.GomegaMatcher) {
		Eventually(func(g Gomega) {
			stdout, stderr, err := client.CollectResponse(
				universal.Cluster, clientName, url,
				client.WithVerbose(),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))
			g.Expect(stdout).To(matcher)
		}, "30s", "500ms").WithOffset(1).Should(Succeed())
	}

	contextFor := func(name, meshName, clientName string) {
		Context(name, func() {
			It("should route to external-service", func() {
				err := universal.Cluster.Install(ResourceUniversal(externalService("ext-srv-1", esHttpHostPort, meshName, false, nil)))
				Expect(err).ToNot(HaveOccurred())

				checkSuccessfulRequest("ext-srv-1.mesh", clientName, And(
					Not(ContainSubstring("HTTPS")),
					// Should rewrite host header
					ContainSubstring(fmt.Sprintf(`"Host":["%s"]`, esHttpContainerName)),
				))
				checkSuccessfulRequest(esHttpHostPort, clientName, Not(ContainSubstring("HTTPS")))
			})

			It("should route to external-service with same hostname but different ports", func() {
				err := universal.Cluster.Install(ResourceUniversal(externalService("ext-srv-1", esHttpHostPort, meshName, false, nil)))
				Expect(err).ToNot(HaveOccurred())

				err = universal.Cluster.Install(ResourceUniversal(externalService("ext-srv-2", esHttp2HostPort, meshName, false, nil)))
				Expect(err).ToNot(HaveOccurred())

				// when access the first external service with .mesh
				checkSuccessfulRequest("ext-srv-1.mesh", clientName,
					And(Not(ContainSubstring("HTTPS")), ContainSubstring("es-http")))
				checkSuccessfulRequest(esHttpHostPort, clientName,
					And(Not(ContainSubstring("HTTPS")), ContainSubstring("es-http")))

				checkSuccessfulRequest("ext-srv-2.mesh", clientName,
					And(Not(ContainSubstring("HTTPS")), ContainSubstring("es-http-2")))
				checkSuccessfulRequest(esHttp2HostPort, clientName,
					And(Not(ContainSubstring("HTTPS")), ContainSubstring("es-http-2")))
			})

			It("should route to external-service over tls", func() {
				// when set invalid certificate
				otherCert := "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURMRENDQWhTZ0F3SUJBZ0lRSGRQaHhPZlhnV3VOeG9GbFYvRXdxVEFOQmdrcWhraUc5dzBCQVFzRkFEQVAKTVEwd0N3WURWUVFERXdScmRXMWhNQjRYRFRJd01Ea3hOakV5TWpnME5Gb1hEVE13TURreE5ERXlNamcwTkZvdwpEekVOTUFzR0ExVUVBeE1FYTNWdFlUQ0NBU0l3RFFZSktvWklodmNOQVFFQkJRQURnZ0VQQURDQ0FRb0NnZ0VCCkFPWkdiV2hTbFFTUnhGTnQ1cC8yV0NLRnlIWjNDdXdOZ3lMRVA3blM0Wlh5a3hzRmJZU3VWM2JJZ0Y3YlQvdXEKYTVRaXJlK0M2MGd1aEZicExjUGgyWjZVZmdJZDY5R2xRekhNVlljbUxHalZRdXlBdDRGTU1rVGZWRWw1STRPYQorMml0M0J2aWhWa0toVXo4eTVSUjVLYnFKZkdwNFoyMEZoNmZ0dG9DRmJlT0RtdkJzWUpGbVVRUytpZm95TVkvClAzUjAzU3U3ZzVpSXZuejd0bWt5ZG9OQzhuR1JEemRENUM4Zkp2clZJMVVYNkpSR3lMS3Q0NW9RWHQxbXhLMTAKNUthTjJ6TlYyV3RIc2FKcDlid3JQSCtKaVpHZVp5dnVoNVV3ckxkSENtcUs3c205VG9kR3p0VVpZMFZ6QWM0cQprWVZpWFk4Z1VqZk5tK2NRclBPMWtOOENBd0VBQWFPQmd6Q0JnREFPQmdOVkhROEJBZjhFQkFNQ0FxUXdIUVlEClZSMGxCQll3RkFZSUt3WUJCUVVIQXdFR0NDc0dBUVVGQndNQk1BOEdBMVVkRXdFQi93UUZNQU1CQWY4d0hRWUQKVlIwT0JCWUVGR01EQlBQaUJGSjNtdjJvQTlDVHFqZW1GVFYyTUI4R0ExVWRFUVFZTUJhQ0NXeHZZMkZzYUc5egpkSUlKYkc5allXeG9iM04wTUEwR0NTcUdTSWIzRFFFQkN3VUFBNElCQVFDLzE3UXdlT3BHZGIxTUVCSjhYUEc3CjNzSy91dG9XTFgxdGpmOFN1MURnYTZDRFQvZVRXSFpyV1JmODFLT1ZZMDdkbGU1U1JJREsxUWhmYkdHdEZQK1QKdlprcm9vdXNJOVVTMmFDV2xrZUNaV0dUbnF2TG1Eb091anFhZ0RvS1JSdWs0bVFkdE5Ob254aUwvd1p0VEZLaQorMWlOalVWYkxXaURYZEJMeG9SSVZkTE96cWIvTU54d0VsVXlhVERBa29wUXlPV2FURGtZUHJHbWFXamNzZlBHCmFPS293MHplK3pIVkZxVEhiam5DcUVWM2huc1V5UlV3c0JsbjkrakRKWGd3Wk0vdE1sVkpyWkNoMFNsZTlZNVoKTU9CMGZDZjZzVE1OUlRHZzVMcGw2dUlZTS81SU5wbUhWTW8zbjdNQlNucEVEQVVTMmJmL3VvNWdJaXE2WENkcAotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg=="
				caCert, _ := base64.StdEncoding.DecodeString(otherCert)
				err := universal.Cluster.Install(ResourceUniversal(externalService("ext-srv-tls", esHttpsHostPort, meshName, true, caCert)))
				Expect(err).ToNot(HaveOccurred())

				// then accessing the secured external service fails
				Eventually(func(g Gomega) {
					response, err := client.CollectFailure(universal.Cluster, clientName, "http://"+esHttpsHostPort)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(response.ResponseCode).To(Equal(503))
				}, "30s", "1s").MustPassRepeatedly(5).Should(Succeed())

				// when set proper certificate
				cert, _, err := universal.Cluster.Exec("", "", "es-https", "cat /certs/cert.pem")
				Expect(err).ToNot(HaveOccurred())
				Expect(cert).ToNot(BeEmpty())

				err = universal.Cluster.Install(ResourceUniversal(externalService("ext-srv-tls", esHttpsHostPort, meshName, true, []byte(cert))))
				Expect(err).ToNot(HaveOccurred())

				// then accessing the secured external service succeeds
				checkSuccessfulRequest("http://"+esHttpsHostPort, clientName, Not(ContainSubstring("HTTPS")))
			})
		})
	}

	contextFor("without default policies", meshNameNoDefaults, "demo-client-no-defaults")
	contextFor("with default policies", meshName, "demo-client")

	// certauth.idrix.fr is a site for testing mTLS authentication
	// This site requires renegotiation because the server asks for the client certs as a second step
	// We want to run this only on demand because we've got bad experience tying up E2E to external service available on the internet
	// It's hard to rebuild this as a local service in the cluster because many servers dropped support for renegotiation.
	PIt("should check allow negotiation", func() {
		// given
		es := fmt.Sprintf(`
type: ExternalService
mesh: "%s"
name: testmtls
tags:
  kuma.io/service: testmtls
  kuma.io/protocol: http
networking:
  address: certauth.idrix.fr:443
  tls:
    enabled: true
    allowRenegotiation: true
    caCert:
      inline: "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUZCakNDQXU2Z0F3SUJBZ0lSQUlwOVBoUFdMekR2STRhOUtRZHJOUGd3RFFZSktvWklodmNOQVFFTEJRQXcKVHpFTE1Ba0dBMVVFQmhNQ1ZWTXhLVEFuQmdOVkJBb1RJRWx1ZEdWeWJtVjBJRk5sWTNWeWFYUjVJRkpsYzJWaApjbU5vSUVkeWIzVndNUlV3RXdZRFZRUURFd3hKVTFKSElGSnZiM1FnV0RFd0hoY05NalF3TXpFek1EQXdNREF3CldoY05NamN3TXpFeU1qTTFPVFU1V2pBek1Rc3dDUVlEVlFRR0V3SlZVekVXTUJRR0ExVUVDaE1OVEdWMEozTWcKUlc1amNubHdkREVNTUFvR0ExVUVBeE1EVWpFeE1JSUJJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQgpDZ0tDQVFFQXVvZThYQnNBT2N2S0NzM1VaeEQ1QVR5bFRxVmh5eWJLVXZzVkFiZTVLUFVvSHUwbnN5UVlPV2NKCkRBanM0RHF3TzNjT3ZmUGxPVlJCREU2dVFkYVpkTjVSMis5Ny8xaTlxTGNUOXQ0eDFmSnl5WEpxQzROMGxaeEcKQUdRVW1mT3gyU0xaemFpU3Fod21lai8rNzFnRmV3aVZnZHR4RDQ3NzR6RUp1d20rVUUxZmo1RjJQVnFkbm9QeQo2Y1JtcytFR1prTklHSUJsb0RjWW1wdUVNcGV4c3IzRStCVUFuU2VJKytKakY1WnNteWRuUzhUYktGNXB3bm53ClNWemdKRkRoeEx5aEJheDdRRzBBdE1KQlA2ZFl1Qy9GWEp1bHV3bWU4Zjdyc0lVNS9hZ0s3MFhFZU90bEtzTFAKWHp6ZTQxeE5HL2NMSnl1cUMwSjNVMDk1YWgySDJRSURBUUFCbzRINE1JSDFNQTRHQTFVZER3RUIvd1FFQXdJQgpoakFkQmdOVkhTVUVGakFVQmdnckJnRUZCUWNEQWdZSUt3WUJCUVVIQXdFd0VnWURWUjBUQVFIL0JBZ3dCZ0VCCi93SUJBREFkQmdOVkhRNEVGZ1FVeGM5R3BPcjB3OEI2YkpYRUxiQmVraThtNDdrd0h3WURWUjBqQkJnd0ZvQVUKZWJSWjVudTI1ZVFCYzRBSWlNZ2FXUGJwbTI0d01nWUlLd1lCQlFVSEFRRUVKakFrTUNJR0NDc0dBUVVGQnpBQwpoaFpvZEhSd09pOHZlREV1YVM1c1pXNWpjaTV2Y21jdk1CTUdBMVVkSUFRTU1Bb3dDQVlHWjRFTUFRSUJNQ2NHCkExVWRId1FnTUI0d0hLQWFvQmlHRm1oMGRIQTZMeTk0TVM1akxteGxibU55TG05eVp5OHdEUVlKS29aSWh2Y04KQVFFTEJRQURnZ0lCQUU3aWlWMEtBeHlRT05EMUgvbHhYUGpEajdJM2lIcHZzQ1VmN2I2MzJJWUdqdWtKaE0xeQp2NEh6L01yUFUwanR2ZlpwUXRTbEVUNDF5Qk95a2gwRlgrb3UxTmo0U2NPdDlabVduTzhtMk9HMEpBdElJRTM4CjAxUzBxY1loeU9FMkcvOTNaQ2tYdWZCTDcxM3F6WG5RdjVDL3ZpT3lrTnBLcVVneGRLbEVDK0hpOWkyRGNhUjEKZTlLVXdRVVpSaHk1ai9QRWRFZ2xLZzNsOWR0RDR0dVRtN2tadEI4djMyb09qekhUWXcrN0tkemRaaXcvc0J0bgpVZmhCUE9STnVheTRwSnhtWS9XcmhTTWR6Rk8ycTNHdTNNVUJjZG8yN2dvWUtqTDlDVEY4ai9aejU1eWN0VW9WCmFuZUNXcy9halVYK0h5cGtCVEErYzhMR0RMbldPMk5LcTBZRC9wbkFSa0FuWUdQZlVEb0hSOWdWU3AvcVJ4K1oKV2doaURMWnNNd2hOMXpqdFNDMHVCV2l1Z0YzdlROellJRUZmYVBHN1dzM2pEckFNTVllYlE5NUpRK0hJQkQvUgpQQnVIUlRCcHFLbHlEbmtTSERIWVBpTlgzYWRQb1BBY2dkRjNIMi9XMHJtb3N3TVdnVGxMbjFXdTBtcmtzNy9xCnBkV2ZTNlBKMWp0eTgwcjJWS3NNL0RqM1lJRGZialhLZGFGVTVDKzhiaGZKR3FVM3RhS2F1dXowd0hWR1QzZW8KNkZsV2tXWXRidDRwZ2RhbWx3VmVaRVcrTE03cVpFSkVzTU5QcmZDMDNBUEttWnNKZ3BXQ0RXT0tadmtaY3ZqVgp1WWtRNG9tWUNUWDVvaHkra25NamRPbWRIOWM3U3BxRVdCREM4NmZpTmV4K08wWE9NRVpTYThEQQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCi0tLS0tQkVHSU4gQ0VSVElGSUNBVEUtLS0tLQpNSUlGYXpDQ0ExT2dBd0lCQWdJUkFJSVF6N0RTUU9OWlJHUGd1Mk9DaXdBd0RRWUpLb1pJaHZjTkFRRUxCUUF3ClR6RUxNQWtHQTFVRUJoTUNWVk14S1RBbkJnTlZCQW9USUVsdWRHVnlibVYwSUZObFkzVnlhWFI1SUZKbGMyVmgKY21Ob0lFZHliM1Z3TVJVd0V3WURWUVFERXd4SlUxSkhJRkp2YjNRZ1dERXdIaGNOTVRVd05qQTBNVEV3TkRNNApXaGNOTXpVd05qQTBNVEV3TkRNNFdqQlBNUXN3Q1FZRFZRUUdFd0pWVXpFcE1DY0dBMVVFQ2hNZ1NXNTBaWEp1ClpYUWdVMlZqZFhKcGRIa2dVbVZ6WldGeVkyZ2dSM0p2ZFhBeEZUQVRCZ05WQkFNVERFbFRVa2NnVW05dmRDQlkKTVRDQ0FpSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnSVBBRENDQWdvQ2dnSUJBSzNvSkhQMEZEZnptNTRyVnlnYwpoNzdjdDk4NGtJeHVQT1pYb0hqM2RjS2kvdlZxYnZZQVR5amIzbWlHYkVTVHRyRmovUlFTYTc4ZjB1b3hteUYrCjBUTTh1a2oxM1huZnM3ai9FdkVobWt2QmlvWnhhVXBtWm15UGZqeHd2NjBwSWdiejVNRG1nSzdpUzQrM21YNlUKQTUvVFI1ZDhtVWdqVStnNHJrOEtiNE11MFVsWGpJQjB0dG92MERpTmV3TndJUnQxOGpBOCtvK3UzZHBqcStzVwpUOEtPRVV0K3p3dm8vN1YzTHZTeWUwcmdUQklsREhDTkF5bWc0Vk1rN0JQWjdobS9FTE5LakQrSm8yRlIzcXlICkI1VDBZM0hzTHVKdlc1aUI0WWxjTkhsc2R1ODdrR0o1NXR1a21pOG14ZEFRNFE3ZTJSQ09GdnUzOTZqM3grVUMKQjVpUE5naVY1K0kzbGcwMmRaNzdEbkt4SFp1OEEvbEpCZGlCM1FXMEt0WkI2YXdCZHBVS0Q5amYxYjBTSHpVdgpLQmRzMHBqQnFBbGtkMjVITjdyT3JGbGVhSjEvY3RhSnhRWkJLVDVaUHQwbTlTVEpFYWRhbzB4QUgwYWhtYlduCk9sRnVoanVlZlhLbkVnVjRXZTArVVhnVkN3T1BqZEF2QmJJK2Uwb2NTM01GRXZ6RzZ1QlFFM3hEazNTenluVG4Kamg4QkNOQXcxRnR4TnJRSHVzRXdNRnhJdDRJN21LWjlZSXFpb3ltQ3pMcTlnd1Fib29NRFFhSFdCZkVid3JidwpxSHlHTzBhb1NDcUkzSGFhZHI4ZmFxVTlHWS9yT1BOazNzZ3JEUW9vLy9mYjRoVkMxQ0xRSjEzaGVmNFk1M0NJCnJVN20yWXM2eHQwblVXNy92R1QxTTBOUEFnTUJBQUdqUWpCQU1BNEdBMVVkRHdFQi93UUVBd0lCQmpBUEJnTlYKSFJNQkFmOEVCVEFEQVFIL01CMEdBMVVkRGdRV0JCUjV0Rm5tZTdibDVBRnpnQWlJeUJwWTl1bWJiakFOQmdrcQpoa2lHOXcwQkFRc0ZBQU9DQWdFQVZSOVlxYnl5cUZEUURMSFlHbWtnSnlrSXJHRjFYSXB1K0lMbGFTL1Y5bFpMCnViaHpFRm5USVpkKzUweHgrN0xTWUswNXFBdnFGeUZXaGZGUURsbnJ6dUJaNmJySkZlK0duWStFZ1BiazZaR1EKM0JlYllodEY4R2FWMG54dnd1bzc3eC9QeTlhdUovR3BzTWl1L1gxK212b2lCT3YvMlgvcWtTc2lzUmNPai9LSwpORnRZMlB3QnlWUzV1Q2JNaW9nemlVd3RoRHlDMys2V1Z3VzZMTHYzeExmSFRqdUN2akhJSW5Oemt0SENnS1E1Ck9SQXpJNEpNUEorR3NsV1lIYjRwaG93aW01N2lhenRYT29Kd1Rkd0p4NG5MQ2dkTmJPaGRqc252enF2SHU3VXIKVGtYV1N0QW16T1Z5eWdocXBaWGpGYUgzcE8zSkxGK2wrLytzS0FJdXZ0ZDd1K054ZTVBVzB3ZGVSbE44TndkQwpqTlBFbHB6Vm1iVXE0SlVhZ0VpdVREa0h6c3hIcEZLVks3cTQrNjNTTTFOOTVSMU5iZFdoc2NkQ2IrWkFKelZjCm95aTNCNDNualRPUTV5T2YrMUNjZVd4RzFiUVZzNVp1ZnBzTWxqcTRVaTAvMWx2aCt3akNoUDRrcUtPSjJxeHEKNFJncXNhaERZVnZUSDl3N2pYYnlMZWlOZGQ4WE0ydzlVL3Q3eTBGZi85eWkwR0U0NFphNHJGMkxOOWQxMVRQQQptUkd1blVIQmNuV0V2Z0pCUWw5bkpFaVUwWnNudmdjL3ViaFBnWFJSNFhxMzdaMGo0cjdnMVNnRUV6d3hBNTdkCmVteVB4Z2NZeG4vZVI0NC9LSjRFQnMrbFZEUjN2ZXlKbStrWFE5OWIyMS8ramg1WG9zMUFuWDVpSXRyZUdDYz0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo="
    clientCert: # we can pass any client tls pair to this server so here are certs generated by kumactl generate tls-certificate --type=client --key-file=client.key --cert-file=client.pem
      inline: "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURJRENDQWdpZ0F3SUJBZ0lSQU4xeHJ2aGhyMExuSlN0czRrdDAyNkF3RFFZSktvWklodmNOQVFFTEJRQXcKRHpFTk1Bc0dBMVVFQXhNRWEzVnRZVEFlRncweU1UQTJNRGt4TURVek16QmFGdzB6TVRBMk1EY3hNRFV6TXpCYQpNQTh4RFRBTEJnTlZCQU1UQkd0MWJXRXdnZ0VpTUEwR0NTcUdTSWIzRFFFQkFRVUFBNElCRHdBd2dnRUtBb0lCCkFRQ3hFMnV0TmRXZERpdXFab2M0bmpsdWRiWkFwdktyekRpRHQvTWhnK1piKzU2djlSMW50ajQ3SjRuQ2RBU20KWFRDenBHclVVSzNuNmkycXE5THl6SVZFWWoySEtsclFJTlVvN3QyUkRvbThwMHRtNFdSWWd6NnYwMlM4c2M5TwpFSjJUSE5RVWFyUnJWQTZxY0lic2RPUk5aTGREVnRXWndkY25WTHNQaUtDakluUGczem5vd21jWjhXbHJqKzNHClg3SVYwR0FWRDNjNmxBVlc3QXFXZVRxdHRMVVNOVitlU2JubHhGRjdveDlkZ1FydjVNdUY2T3BUUnRod1N3VHoKOG5jeHBVanhvTTBHYzhRUDdvSUQ3V3FPNlFUU2dFUUJaMFpJMFY2OC9zZUV3cVQxZ2F0YnorT2hPOVduaFRKLwpkMnBPWEZNNHkycWJ2bzJ4OW52MjVsTlZBZ01CQUFHamR6QjFNQTRHQTFVZER3RUIvd1FFQXdJQ3BEQWRCZ05WCkhTVUVGakFVQmdnckJnRUZCUWNEQVFZSUt3WUJCUVVIQXdJd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBZEJnTlYKSFE0RUZnUVU2N3pEaG1aUzB3cTE0b2RqN0JoS29GQ0Z2L1l3RkFZRFZSMFJCQTB3QzRJSmJHOWpZV3hvYjNOMApNQTBHQ1NxR1NJYjNEUUVCQ3dVQUE0SUJBUUExZDV0T1B4VUIwWmh5KzdRVG5YU25SdFl1L2VhaWw3WnduK1NTCndrNDlUSlg5L0RJeFkzVFd1aTYrZE5rcVFnQ1lNbm4vN2pLOTF2Y1dLTVNMSXgxUVNlT3BqYkR5cHJYQlduVHUKWDNaeENlUkQraVFEL0pPQ3ZXZ1ljT0daSnU2MmVvVmh6bzdzZU8zVnVpRmlSOVNsRTU1TE9ETC9aaFBzRjVxWQp3NzFBZm1ZQXNXQ1ZlT3A1cjBpK3pYU0pyaDh6V2xSQllrTDhPZlppMUtDT1liYlhxaHRaZGJkeTBDQStreVVGCkN4bm00dFBwNkE1UEpVNGNhYmppWUVQRGRqOS9BMnY5SlE2dDJhVHVKaE42WUo4enVNc2NaeVJUaFlnd0lBZGsKckRLWEF4NlpndzV2ejFXMnVDTGpzQVJPUXpoVU5TR3FPajVjUVZDNklDaVRNQzZECi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K"
    clientKey:
      inline: "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFb2dJQkFBS0NBUUVBc1JOcnJUWFZuUTRycW1hSE9KNDViblcyUUtieXE4dzRnN2Z6SVlQbVcvdWVyL1VkClo3WStPeWVKd25RRXBsMHdzNlJxMUZDdDUrb3RxcXZTOHN5RlJHSTloeXBhMENEVktPN2RrUTZKdktkTFp1RmsKV0lNK3I5Tmt2TEhQVGhDZGt4elVGR3EwYTFRT3FuQ0c3SFRrVFdTM1ExYlZtY0hYSjFTN0Q0aWdveUp6NE44NQo2TUpuR2ZGcGE0L3R4bCt5RmRCZ0ZROTNPcFFGVnV3S2xuazZyYlMxRWpWZm5rbTU1Y1JSZTZNZlhZRUs3K1RMCmhlanFVMGJZY0VzRTgvSjNNYVZJOGFETkJuUEVEKzZDQSsxcWp1a0Uwb0JFQVdkR1NORmV2UDdIaE1LazlZR3IKVzgvam9UdlZwNFV5ZjNkcVRseFRPTXRxbTc2TnNmWjc5dVpUVlFJREFRQUJBb0lCQURqaGdDazNyZEt4aHAxSwpLZzJwNWREeHh3V2xtelpNZDZyNElBV1lGUnRmREc2QlVsektVZHMyckMzbWpzZlNENTdsSmR2bHZyZE1wamE0CjB4NWpURHZYUXVSMFdvK1l2R0JWdXA2cUNOeXM4Syt2bjBnL2dKZUNWRTI0NEZxM1E2YktEK1l2RUoyWmRzeVIKTVFZcjFscDJDOWg1d0V1UDFNa3hrcFUrMGpzVWdVWFpBeStVNWQ2RS9CU0s0UTZSQkZBMnY4VEViMGxWdGpVWgpaajRiRUxNL2Z1MXpibEFwSWc3Q3A0d2lObktXRjM3N0IyUEl6eGhIMWNmL1VmSVFTL0h3bDRCelV0c1hiRVlZCnU4UXQ3c2NFdElqeFkwSVI2NGVUNGZDYklNVzNEV2cydWR3WFFWSnJpdEh5UGRUaFRwYm04bFJsMW1sSHJMa3YKdXBUd1k0RUNnWUVBenhTYUVPTFpIRVFsUStTelBVM0l4S1pTdXNNQStFZkVncEtuVUZKMEVyZ3oxOHEyYzROWApCRnVRNU5uYlhKQ2dHZWUycmlGUmxNUm5UeFM2MWFMZDFBVXF5SnFyay9jaTNCQjZyejVUMnl3dlB6aHhSM1JXCmdrMGxYcW5xVGxHd3pENHVmWmEwNlJpMnZ6YzF0M3BpN3RZNldXeXRiTGRQMUN6L0pHZTBpMzBDZ1lFQTJ1aEMKaFdMUXdtY203YmRyUE80VkJROXZjVlFsVHQ2RDFJd2tHT0VkdkF0SzJJeXZGSWdDUzh0cm9EVG1PazE2NmtINwo3OGdiOGNmOXhEZERaNnpqYjd1R3lrTVQ3SkRxOWlKUTdMdzljaVA4QnVoQUtSejdNcXcvWFk2MnJRcUJQd1NkClZQRFNERVJjMkpqcHhpSUlPeW1ueGdDTTlSbFVQVWRVK3NxK2Zya0NnWUIra3I4Zzl5ZHhpWTJsbEJLaXMvcTEKaUZ3azM3Q21FV2ZoejdZSStIME9QQjBrRnpteUhXT0F2Rjh5SXA5Y1V1SXBNMktMeUwzT3lzWENwbzhVcWZvZwo4QStZa2tHeHJXdFhTNU5ScmkwZldFQ0F5Z1VqZ2M2bTBuUzNDZkMzY21NNFZBR2lyZzFpTk1MdTJkWXhrZE1LCjNWTEkrZzUrMXdVcVVWNmFaL0VKR1FLQmdCQlRzbUp3ZEZHTGtBTzY0bXl3OVRCamJsUnRpanJQcmRWMGZseTgKclpNUTVJd3lNZnkrQ0MzUEJqLzBzaGMzSUN2SXNCbTZPeHRWWnovelB6dkVVVkpNRWttVHB6REZ2a0NOWHF2SgpmbXU4ODFjd2kxaUZxTmFtc2pNd0tiL09RTVdLZXBHVFJKZFZvZmNsc0ludWo5Nlp4TUduMk51UEFCRngrSXljCkFvbEJBb0dBUmdLeUlKa2xocUN4elFUUEJQK0VpM2ZXNzV3NWp5bjQ2N0dqQVQ1NVRkb1VsYWJxTTVPTTJwUkMKWXByMTEyNnZEdkU3VDJWdkcwS1RqRFJoai82YmFnSjE5ZTNqc2twQVZxdGUxM3lGUFk4ZTdaMkNKU1hBUS9FVQpsL2grcnJxb0ozNjNRdVB4eGhCWDRTMkMxRG9ndWlrSHprMW5iNUdCeXN1WjVzeE9RbE09Ci0tLS0tRU5EIFJTQSBQUklWQVRFIEtFWS0tLS0tCg=="
`, meshName)

		Expect(universal.Cluster.Install(YamlUniversal(es))).To(Succeed())

		// then
		Eventually(func(g Gomega) {
			stdout, stderr, err := client.CollectResponse(universal.Cluster, "demo-client", "testmtls.mesh", client.WithVerbose())
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))
			g.Expect(stdout).To(ContainSubstring("TLSv1.2 Authentication OK!"))
			g.Expect(stdout).To(ContainSubstring("CN=kuma"))
		}).Should(Succeed())
	})
}
