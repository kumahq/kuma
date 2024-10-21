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
			stdout, _, err := client.CollectResponse(
				universal.Cluster, clientName, url,
				client.WithVerbose(),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
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
      inline: "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURTakNDQWpLZ0F3SUJBZ0lRUksrd2dOYWpKN3FKTURtR0x2aEFhekFOQmdrcWhraUc5dzBCQVFVRkFEQS8KTVNRd0lnWURWUVFLRXh0RWFXZHBkR0ZzSUZOcFoyNWhkSFZ5WlNCVWNuVnpkQ0JEYnk0eEZ6QVZCZ05WQkFNVApEa1JUVkNCU2IyOTBJRU5CSUZnek1CNFhEVEF3TURrek1ESXhNVEl4T1ZvWERUSXhNRGt6TURFME1ERXhOVm93ClB6RWtNQ0lHQTFVRUNoTWJSR2xuYVhSaGJDQlRhV2R1WVhSMWNtVWdWSEoxYzNRZ1EyOHVNUmN3RlFZRFZRUUQKRXc1RVUxUWdVbTl2ZENCRFFTQllNekNDQVNJd0RRWUpLb1pJaHZjTkFRRUJCUUFEZ2dFUEFEQ0NBUW9DZ2dFQgpBTit2NlpkUUNJTlh0TXhpWmZhUWd1ekgweXhyTU1wYjdObkRmY2RBd1JnVWkrRG9NM1pKS3VNL0lVbVRyRTRPCnJ6NUl5Mlh1L05NaEQyWFNLdGt5ajR6bDkzZXdFbnUxbGNDSm82bTY3WE11ZWd3R01vT2lmb29VTU0wUm9PRXEKT0xsNUNqSDlVTDJBWmQrM1VXT0R5T0tJWWVwTFlZSHNVbXU1b3VKTEdpaWZTS09lRE5vSmpqNFhMaDdkSU45Ygp4aXFLcXk2OWNLM0ZDeG9sa0hSeXhYdHFxelRXTUluLzVXZ1RlMVFMeU5hdTdGcWNraDQ5WkxPTXh0Ky95VUZ3CjdCWnkxU2JzT0ZVNVE5RDgvUmhjUVBHWDY5V2FtNDBkdXRvbHVjYlkzOEVWQWpxcjJtN3hQaTcxWEFpY1BOYUQKYWVRUW14a3F0aWxYNCtVOW01L3dBbDBDQXdFQUFhTkNNRUF3RHdZRFZSMFRBUUgvQkFVd0F3RUIvekFPQmdOVgpIUThCQWY4RUJBTUNBUVl3SFFZRFZSME9CQllFRk1TbnNhUjdMSEg2MitGTGtIWC94QlZnaFlrUU1BMEdDU3FHClNJYjNEUUVCQlFVQUE0SUJBUUNqR2l5YkZ3QmNxUjd1S0dZM09yK0R4ejlMd3dtZ2xTQmQ0OWxaUk5JK0RUNjkKaWt1Z2RCL09FSUtjZEJvZGZwZ2EzY3NUUzdNZ1JPU1I2Y3o4ZmFYYmF1WCs1djNnVHQyM0FEcTFjRW12OHVYcgpBdkhSQW9zWnk1UTZYa2pFR0I1WUdWOGVBbHJ3RFBHeHJhbmNXWWFMYnVtUjlZYksrcmxtTTZwWlc4N2lweFp6ClI4c3J6Sm13TjBqUDQxWkw5YzhQREhJeWg4YndSTHRUY20xRDlTWkltbEpudDFpci9tZDJjWGpiRGFKV0ZCTTUKSkRHRm9xZ0NXakJINGQxUUI3d0NDWkFBNjJSallKc1d2SWpKRXViU2ZaR0wrVDB5aldXMDZYeXhWM2JxeGJZbwpPYjhWWlJ6STluZVdhZ3FOZHd2WWtRc0VqZ2ZiS2JZSzdwMkNOVFVRCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0KLS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUVaVENDQTAyZ0F3SUJBZ0lRUUFGMUJJTVVwTWdoaklTcERCYk4zekFOQmdrcWhraUc5dzBCQVFzRkFEQS8KTVNRd0lnWURWUVFLRXh0RWFXZHBkR0ZzSUZOcFoyNWhkSFZ5WlNCVWNuVnpkQ0JEYnk0eEZ6QVZCZ05WQkFNVApEa1JUVkNCU2IyOTBJRU5CSUZnek1CNFhEVEl3TVRBd056RTVNakUwTUZvWERUSXhNRGt5T1RFNU1qRTBNRm93Ck1qRUxNQWtHQTFVRUJoTUNWVk14RmpBVUJnTlZCQW9URFV4bGRDZHpJRVZ1WTNKNWNIUXhDekFKQmdOVkJBTVQKQWxJek1JSUJJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQkNnS0NBUUVBdXdJVktNejJvSlRURHhMcwpqVldTdy9pQzhabW1la0tJcDEwbXFyVXJ1Y1ZNc2ErT2EvbDF5S1BYRDBlVUZGVTFWNHllcUtJNUdmV0NQRUtwClRtNzFPOE11MjQzQXNGenpXVGpuN2M5cDhGb0xHNzdBbENRbGgvbzNjYk1UNXh5czRadnYyK1E3UlZKRmxxbkIKVTg0MHlGTHV0YTd0ajk1Z2NPS2xWS3UyYlE2WHBVQTBheXZUdkdiclpqUjgrbXVMajFjcG1mZ3dGMTI2Y20vNwpnY1d0MG9aWVBSZkg1d203OFN2M2h0ekIybkZkMUVianpLMGx3WWk4WUdkMVpyUHhHUGVpWE9aVC96cUl0a2VsCi94TVk2cGdKZHorZFUvblBBZVgxcG5BWEZLOWpwUCtaczVPZDNGT25CdjVJaFIyaGFhNGxkYnNUekZJRDllMVIKb1l2YkZRSURBUUFCbzRJQmFEQ0NBV1F3RWdZRFZSMFRBUUgvQkFnd0JnRUIvd0lCQURBT0JnTlZIUThCQWY4RQpCQU1DQVlZd1N3WUlLd1lCQlFVSEFRRUVQekE5TURzR0NDc0dBUVVGQnpBQ2hpOW9kSFJ3T2k4dllYQndjeTVwClpHVnVkSEoxYzNRdVkyOXRMM0p2YjNSekwyUnpkSEp2YjNSallYZ3pMbkEzWXpBZkJnTlZIU01FR0RBV2dCVEUKcDdHa2V5eHgrdHZoUzVCMS84UVZZSVdKRURCVUJnTlZIU0FFVFRCTE1BZ0dCbWVCREFFQ0FUQS9CZ3NyQmdFRQpBWUxmRXdFQkFUQXdNQzRHQ0NzR0FRVUZCd0lCRmlKb2RIUndPaTh2WTNCekxuSnZiM1F0ZURFdWJHVjBjMlZ1ClkzSjVjSFF1YjNKbk1Ed0dBMVVkSHdRMU1ETXdNYUF2b0MyR0syaDBkSEE2THk5amNtd3VhV1JsYm5SeWRYTjAKTG1OdmJTOUVVMVJTVDA5VVEwRllNME5TVEM1amNtd3dIUVlEVlIwT0JCWUVGQlF1c3hlM1dGYkxybEFKUU9ZZgpyNTJMRk1MR01CMEdBMVVkSlFRV01CUUdDQ3NHQVFVRkJ3TUJCZ2dyQmdFRkJRY0RBakFOQmdrcWhraUc5dzBCCkFRc0ZBQU9DQVFFQTJVemd5ZldFaURjeDI3c1Q0clA4aTJ0aUVteFl0MGwrUEFLM3FCOG9ZZXZPNEM1ejcwa0gKZWpXRUh4MnRhUERZL2xhQkwyMS9XS1p1TlRZUUhIUEQ1YjF0WGdIWGJuTDdLcUM0MDFkazVWdkNhZFRRc3ZkOApTOE1Yam9oeWM5ejkvRzI5NDhrTGptRTZGbGg5ZERZclZZQTl4Mk8raEVQR09hRU9hMWVlUHluQmdQYXl2VWZMCnFqQnN0ekxoV1ZRTEdBa1hYbU5zKzVablBCeHpESk9MeGhGMkpJYmVRQWNINUgwdFpyVWxvNVpZeU9xQTdzOXAKTzViODVvM0FNL09KK0NrdEZCUXRmdkJoY0pWZDl3dmx3UHNrK3V5T3kySEk3bU54S0tnc0JUdDM3NXRlQTJUdwpVZEhraFZOY3NBS1gxSDdHTk5MT0VBRGtzZDg2d3VvWHZnPT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQ=="
    clientCert: # we can pass any client tls pair to this server so here are certs generated by kumactl generate tls-certificate --type=client --key-file=client.key --cert-file=client.pem
      inline: "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURJRENDQWdpZ0F3SUJBZ0lSQU4xeHJ2aGhyMExuSlN0czRrdDAyNkF3RFFZSktvWklodmNOQVFFTEJRQXcKRHpFTk1Bc0dBMVVFQXhNRWEzVnRZVEFlRncweU1UQTJNRGt4TURVek16QmFGdzB6TVRBMk1EY3hNRFV6TXpCYQpNQTh4RFRBTEJnTlZCQU1UQkd0MWJXRXdnZ0VpTUEwR0NTcUdTSWIzRFFFQkFRVUFBNElCRHdBd2dnRUtBb0lCCkFRQ3hFMnV0TmRXZERpdXFab2M0bmpsdWRiWkFwdktyekRpRHQvTWhnK1piKzU2djlSMW50ajQ3SjRuQ2RBU20KWFRDenBHclVVSzNuNmkycXE5THl6SVZFWWoySEtsclFJTlVvN3QyUkRvbThwMHRtNFdSWWd6NnYwMlM4c2M5TwpFSjJUSE5RVWFyUnJWQTZxY0lic2RPUk5aTGREVnRXWndkY25WTHNQaUtDakluUGczem5vd21jWjhXbHJqKzNHClg3SVYwR0FWRDNjNmxBVlc3QXFXZVRxdHRMVVNOVitlU2JubHhGRjdveDlkZ1FydjVNdUY2T3BUUnRod1N3VHoKOG5jeHBVanhvTTBHYzhRUDdvSUQ3V3FPNlFUU2dFUUJaMFpJMFY2OC9zZUV3cVQxZ2F0YnorT2hPOVduaFRKLwpkMnBPWEZNNHkycWJ2bzJ4OW52MjVsTlZBZ01CQUFHamR6QjFNQTRHQTFVZER3RUIvd1FFQXdJQ3BEQWRCZ05WCkhTVUVGakFVQmdnckJnRUZCUWNEQVFZSUt3WUJCUVVIQXdJd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBZEJnTlYKSFE0RUZnUVU2N3pEaG1aUzB3cTE0b2RqN0JoS29GQ0Z2L1l3RkFZRFZSMFJCQTB3QzRJSmJHOWpZV3hvYjNOMApNQTBHQ1NxR1NJYjNEUUVCQ3dVQUE0SUJBUUExZDV0T1B4VUIwWmh5KzdRVG5YU25SdFl1L2VhaWw3WnduK1NTCndrNDlUSlg5L0RJeFkzVFd1aTYrZE5rcVFnQ1lNbm4vN2pLOTF2Y1dLTVNMSXgxUVNlT3BqYkR5cHJYQlduVHUKWDNaeENlUkQraVFEL0pPQ3ZXZ1ljT0daSnU2MmVvVmh6bzdzZU8zVnVpRmlSOVNsRTU1TE9ETC9aaFBzRjVxWQp3NzFBZm1ZQXNXQ1ZlT3A1cjBpK3pYU0pyaDh6V2xSQllrTDhPZlppMUtDT1liYlhxaHRaZGJkeTBDQStreVVGCkN4bm00dFBwNkE1UEpVNGNhYmppWUVQRGRqOS9BMnY5SlE2dDJhVHVKaE42WUo4enVNc2NaeVJUaFlnd0lBZGsKckRLWEF4NlpndzV2ejFXMnVDTGpzQVJPUXpoVU5TR3FPajVjUVZDNklDaVRNQzZECi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K"
    clientKey:
      inline: "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFb2dJQkFBS0NBUUVBc1JOcnJUWFZuUTRycW1hSE9KNDViblcyUUtieXE4dzRnN2Z6SVlQbVcvdWVyL1VkClo3WStPeWVKd25RRXBsMHdzNlJxMUZDdDUrb3RxcXZTOHN5RlJHSTloeXBhMENEVktPN2RrUTZKdktkTFp1RmsKV0lNK3I5Tmt2TEhQVGhDZGt4elVGR3EwYTFRT3FuQ0c3SFRrVFdTM1ExYlZtY0hYSjFTN0Q0aWdveUp6NE44NQo2TUpuR2ZGcGE0L3R4bCt5RmRCZ0ZROTNPcFFGVnV3S2xuazZyYlMxRWpWZm5rbTU1Y1JSZTZNZlhZRUs3K1RMCmhlanFVMGJZY0VzRTgvSjNNYVZJOGFETkJuUEVEKzZDQSsxcWp1a0Uwb0JFQVdkR1NORmV2UDdIaE1LazlZR3IKVzgvam9UdlZwNFV5ZjNkcVRseFRPTXRxbTc2TnNmWjc5dVpUVlFJREFRQUJBb0lCQURqaGdDazNyZEt4aHAxSwpLZzJwNWREeHh3V2xtelpNZDZyNElBV1lGUnRmREc2QlVsektVZHMyckMzbWpzZlNENTdsSmR2bHZyZE1wamE0CjB4NWpURHZYUXVSMFdvK1l2R0JWdXA2cUNOeXM4Syt2bjBnL2dKZUNWRTI0NEZxM1E2YktEK1l2RUoyWmRzeVIKTVFZcjFscDJDOWg1d0V1UDFNa3hrcFUrMGpzVWdVWFpBeStVNWQ2RS9CU0s0UTZSQkZBMnY4VEViMGxWdGpVWgpaajRiRUxNL2Z1MXpibEFwSWc3Q3A0d2lObktXRjM3N0IyUEl6eGhIMWNmL1VmSVFTL0h3bDRCelV0c1hiRVlZCnU4UXQ3c2NFdElqeFkwSVI2NGVUNGZDYklNVzNEV2cydWR3WFFWSnJpdEh5UGRUaFRwYm04bFJsMW1sSHJMa3YKdXBUd1k0RUNnWUVBenhTYUVPTFpIRVFsUStTelBVM0l4S1pTdXNNQStFZkVncEtuVUZKMEVyZ3oxOHEyYzROWApCRnVRNU5uYlhKQ2dHZWUycmlGUmxNUm5UeFM2MWFMZDFBVXF5SnFyay9jaTNCQjZyejVUMnl3dlB6aHhSM1JXCmdrMGxYcW5xVGxHd3pENHVmWmEwNlJpMnZ6YzF0M3BpN3RZNldXeXRiTGRQMUN6L0pHZTBpMzBDZ1lFQTJ1aEMKaFdMUXdtY203YmRyUE80VkJROXZjVlFsVHQ2RDFJd2tHT0VkdkF0SzJJeXZGSWdDUzh0cm9EVG1PazE2NmtINwo3OGdiOGNmOXhEZERaNnpqYjd1R3lrTVQ3SkRxOWlKUTdMdzljaVA4QnVoQUtSejdNcXcvWFk2MnJRcUJQd1NkClZQRFNERVJjMkpqcHhpSUlPeW1ueGdDTTlSbFVQVWRVK3NxK2Zya0NnWUIra3I4Zzl5ZHhpWTJsbEJLaXMvcTEKaUZ3azM3Q21FV2ZoejdZSStIME9QQjBrRnpteUhXT0F2Rjh5SXA5Y1V1SXBNMktMeUwzT3lzWENwbzhVcWZvZwo4QStZa2tHeHJXdFhTNU5ScmkwZldFQ0F5Z1VqZ2M2bTBuUzNDZkMzY21NNFZBR2lyZzFpTk1MdTJkWXhrZE1LCjNWTEkrZzUrMXdVcVVWNmFaL0VKR1FLQmdCQlRzbUp3ZEZHTGtBTzY0bXl3OVRCamJsUnRpanJQcmRWMGZseTgKclpNUTVJd3lNZnkrQ0MzUEJqLzBzaGMzSUN2SXNCbTZPeHRWWnovelB6dkVVVkpNRWttVHB6REZ2a0NOWHF2SgpmbXU4ODFjd2kxaUZxTmFtc2pNd0tiL09RTVdLZXBHVFJKZFZvZmNsc0ludWo5Nlp4TUduMk51UEFCRngrSXljCkFvbEJBb0dBUmdLeUlKa2xocUN4elFUUEJQK0VpM2ZXNzV3NWp5bjQ2N0dqQVQ1NVRkb1VsYWJxTTVPTTJwUkMKWXByMTEyNnZEdkU3VDJWdkcwS1RqRFJoai82YmFnSjE5ZTNqc2twQVZxdGUxM3lGUFk4ZTdaMkNKU1hBUS9FVQpsL2grcnJxb0ozNjNRdVB4eGhCWDRTMkMxRG9ndWlrSHprMW5iNUdCeXN1WjVzeE9RbE09Ci0tLS0tRU5EIFJTQSBQUklWQVRFIEtFWS0tLS0tCg=="
`, meshName)

		Expect(universal.Cluster.Install(YamlUniversal(es))).To(Succeed())

		// then
		Eventually(func(g Gomega) {
			stdout, _, err := client.CollectResponse(universal.Cluster, "demo-client", "testmtls.mesh")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
			g.Expect(stdout).To(ContainSubstring("TLSv1.2 Authentication OK!"))
			g.Expect(stdout).To(ContainSubstring("CN=kuma"))
		}).Should(Succeed())
	})
}
