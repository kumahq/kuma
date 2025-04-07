package mtls

import (
	"fmt"
	"net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/universal"
)

func Policy() {
	meshName := "mtls-test"

	AfterEachFailure(func() {
		DebugUniversal(universal.Cluster, meshName)
	})

	E2EAfterEach(func() {
		Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	curlAddr := func(addr string, fn ...client.CollectResponsesOptsFn) (string, string, error) {
		fn = append(fn, client.WithVerbose())

		return client.CollectResponse(
			universal.Cluster, "demo-client", addr, fn...,
		)
	}
	trafficAllowed := func(addr string, fn ...client.CollectResponsesOptsFn) {
		GinkgoHelper()
		expectation := func(g Gomega) {
			_, stderr, err := curlAddr(addr, fn...)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))
		}
		Eventually(expectation, "30s", "250ms").MustPassRepeatedly(10).Should(Succeed())
	}
	trafficBlocked := func(addr string, fn ...client.CollectResponsesOptsFn) {
		GinkgoHelper()
		expectation := func(g Gomega) {
			_, _, err := curlAddr(addr, fn...)
			g.Expect(err).To(HaveOccurred())
		}
		Eventually(expectation, "30s", "250ms").MustPassRepeatedly(10).Should(Succeed())
	}

	Describe("should support mode", func() {
		publicAddress := func() string {
			return net.JoinHostPort(universal.Cluster.GetApp("test-server").GetIP(), "80")
		}
		setupTest := func(mode string) {
			meshYaml := fmt.Sprintf(
				`
type: Mesh
name: %s
mtls:
  enabledBackend: ca-1
  backends:
  - name: ca-1
    type: builtin
    mode: %s`, meshName, mode)
			err := NewClusterSetup().
				Install(YamlUniversal(meshYaml)).
				Install(TestServerUniversal("test-server", meshName, WithArgs([]string{"echo", "--instance", "echo-v1"}))).
				Install(DemoClientUniversal("demo-client", meshName, WithTransparentProxy(true))).
				Install(MeshTrafficPermissionAllowAllUniversal(meshName)).
				Setup(universal.Cluster)
			Expect(err).ToNot(HaveOccurred())
		}
		It("PERMISSIVE server without TLS", func() {
			setupTest("PERMISSIVE")
			By("Check inside-mesh communication")
			trafficAllowed("test-server.mesh")

			By("Check outside-mesh communication")
			trafficAllowed(publicAddress())
		})
		It("STRICT server without TLS", func() {
			setupTest("STRICT")
			By("Check inside-mesh communication")
			trafficAllowed("test-server.mesh")

			By("Check outside-mesh communication")
			trafficBlocked(publicAddress())
		})
		It("PERMISSIVE server with TLS", func() {
			setupTest("PERMISSIVE")
			By("Check inside-mesh communication")
			trafficAllowed("test-server.mesh", client.WithCACert("/kuma/server.crt"))

			By("Check outside-mesh communication")
			// we're using curl with '--resolve' flag to verify certificate Common Name 'test-server.mesh'
			host := universal.Cluster.GetApp("test-server").GetIP()
			trafficAllowed(publicAddress(),
				client.WithCACert("/kuma/server.crt"),
				client.Resolve("test-server.mesh:80", fmt.Sprintf("[%s]", host)))
		})
		It("STRICT server with TLS", func() {
			setupTest("STRICT")
			By("Check inside-mesh communication")
			trafficAllowed("test-server.mesh", client.WithCACert("/kuma/server.crt"))

			By("Check outside-mesh communication")
			// we're using curl with '--resolve' flag to verify certificate Common Name 'test-server.mesh'
			host := universal.Cluster.GetApp("test-server").GetIP()
			trafficBlocked(publicAddress(),
				client.WithCACert("/kuma/server.crt"),
				client.Resolve("test-server.mesh:80", fmt.Sprintf("[%s]", host)))
		})
	})

	It("enabling PERMISSIVE with no failed requests", func() {
		Expect(universal.Cluster.Install(MeshUniversal(meshName))).To(Succeed())

		// Disable retries so that we see every failed request
		kumactl := universal.Cluster.GetKumactlOptions()
		Expect(kumactl.KumactlDelete("meshretry", fmt.Sprintf("mesh-retry-all-%s", meshName), meshName)).To(Succeed())

		// We must start client before server to test this properly. The client
		// should get XDS refreshes first to trigger the race condition fixed by
		// kumahq/kuma#3019
		Expect(universal.Cluster.Install(DemoClientUniversal("demo-client", meshName, WithTransparentProxy(true)))).To(Succeed())

		Expect(universal.Cluster.Install(TestServerUniversal("test-server", meshName, WithArgs([]string{"echo", "--instance", "echo-v1"})))).To(Succeed())

		By("Check inside-mesh communication")
		trafficAllowed("test-server.mesh")

		By("Enable permissions mtls")
		Expect(universal.Cluster.Install(MeshTrafficPermissionAllowAllUniversal(meshName))).To(Succeed())
		meshYaml := fmt.Sprintf(
			`
type: Mesh
name: %s
mtls:
  enabledBackend: ca-1
  backends:
  - name: ca-1
    type: builtin
    mode: PERMISSIVE`, meshName)
		Expect(universal.Cluster.Install(YamlUniversal(meshYaml))).To(Succeed())

		By("inside-mesh communication never fails")
		Consistently(func(g Gomega) {
			_, stderr, err := curlAddr("test-server.mesh")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))
		}).Should(Succeed())
	})

	DescribeTable("should enforce traffic permissions",
		func(yaml string) {
			err := NewClusterSetup().
				Install(MeshUniversal(meshName)).
				Install(TestServerUniversal("test-server", meshName, WithArgs([]string{"echo", "--instance", "echo-v1"}))).
				Install(DemoClientUniversal("demo-client", meshName, WithTransparentProxy(true))).
				Setup(universal.Cluster)
			Expect(err).ToNot(HaveOccurred())

			By("Can access test-server")
			trafficAllowed("test-server.mesh")

			By("When adding mTLS mesh")
			Expect(universal.Cluster.Install(YamlUniversal(yaml))).To(Succeed())

			By("Can't access test-server")
			trafficBlocked("test-server.mesh")

			By("When adding traffic-permission")
			perm := fmt.Sprintf(`
type: TrafficPermission
name: example
mesh: "%s"
sources:
  - match:
      kuma.io/service: demo-client
destinations:
  - match:
      kuma.io/service: test-server
`, meshName)
			Expect(universal.Cluster.Install(YamlUniversal(perm))).To(Succeed())

			By("Can access test-server again")
			trafficAllowed("test-server.mesh")
		},
		Entry("builtin CA", fmt.Sprintf(`
type: Mesh
name: "%s" 
mtls:
  enabledBackend: ca-builtin
  backends:
  - name: ca-builtin
    type: builtin
`, meshName),
		),
		Entry("Provided CA with Root CA", fmt.Sprintf(`
type: Mesh
name: "%s" 
mtls:
  enabledBackend: ca-provided-root
  backends:
  - name: ca-provided-root
    type: provided
    conf:
      cert:
        inline: "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURiakNDQWxhZ0F3SUJBZ0lKQUxETU1hOXJYS0xQTUEwR0NTcUdTSWIzRFFFQkN3VUFNRVF4Q3pBSkJnTlYKQkFZVEFrZENNUkF3RGdZRFZRUUlEQWRGYm1kc1lXNWtNUkl3RUFZRFZRUUtEQWxCYkdsalpTQk1kR1F4RHpBTgpCZ05WQkFNTUJrdDFiV0ZEUVRBZUZ3MHlNVEExTVRJeE16RTJNakZhRncwME1UQTFNRGN4TXpFMk1qRmFNRVF4CkN6QUpCZ05WQkFZVEFrZENNUkF3RGdZRFZRUUlEQWRGYm1kc1lXNWtNUkl3RUFZRFZRUUtEQWxCYkdsalpTQk0KZEdReER6QU5CZ05WQkFNTUJrdDFiV0ZEUVRDQ0FTSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQwpnZ0VCQU5DSnFWSmpZT1dGVVpjZGhyZkJ4Z29DWk5FK0xGcTlzaWVQMnlSR3JZekpzQ2R3cGhINkw3R3NXZHM4ClZqbG9iZklQNG5BMjNUSmlNV2xzeDEyNnI3cFNSYlZFcTgvSm9OYTB2TXNwRW10akhaaFN3ZUlYV1hYN284VisKRlJLYkNXNU55cUdpSEYwU2NFNFZwTmMzdVdDQTJ6Y2FVODBHOVNBS0k4M2NVam5wMkp6TFBNcXBwUStwajZIcwpHKzgzMjJGUEEyTDExZnNDQXFkQ1crZ3dKV3BLemxmQlB5ZU5UVU9NcGNQOG4rWWpjYWg0dHFjQ1kyUFo3bkg3CmNaTjF2SEdoVDUvUG4zVlJhTkhVcTR5MVpuL3dKbmpsT2NENERiVkZYWXBZSWxQeCt5QXM1NkZYZDNhN0ltZmcKNTZIek9MT1pjRFkvK1N4eTdKMlBxOGNpcFRjQ0F3RUFBYU5qTUdFd0hRWURWUjBPQkJZRUZFcnRoT0ZIdVdqOAozVmtBZ2phZCtqMzk0bUczTUI4R0ExVWRJd1FZTUJhQUZFcnRoT0ZIdVdqODNWa0FnamFkK2ozOTRtRzNNQThHCkExVWRFd0VCL3dRRk1BTUJBZjh3RGdZRFZSMFBBUUgvQkFRREFnRUdNQTBHQ1NxR1NJYjNEUUVCQ3dVQUE0SUIKQVFDQnFqOUYrT0paWGlmeVVHcTliQWl5YnBQOVJZbktkMEpDaUJ5dk8vUzk1djZCejlSbndydmdONzVtenBQZApPTTUxTVlLeUJMRktKcHZybXlRK25qY3NWTW52Ly9NSDdjSEU4aDZXa3dQOUlnZ05nMEsyMUoxemtTOEFwZlR3CjdidVVlbVpuNk5GcUhneXNBVW5XcThXTThZeGZFRXJ1YmJUQ202d3NsVEx6TGRibEJHTGpoN3FPekRHaDhuMGUKQmpxV2dDWWpiRXNCNHREeGpmU2pMalN5bGR2bklNVHlXckE4YS8xaUNORFhqMHdNdEhvQmppMzA3ZHNJNWRycApWb2tFTHdldTZTUzdNNE9ERTgvQ2kzUUxTL21teCsrOXMya0NDcXE0OWR5QTIvWmFiTGIybkJGOTZ3by9SRHA5CjNrSXpmTnZ6TWtDM1ZSd0VTVitTVUcweAotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg=="
      key:
        inline: "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcEFJQkFBS0NBUUVBMEltcFVtTmc1WVZSbHgyR3Q4SEdDZ0prMFQ0c1dyMnlKNC9iSkVhdGpNbXdKM0NtCkVmb3ZzYXhaMnp4V09XaHQ4Zy9pY0RiZE1tSXhhV3pIWGJxdnVsSkZ0VVNyejhtZzFyUzh5eWtTYTJNZG1GTEIKNGhkWmRmdWp4WDRWRXBzSmJrM0tvYUljWFJKd1RoV2sxemU1WUlEYk54cFR6UWIxSUFvanpkeFNPZW5Zbk1zOAp5cW1sRDZtUG9ld2I3emZiWVU4RFl2WFYrd0lDcDBKYjZEQWxha3JPVjhFL0o0MU5RNHlsdy95ZjVpTnhxSGkyCnB3SmpZOW51Y2Z0eGszVzhjYUZQbjgrZmRWRm8wZFNyakxWbWYvQW1lT1U1d1BnTnRVVmRpbGdpVS9IN0lDem4Kb1ZkM2Ryc2laK0Rub2ZNNHM1bHdOai81TEhMc25ZK3J4eUtsTndJREFRQUJBb0lCQVFDWTlHK3FDMVExNU14TQpYNDdCVnpKdmd3UDVhWFhVOUpBb2JsNVl4REpsTWtXdkYvUG0rYTlqelR0M0QxRmErQnEwVWl4UERCNi81ci9CClNOVU9EWEQ0NDRGWGpFL01yMkgyT1VqRVpwS3BDMkRZcWRLbGgyVEpvZEdrZUc3eVg1N1NOZUtySFNGYXJyeUsKdVZ0WHMvcVhLc3dmSllOVHZZZXJnV1J6aU9jU3EveUtZR2s4TDE5Nk9yUlUzL1pmbWpBWk1zN1JTN3pwTEJNSQpKc3VkczUvTUlUdzROMWJwTkJ3KzJyYkhqR2owS2VaNjBDbzJ4QUY0WUQ1VGxxWi95Y01FT2g5NVJnTU1iUjRkCmpzNXRNK0RlMHlGcG11amM3L3U0NnRPOENHT3M0eG1zK1JBenpycFBMRnBUUTFDZW5BempTQ0s1akpvZUltU3gKUEVua0xUK3BBb0dCQU9uYVZTeWtCU0F5dzl3NVMvcmlPc0xBUU9LNG90a3krSTduMDVpSDVDdk9HV2VIM3RmdgpZMHJjTm5mNHBCSmdBS0dnMk9pakJPQ2dqTjcwMGJJU3pPRWJPN3RKdGxJd2wvdEhKR0t3WTk1WFlWRUtWV2hNCjlDbVV2QUl0dzNPSnYxQlVhRnNQVnpLVW8rTUtQYXdBVlZZbVJEMnNqSGRxM3hYa0UvZEFHcVpiQW9HQkFPUkoKa3VvVEVxZmkrOG5xME1HczdTblMzSVl1enQ5TEtSZEZnZTBjWUl1a09yQVBHNXhuYjBpTnZZVXhUM0Y0ZHp0WgpXcDVrekoxUE1vdW9jZXAxQ0VHMmlsN3RwaWl4MTF4VFdEYkdTalFPa1pIaDVKaGJmeUVOcWl6eG04dEttSXRkCkNhUzlLb0dqbHlnd3dwTStrNjJtZmIza0tWTENLbm42MXd6dndvdFZBb0dCQU5UcEo2c2hHbG1hWHFCZXVrS04KUHBxWmwzblVTTkFmakJYd0U4Skgxd0hhLzE0M1lqaVBoNE5jdzJxdlFoSkl2Y3BxTzVKeStiblo4dWY3VmdBZgpCZEhkamFDVEdCLzBoaXNOTnA5em9UbUpyTnl2MzlxNlZZS1dIQ2FQcStmQmZpR1ErQUlRRVgvSHZQNjFFRGxOCmhHU1BLb3BNVXdkV2tnM0lQalZhYytrSkFvR0FJRzRMUHRGaXp4TEJyaGQ3ZkdmeWNRU1JhMFp2QU8yT2NzM2UKL1M0UTBRV05pTUU4ck9WTXU4UFc3bnJvekRmT3lGR1RPL2taMENjV0NSenV3ZDNLUkh1SUFLQkdBSFh6SUJ4Kwo1WmtacFhlRVduTDZwR0lyRnlqM3lkYXd5UnBadlVLRVFqRFZQd0ZjVWN0TGVOdGs0MEJKa0pZL0FKQ3d0QTljClNXd3QrTmtDZ1lCa2hjZk9ZTXhlWUd6Q0JCYVZMb0pKUEY2bUowUXRHRVZaYzRsdThEam81MXNKbVpyMW1XWmwKS2ZlTFArdncyc2o4SFZzRlBwVHppTzhFTmZzVUlObXpsajhHamFwcGdJNks4dTQrYVlBRStmMEhHZzFzZ08ydQpMNDhja2JPV05FUS9nQ0NTemY1MHJtQ2FuQ0VDbEUranZCWXZqZGhZclppNWltWWtYMytOakE9PQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo="
`, meshName)),
		Entry("Provided CA with Intermediate CA", fmt.Sprintf(`
type: Mesh
name: "%s" 
mtls:
  enabledBackend: ca-provided-intermediate
  backends:
  - name: ca-provided-intermediate
    type: provided
    conf:
      cert:
        inline: "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURkakNDQWw2Z0F3SUJBZ0lDRUFFd0RRWUpLb1pJaHZjTkFRRUxCUUF3UkRFTE1Ba0dBMVVFQmhNQ1IwSXgKRURBT0JnTlZCQWdNQjBWdVoyeGhibVF4RWpBUUJnTlZCQW9NQ1VGc2FXTmxJRXgwWkRFUE1BMEdBMVVFQXd3RwpTM1Z0WVVOQk1CNFhEVEl4TURVeE1qRXpNelUxTVZvWERUTXhNRFV4TURFek16VTFNVm93VURFTE1Ba0dBMVVFCkJoTUNSMEl4RURBT0JnTlZCQWdNQjBWdVoyeGhibVF4RWpBUUJnTlZCQW9NQ1VGc2FXTmxJRXgwWkRFYk1Ca0cKQTFVRUF3d1NTM1Z0WVVsdWRHVnliV1ZrYVdGMFpVTkJNSUlCSWpBTkJna3Foa2lHOXcwQkFRRUZBQU9DQVE4QQpNSUlCQ2dLQ0FRRUExVnpZOXZPcjgrU0lOenFBOFJ3azRicGVleDMyWm45QkdBVVR3ZVJnb21RQzdZZnpybTYvClZrNzQvVC80Nm4zRnlkcGRFWlRkb0ZLQ0Y4RXNBMGVxQUVmV2k2dHU3RDQxR09VRlVZcGRSSkJKRXErSEUxN1EKTjhTRk1xdXk4TmhDdEs4dGg4eXRTdTJUaHZDT3ExTUhUNVdqdFFVbVJHU0pNbGNmV0E1VHNDSUswU2IzY1NmMwpqYWRqRXFjbWN2Sk42WGEwWTBWaXZjUGc1ZUIrV2U3Qk5ucDRvZ3FtWncwdmVvUGpjMTRIVlpwcXhycmE5WWV6CkRSYWk2cm5IcURqbmtNTWhlOU1tU2tDS0Q5TGR3ZHVxMFpmdU9RRklCT2FYKzRNS1V5RE40dFRNQ2NSUmwvTmwKQTRKZ3JOTldDRmZVUVYwVm1RMFRjOCtjbi8rZ29rSEFad0lEQVFBQm8yWXdaREFkQmdOVkhRNEVGZ1FVR05qegpUZTcyN0hYNEFxWkRNbjFMOVh6a1RhWXdId1lEVlIwakJCZ3dGb0FVU3UyRTRVZTVhUHpkV1FDQ05wMzZQZjNpClliY3dFZ1lEVlIwVEFRSC9CQWd3QmdFQi93SUJBREFPQmdOVkhROEJBZjhFQkFNQ0FRWXdEUVlKS29aSWh2Y04KQVFFTEJRQURnZ0VCQUN1T2N6SmxmNHdjVDlyZkFJclpIdUk1YUN6WVRLT3hKbGxoTjVlL2VFaE1ZcHNveDZaYgo0Q1pYUzN3ZEozZlZ1Z2RkTFdEeklBanJORTFEck9wdWdVUHVyTklwSHNUNnUrU0hGWGtSc1h5SEZmTUErQ1pKCjB0T1lFdFAxcjNCbnFzWS9uaDBHSnFISnhhSm9sRWFxRmFLZ0tUUVBUaW5PeFRLRnhzSGExT0hsc3ZrZHh2b3QKZDJCUWhQUVlXZXMzTE1QeHRHaFM1a3dLYVhhQjNnelRuempHdmdHTmVKK2wwQWlXcVhraXZpeHBveDMvNm1NYQo5MG13c3NsNHNSUVFMUjFrTEZVNGh3Z2hObTUyUGs3bzdIU1RFWHNuQitaaEhCOXNrcGV0WTZSNHVLV2g4eGFwClhtajRQRHJBQTVPS1p6U083WWhkdDB2WFBPSXJqU2hNeHZBPQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCi0tLS0tQkVHSU4gQ0VSVElGSUNBVEUtLS0tLQpNSUlEYmpDQ0FsYWdBd0lCQWdJSkFMRE1NYTlyWEtMUE1BMEdDU3FHU0liM0RRRUJDd1VBTUVReEN6QUpCZ05WCkJBWVRBa2RDTVJBd0RnWURWUVFJREFkRmJtZHNZVzVrTVJJd0VBWURWUVFLREFsQmJHbGpaU0JNZEdReER6QU4KQmdOVkJBTU1Ca3QxYldGRFFUQWVGdzB5TVRBMU1USXhNekUyTWpGYUZ3MDBNVEExTURjeE16RTJNakZhTUVReApDekFKQmdOVkJBWVRBa2RDTVJBd0RnWURWUVFJREFkRmJtZHNZVzVrTVJJd0VBWURWUVFLREFsQmJHbGpaU0JNCmRHUXhEekFOQmdOVkJBTU1Ca3QxYldGRFFUQ0NBU0l3RFFZSktvWklodmNOQVFFQkJRQURnZ0VQQURDQ0FRb0MKZ2dFQkFOQ0pxVkpqWU9XRlVaY2RocmZCeGdvQ1pORStMRnE5c2llUDJ5UkdyWXpKc0Nkd3BoSDZMN0dzV2RzOApWamxvYmZJUDRuQTIzVEppTVdsc3gxMjZyN3BTUmJWRXE4L0pvTmEwdk1zcEVtdGpIWmhTd2VJWFdYWDdvOFYrCkZSS2JDVzVOeXFHaUhGMFNjRTRWcE5jM3VXQ0EyemNhVTgwRzlTQUtJODNjVWpucDJKekxQTXFwcFErcGo2SHMKRys4MzIyRlBBMkwxMWZzQ0FxZENXK2d3SldwS3psZkJQeWVOVFVPTXBjUDhuK1lqY2FoNHRxY0NZMlBaN25INwpjWk4xdkhHaFQ1L1BuM1ZSYU5IVXE0eTFabi93Sm5qbE9jRDREYlZGWFlwWUlsUHgreUFzNTZGWGQzYTdJbWZnCjU2SHpPTE9aY0RZLytTeHk3SjJQcThjaXBUY0NBd0VBQWFOak1HRXdIUVlEVlIwT0JCWUVGRXJ0aE9GSHVXajgKM1ZrQWdqYWQrajM5NG1HM01COEdBMVVkSXdRWU1CYUFGRXJ0aE9GSHVXajgzVmtBZ2phZCtqMzk0bUczTUE4RwpBMVVkRXdFQi93UUZNQU1CQWY4d0RnWURWUjBQQVFIL0JBUURBZ0VHTUEwR0NTcUdTSWIzRFFFQkN3VUFBNElCCkFRQ0JxajlGK09KWlhpZnlVR3E5YkFpeWJwUDlSWW5LZDBKQ2lCeXZPL1M5NXY2Qno5Um53cnZnTjc1bXpwUGQKT001MU1ZS3lCTEZLSnB2cm15UStuamNzVk1udi8vTUg3Y0hFOGg2V2t3UDlJZ2dOZzBLMjFKMXprUzhBcGZUdwo3YnVVZW1abjZORnFIZ3lzQVVuV3E4V004WXhmRUVydWJiVENtNndzbFRMekxkYmxCR0xqaDdxT3pER2g4bjBlCkJqcVdnQ1lqYkVzQjR0RHhqZlNqTGpTeWxkdm5JTVR5V3JBOGEvMWlDTkRYajB3TXRIb0JqaTMwN2RzSTVkcnAKVm9rRUx3ZXU2U1M3TTRPREU4L0NpM1FMUy9tbXgrKzlzMmtDQ3FxNDlkeUEyL1phYkxiMm5CRjk2d28vUkRwOQoza0l6Zk52ek1rQzNWUndFU1YrU1VHMHgKLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQ=="
      key:
        inline: "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFb2dJQkFBS0NBUUVBMVZ6WTl2T3I4K1NJTnpxQThSd2s0YnBlZXgzMlpuOUJHQVVUd2VSZ29tUUM3WWZ6CnJtNi9Wazc0L1QvNDZuM0Z5ZHBkRVpUZG9GS0NGOEVzQTBlcUFFZldpNnR1N0Q0MUdPVUZVWXBkUkpCSkVxK0gKRTE3UU44U0ZNcXV5OE5oQ3RLOHRoOHl0U3UyVGh2Q09xMU1IVDVXanRRVW1SR1NKTWxjZldBNVRzQ0lLMFNiMwpjU2YzamFkakVxY21jdkpONlhhMFkwVml2Y1BnNWVCK1dlN0JObnA0b2dxbVp3MHZlb1BqYzE0SFZacHF4cnJhCjlZZXpEUmFpNnJuSHFEam5rTU1oZTlNbVNrQ0tEOUxkd2R1cTBaZnVPUUZJQk9hWCs0TUtVeURONHRUTUNjUlIKbC9ObEE0SmdyTk5XQ0ZmVVFWMFZtUTBUYzgrY24vK2dva0hBWndJREFRQUJBb0lCQUVZelhXT3JldWt0U1RBNgp2SkFZUFg1VHJxQjlsRVBYSE5qRld1SFJ2WVhRdURIbEtHVTlKUkZGdktDK3VxeVVvMDR0M2E5YU5xMTRXRHR1Cm9JZVh2YlNIN214WUJKQVFTN0ljTUdyS3hyTTNjZ09HNHorWlU1TDF3d21QK3JsSnZuRHFybVZmRDZsMmo3SzMKVlluRE1NV2JxNWgwMEVseWRyMnNjckNyVGhmR0VhanRRems5VWxqTWZhenlSODVlYXdLdmlMZGpPVERXRGtFUgpZb3ZGYjZzRkNLaUc5T1RTY3FRS296Sm5mYm42VGtyT1kzejV6WVdTbCtGS0dxbFp1Z0JhNHlkSWpYNnd2em83CjR0NWwrMjdTcXlRZFhoTWtjMkRCREgvdmZrNElMM3dwWWVuUTR4UDhxaXNYMjluakFQbUhRL0RZTnlPRHZEUTUKU0paa3lna0NnWUVBOU9LS1k4aVh3TXN1RmF5Z0lFZGNHS0RXMFZjRWNXWnNrNGVVajZtbHVnQ3k2MnZ1MUI1QgpHeW9tbG9zekNCQUs3bHRHM3JWN1pYck9BZXlVT1dtRlNwclF4aFczYUZHSTJNUkVQWlo2VlV4Tmh0S3dWczVGCkgzOGxPdUEzTGxnNHh4VnNqVGtERzBDUkJsMDBHVVlhSDIzQlE3aFB0cGlPZmxJUVRIZWJPQU1DZ1lFQTN3d0kKSmg2alBQR1NucjNZdHJDcjI4Y3M5WmZZKzZWQldYYnlrek1kODEzQnR0MDg4SmpBbCt4V2hNRFF1em9GTlJvbgo5dGdSWnZNZnNDMitGQWI0RFZlR2lpWkdiTnVkcm50T1ErdE95cTJBSTBxZnk0bHg1SUlucnBsRmp2OUxuV0NiClhjMytVb05HZVhwMzZ0VHhBMzlOS1RoajVPUWw3Q3N0MzZzc29zMENnWUFZbzVxTkE2ZlJMQ0JNNmZ1S2crT08KVHRDT2E1VDAyL3RjdEsyTDd1UFAzVFlqWGM2LzVQTmtDaytyb2dIV2M5YkZ1TVZlcngvbFMvL2lUYTEwUVZ1NQo3KzNGb0hXOXQwWnZtUC9NdXBGQWQ5YnRFOUhPU2g3R1ZvS21jOXpaZXVMcmxRcEJBMVYrcm5acEQ0T29iMWM5CmhrdUZ4c3V1Y1pjVXVxa05LSk9qaFFLQmdBNDZ3VVpWVEFxMlNxbFA2VVIyYnZCZGU1UExkUzlRc3FPWGdCQSsKQVpvbUVCYXZkSlRRMmZDWFJrbS8xMUVxZVd5UzE1dmEydmxiWjFraEFmQmJKWFlNY0d2ZDF6NVlvRzJpTmpNRgpEd2pGR3RpbGlSNCtEWU1MZnFhWDVxVWh5bHdtN3FLRVlzWTIwOGNxTmY1SVNYdjBvaUtRRTJkbDJybC9ZN1RTClFjMjlBb0dBTDQ2c1dxaDR6bzk1cTU1dVNPNlVwNDVRc1NaVlZOdERZRk0rUTVSa3hUZFFULy82bkYyYVh1Z3AKcEJ4dm1GTGs2bVV1RDJlVzk0QzRxbDBWNmFWNXJVUFNXdXk1dUhmN3k5aG1vUTkvaDZZc0lyMjFnZmsyV0lkcwpjMSthTjJYSDBQUW1OemVYbmdLUi9BNlYvYnMvVU9DdklEeWZNY1FUanRnRlhWbjdGejA9Ci0tLS0tRU5EIFJTQSBQUklWQVRFIEtFWS0tLS0t"
`, meshName)),
	)
}
