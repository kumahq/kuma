package gatewayapi

import (
	"net"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	client "github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func GatewayAPICRDs(cluster Cluster) error {
	out, err := k8s.RunKubectlAndGetOutputE(
		cluster.GetTesting(),
		cluster.GetKubectlOptions(),
		"kustomize", "github.com/kubernetes-sigs/gateway-api/config/crd?ref=v0.4.0",
	)
	if err != nil {
		return err
	}

	return k8s.KubectlApplyFromStringE(cluster.GetTesting(), cluster.GetKubectlOptions(), out)
}

const gatewayClass = `
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: GatewayClass
metadata:
  name: kuma
spec:
  controllerName: gateways.kuma.io/controller
`

var cluster *K8sCluster

var _ = E2EBeforeSuite(func() {
	if Config.IPV6 {
		return // KIND which is used for IPV6 tests does not support load balancer that is used in this test.
	}

	cluster = NewK8sCluster(NewTestingT(), Kuma1, Silent)

	err := NewClusterSetup().
		Install(GatewayAPICRDs).
		Install(Kuma(config_core.Standalone,
			WithCtlOpts(map[string]string{"--experimental-meshgateway": "true"}),
			WithCtlOpts(map[string]string{"--experimental-gatewayapi": "true"}),
		)).
		Install(NamespaceWithSidecarInjection(TestNamespace)).
		Install(testserver.Install(
			testserver.WithName("test-server-1"),
			testserver.WithNamespace(TestNamespace),
			testserver.WithArgs("echo", "--instance", "test-server-1"),
		)).
		Install(testserver.Install(
			testserver.WithName("test-server-2"),
			testserver.WithNamespace(TestNamespace),
			testserver.WithArgs("echo", "--instance", "test-server-2"),
		)).
		Install(YamlK8s(gatewayClass)).
		Setup(cluster)
	Expect(err).ToNot(HaveOccurred())

	E2EDeferCleanup(func() {
		Expect(cluster.DeleteNamespace(TestNamespace)).To(Succeed())
		Expect(cluster.DeleteKuma()).To(Succeed())
		Expect(cluster.DismissCluster()).To(Succeed())
	})
})

func GatewayAPI() {
	if Config.IPV6 {
		return // KIND which is used for IPV6 tests does not support load balancer that is used in this test.
	}

	GatewayIP := func() string {
		var ip string
		Eventually(func(g Gomega) {
			out, err := k8s.RunKubectlAndGetOutputE(
				cluster.GetTesting(),
				cluster.GetKubectlOptions(TestNamespace),
				"get", "gateway", "kuma", "-ojsonpath={.status.addresses[0].value}",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(out).ToNot(BeEmpty())
			ip = out
		}, "60s", "1s").Should(Succeed(), "could not get a LoadBalancer IP of the Gateway")
		return ip
	}

	Context("HTTP Gateway", func() {
		gateway := `
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: Gateway
metadata:
  name: kuma
  namespace: kuma-test
spec:
  gatewayClassName: kuma
  listeners:
  - name: proxy
    port: 8080
    protocol: HTTP`

		var address string

		BeforeEach(func() {
			err := k8s.RunKubectlE(cluster.GetTesting(), cluster.GetKubectlOptions(), "delete", "gateway", "--all")
			Expect(err).ToNot(HaveOccurred())
			Expect(YamlK8s(gateway)(cluster)).To(Succeed())
			address = net.JoinHostPort(GatewayIP(), "8080")
		})

		It("should route the traffic to test-server by path", func() {
			// given
			route := `
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: HTTPRoute
metadata:
  name: test-server-paths
  namespace: kuma-test
spec:
  parentRefs:
  - name: kuma
  rules:
  - backendRefs:
    - name: test-server-1
      port: 80
    matches:
    - path:
        type: PathPrefix
        value: /1
  - backendRefs:
    - name: test-server-2
      port: 80
    matches:
    - path:
        type: PathPrefix
        value: /2`

			// when
			err := YamlK8s(route)(cluster)

			// then
			Expect(err).ToNot(HaveOccurred())

			Eventually(func(g Gomega) {
				resp, err := client.CollectResponseDirectly("http://" + address + "/1")
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(resp.Instance).To(Equal("test-server-1"))
			}, "30s", "1s").Should(Succeed())

			Eventually(func(g Gomega) {
				resp, err := client.CollectResponseDirectly("http://" + address + "/2")
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(resp.Instance).To(Equal("test-server-2"))
			}, "30s", "1s").Should(Succeed())
		})

		It("should route the traffic to test-server by header", func() {
			// given
			routes := `
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: HTTPRoute
metadata:
  name: test-server-1
  namespace: kuma-test
spec:
  parentRefs:
  - name: kuma
  hostnames:
  - "test-server-1.com"
  rules:
  - backendRefs:
    - name: test-server-1
      port: 80
---
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: HTTPRoute
metadata:
  name: test-server-2
  namespace: kuma-test
spec:
  parentRefs:
  - name: kuma
  hostnames:
  - "test-server-2.com"
  rules:
  - backendRefs:
    - name: test-server-2
      port: 80
`

			// when
			err := YamlK8s(routes)(cluster)

			// then
			Expect(err).ToNot(HaveOccurred())

			Eventually(func(g Gomega) {
				resp, err := client.CollectResponseDirectly("http://"+address, client.WithHeader("host", "test-server-1.com"))
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(resp.Instance).To(Equal("test-server-1"))
			}, "30s", "1s").Should(Succeed())

			Eventually(func(g Gomega) {
				resp, err := client.CollectResponseDirectly("http://"+address, client.WithHeader("host", "test-server-2.com"))
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(resp.Instance).To(Equal("test-server-2"))
			}, "30s", "1s").Should(Succeed())
		})
	})

	Context("HTTPS Gateway", func() {
		secret := `
apiVersion: v1
kind: Secret
metadata:
  name: secret-tls
  namespace: kuma-test
type: kubernetes.io/tls
data:
  tls.crt: "MIIEOzCCAyOgAwIBAgIRANQR+/q3DZNc//4rGW+4yjowDQYJKoZIhvcNAQELBQAwHTEbMBkGA1UEAxMSa3VtYS1jb250cm9sLXBsYW5lMB4XDTIyMDEyNDEwMzA1NloXDTMyMDEyMjEwMzA1NlowHTEbMBkGA1UEAxMSa3VtYS1jb250cm9sLXBsYW5lMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAt3A6o4oOp72vLWVfMS5YaGG0tQGcAQTAoGwbceayAZ7lPlIMeQO5wMbZEPl7MC/ebqCsMZ0tLrhCKNPxvAKyXm982+iQVY9nCYziPdffdF5iY6H8P/Aq45Rl5HbjfpSHgrQbTWQFBmCl6BkWpOLF0q8NJ3WDiTs/vydIazCKNN4lNR1TAR87W/G70tqVwvGQD7V4UqEP4bkNgAUf5nbJkYM+yTNHokzdZPXrhyf2HjayszDPehD18e2CZx2aXC11QQJqGBjzAk/SaNUJrkZhsIib9HfYCpP0zkNcXZk42rO1WL9TZ46PR3MUSXkP8CYdqLkWZFsI0PVF6NYtp5pBTQIDAQABo4IBdDCCAXAwDgYDVR0PAQH/BAQDAgKkMBMGA1UdJQQMMAoGCCsGAQUFBwMBMA8GA1UdEwEB/wQFMAMBAf8wHQYDVR0OBBYEFGID0/QTirRTLKpVjYsORV8chFY1MIIBFwYDVR0RBIIBDjCCAQqCB2FudC1kZXaCCWxvY2FsaG9zdIcEZEPlAYcEfwAAAYcErBEAAYcErBIAAYcEwKgBM4cQAAAAAAAAAAAAAAAAAAAAAYcQ/QD9EjRWAAAAAAAAAAAAAYcQ/XoRXKHgqxJIQ82WYkPlAYcQ/oAAAAAAAAAAAAAAAAAAAYcQ/oAAAAAAAAAgPzb//uxIXYcQ/oAAAAAAAAAAQgj//uUH2YcQ/oAAAAAAAAAAQp3//jgBp4cQ/oAAAAAAAABQVAD//u85EocQ/oAAAAAAAABUmID//nxtvYcQ/oAAAAAAAABgFXX//rgUtIcQ/oAAAAAAAADI/3///mfBuYcQ/oAAAAAAAAD2y24uiMcHnTANBgkqhkiG9w0BAQsFAAOCAQEALdMPghMlDgRAm8Prp/AqtDFY4K7zxBhssgS5cgQkJvu7+IVk3Cj6iwNmAat6BtRXRdD851vRqD0s6OtApTeyriW+fX0qcuQW55uPm6E3BDdg5OjetaaOayziRdsy7NSCvmkkYDQQT/M1zX0WvPWNDtH8iwg5hJh8qk+p4Cc7nP/H0JPiiT5LK5lMZ8fSDz0PpFy1t1Gw7tsNHAtsz6FFh2NgQmvLi4RikRxHebEiYw9D8mK47Y+1Vq+XT7xwui6tc0BAtQJvUID+zc/k89ANybjEHSo0m5wdrNjbG0AoLqmPls9Pv4p3Hn2Q2TZUnqwnt6MvrmsbUiHXFaql8OErZA=="
  tls.key: "MIIEowIBAAKCAQEAt3A6o4oOp72vLWVfMS5YaGG0tQGcAQTAoGwbceayAZ7lPlIMeQO5wMbZEPl7MC/ebqCsMZ0tLrhCKNPxvAKyXm982+iQVY9nCYziPdffdF5iY6H8P/Aq45Rl5HbjfpSHgrQbTWQFBmCl6BkWpOLF0q8NJ3WDiTs/vydIazCKNN4lNR1TAR87W/G70tqVwvGQD7V4UqEP4bkNgAUf5nbJkYM+yTNHokzdZPXrhyf2HjayszDPehD18e2CZx2aXC11QQJqGBjzAk/SaNUJrkZhsIib9HfYCpP0zkNcXZk42rO1WL9TZ46PR3MUSXkP8CYdqLkWZFsI0PVF6NYtp5pBTQIDAQABAoIBAHrDrTrNpkk0dQxYj4CGl7wjx6Br11AHMjMqpqNv1SmogZtXpelHSQVvDs6BaKQzJRW8igEaQ6bEweI5FcrRszXoPxOdbRsVwctucesZkf57PDWZrwvLW6i7JAXmWxHXrWkXyD3e9k3yWJYgVDs9WU9Kv+7sgn9RG7R+QcUa0yPVc90SuEM/zx1ToYGOP/vIR3pVY0Uoc4qV1kzgJQBoUlApfISL6MjDXNDzI5B4d5jr1eVNbY4GqnyR49N1tgF4ctQXK3oaBScPRhyv/+ErNT6b41TyK2TTXBCLt34bytDKqsy1Ur/tMe9a96uqSVHRAco57b6w0hYWJoRPV4ET9gECgYEAwMLH4xeCuB0khODE2lmOYzBCCTiwztnImVTW7HgKPvWASGZUE/uvGAfZrEHwJ44DoOjgNH27Oj/fqe3Hx6tUsANRwLqvvHBrWKufQKOtLqGO3z1IBXkVamWT2k2I15fjEUnuBpfdsTNPGE4qVyNwWk3uJM7+lBQPbDjak/hSKC0CgYEA8559v+lIutPzHRf0hg1IKsebyqdIbn5G+P9SDQuuhw7PXJBxWKblRgz11pumAR/zStgACqJgqBand3WRQYYbfpb7Ei+AhJDi1ZwddFTk2mA9aKlVUNuyyTx9r3BN9sWYqDegn+tbf/yr6pKhpQwnU0Um1nNaWalm1s0n3grUEaECgYAhqmMqwFJuQXi9VFxNHlMF88m0vpfyqIqmbPDUf+qaMFplSqnoi457DfPwZ9u/rMfpdIKj6Emo1LsFfKflsYCq9Ql0Naa3rJKy+9Zmfa+jc0f2qUdI3WrmGDOIbv41WSupO1Y9BI0Ng76Oqigu69uVigLLnvNLfW1sI0nZigcfSQKBgBBf5cnhbz8Hgf7BnnDoMaKWehU7+zVaDYEtACHaWCfByhRJrSStSxnTQy7ilVzb/elY7V/JnD+QDj+MSnAiCHUQxt1pDfVbG7QJ4zzve9Zlw5rmTtK5gaHfC/+fx82/aExeONCm7CaFIDULGAxU7cu+CSc+56LBLSVg8r4M8kYhAoGBAKLh5gnV1hHXqcYrFiZ5iYV9/pIQqLvB8UsGUIcFmUoxggBJqzQQA1h0Aq9qn+M+ARSMEI4XUfR8jBuS6MlZhzs4rytTrKtugfHZK4GFB1nc3OJTyBjSYAfdykkxQKkqgxhDCeAoxmjVDTP8KDh3ghrIAAYj3H1sKD4+WfEvrDpr"
`
		gateway := `
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: Gateway
metadata:
  name: kuma
  namespace: kuma-test
spec:
  gatewayClassName: kuma
  listeners:
  - name: proxy
    port: 8090
    hostname: 'test-server-1.com'
    protocol: HTTPS
    tls:
      certificateRefs:
      - name: secret-tls`

		var address string

		BeforeEach(func() {
			err := k8s.RunKubectlE(cluster.GetTesting(), cluster.GetKubectlOptions(), "delete", "gateway", "--all")
			Expect(err).ToNot(HaveOccurred())
			Expect(YamlK8s(secret)(cluster)).To(Succeed())
			Expect(YamlK8s(gateway)(cluster)).To(Succeed())
			address = net.JoinHostPort(GatewayIP(), "8090")
		})

		It("should route the traffic using TLS", func() {
			// given
			route := `
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: HTTPRoute
metadata:
  name: test-server-paths
  namespace: kuma-test
spec:
  parentRefs:
  - name: kuma
  rules:
  - backendRefs:
    - name: test-server-1
      port: 80
    matches:
    - path:
        type: PathPrefix
        value: /`

			// when
			err := YamlK8s(route)(cluster)

			// then
			Expect(err).ToNot(HaveOccurred())

			Eventually(func(g Gomega) {
				resp, err := client.CollectResponseDirectly("https://"+address, client.WithHeader("host", "test-server-1.com"))
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(resp.Instance).To(Equal("test-server-1"))
			}, "30s", "1s").Should(Succeed())
		})

		It("should manage Kuma Secret", func() {
			// given converted Kuma secret
			const convertedSecretName = "gapi-kuma-test-secret-tls"
			var kumaSecret string
			Eventually(func(g Gomega) {
				out, err := cluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "secret", convertedSecretName, "-o", "json")
				g.Expect(err).ToNot(HaveOccurred())
				kumaSecret = out
			}, "30s", "1s").Should(Succeed())

			// when original secret is changed
			secret = `
apiVersion: v1
kind: Secret
metadata:
  name: secret-tls
  namespace: kuma-test
type: kubernetes.io/tls
data:
  tls.crt: "MIIDKDCCAhCgAwIBAgIQeLo41q+2LqjMZ3nnvR1gvjANBgkqhkiG9w0BAQsFADAXMRUwEwYDVQQDEwx0ZXN0Lmt1bWEuaW8wHhcNMjIwMzI0MTM1MDM1WhcNMzIwMzIxMTM1MDM1WjAXMRUwEwYDVQQDEwx0ZXN0Lmt1bWEuaW8wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQCmn5iSOWHzcV+XKqCVxA4hiHAwe+nvMZLqng0ZFqcQb6EYMl7ms3ANUiyQQSTEVrZGXMrCef47EXufu6UExk26G++aSr1i+6edND/02fIP/RpbfC1yjvxcDbzya8IIH4nCPNK4IOVjAfkUEZfJMZTdNEFy1CtKshwhQcpOwwzznn6hQ3o8SwwHXeHle4aBvs/NyiOQSlSWO5qqK3ttGIlZFc/lbIYU47djwkLTPMTZWMoAv6oHHUw7oJDCGiFhjsB07PTLi/nu5+zHJhmcYBSH/BRJk0zW19KKOhuXeCiKD+IWE9dCun6VJjTopzZYMg7Abv02gdE+cbEBz4Yq6rGHAgMBAAGjcDBuMA4GA1UdDwEB/wQEAwICpDATBgNVHSUEDDAKBggrBgEFBQcDATAPBgNVHRMBAf8EBTADAQH/MB0GA1UdDgQWBBR3oYhD31+G5MADuYTGkWo9Ws2I4DAXBgNVHREEEDAOggx0ZXN0Lmt1bWEuaW8wDQYJKoZIhvcNAQELBQADggEBACaMFYMdyDKtZXK7XIERvrjqVT8UyjMZyDqpW3TyokIYb0cf8snhEGAFezXb6fTNwxM8DiKEXFg+DLJsIXL3q62+pV67NN25GG5zIFJ+xloXSSVmtLXHoQZIHcKiBgzuVhZ7wsLvj5R62NYxwdiKjbgERnTejWiC4eEBl3HGJNu55zF25pTYuvhphpJvfaHsuhvnwHXMVmxCXa+dQws5wOwjjSHbAL8DwFOiwBAMs+4B6AvP70X7CduDWWKfS7fTMU7yo4kNORFD79MpsXCvAYPzrpb68p7Iv9WA9RSGKIxkv1nhSgn8mUfohpPUU5/dn37qBqYMKSlonVVVaCT9e4c="
  tls.key: "MIIEpAIBAAKCAQEApp+Ykjlh83FflyqglcQOIYhwMHvp7zGS6p4NGRanEG+hGDJe5rNwDVIskEEkxFa2RlzKwnn+OxF7n7ulBMZNuhvvmkq9YvunnTQ/9NnyD/0aW3wtco78XA288mvCCB+JwjzSuCDlYwH5FBGXyTGU3TRBctQrSrIcIUHKTsMM855+oUN6PEsMB13h5XuGgb7PzcojkEpUljuaqit7bRiJWRXP5WyGFOO3Y8JC0zzE2VjKAL+qBx1MO6CQwhohYY7AdOz0y4v57ufsxyYZnGAUh/wUSZNM1tfSijobl3goig/iFhPXQrp+lSY06Kc2WDIOwG79NoHRPnGxAc+GKuqxhwIDAQABAoIBAGMlsm7IMG3gt4XG+rlDaTkw67kd8uy+7fInzBlyrkSMeCpixq+2dGWo0RWhfdRK2Llzu78PQmU5mPtKd/4oVdX5i5CfCqScpHdZwPjcuzoiXI21jYGObcIE9rq1vkaBJcLr2GNyR6yrXuA/O7efhjo+PveqInyXEUAE2vIVBF7tqhaZHjP7KWYkE5r4eHc+3CRmPyUaRL8luzQ70sXKdBT2fyZtSsCZus7hKwtsw+uztV9B3tV04WpN48J2C3B1hjscSlVy/T4z9gVABVSmZ0CgUF6Ve6Ams6bCiW1GvV9EnNpRRJWoPP44ezWrSVXRLEq8dVkrIiHII+AmyEpObdECgYEAzanD4E96Yoq5o+4P9XlDSu6ewrDd9hszfafSqz/plzImsB//xAK58TN4OkkGGOD/6p0KGq3uOW7fysP+d6AmfK5xYpCoRp8esX5CRJlJKZAcOXXqFhqpIXPGPp4ECKXDbBN+gA6RfosW7Za27oHGt3jb+pIqZx4B0CCH/8T4wz8CgYEAz2e4E/DWOEPeIxIrozSxZ8Tm4uOaLYnpF96LGsd7pcoy2bUno1513/8ljei9FG+q+3QSvMX3xiZ2g59Djo/EFgCxuEEDGdl9kQKzbxlEOjdMRp/45MWUrJneRIaZIADa1VktInvfgvyw5ykkkwIaejXgsgE4jxYEZ6mgfFPZJ7kCgYEAjC+r1LpYMdNdtuAPAMQnmmwMy6jDo33nGz/J2fE2yFjnBibJsFIrbL4otZIRFeRIjSN+P7FPMNbitNPkIKaJlXNS1lzEV18fDN2DTj8uH6ablo9JgMergaHo/8W1+i8DhifDkoZbpmYousqA5xKO4YEAcUwwmxlZIwUJrs4UwwECgYApZcFnMYYAmwNGqsTNAaJ7XODc3qMSfdlpq0DGqpRyhgZaT/9Ga39Jo2rChXbgEl3lbMikpzsKcPjs0qgwV1/DKIZQiaFt0mxuukRIY5mrqQfvfNS4DPvc66Vith7wNVt4hCEvJdx6D6fbq+mh8iSIyiI8RWQdoz5j1oay6iWI+QKBgQCveTt3CXZnOaNUzG761sowXXxdjB0YoPLcMrF0VrhkMfVkbty535l8Wis9Wd5XUImKWmcXLsRVFs4jV/YX+KpszNgRuefTJ2p8V/RB0Fwh7nWveC3n8vNgN6AVwEAs7DEXqdVXzNq6l3dyNxX1L3MVPeVbPTEfZUNvKj6ed22tfA=="
`
			err := YamlK8s(secret)(cluster)

			// then copied secret is also changed
			Expect(err).ToNot(HaveOccurred())
			Eventually(func(g Gomega) {
				out, err := cluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "secret", convertedSecretName, "-o", "json")
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(out).ToNot(MatchJSON(kumaSecret))
			}, "30s", "1s").Should(Succeed())

			// when original secret is removed
			err = k8s.RunKubectlE(cluster.GetTesting(), cluster.GetKubectlOptions(TestNamespace), "delete", "secret", "secret-tls")

			// then copied secret is removed
			Expect(err).ToNot(HaveOccurred())
			Eventually(func(g Gomega) {
				out, err := cluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "secrets")
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(out).ToNot(ContainSubstring(convertedSecretName))
			}, "30s", "1s").Should(Succeed())
		})
	})
}
