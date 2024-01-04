package api

import (
	"encoding/json"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func Api() {
	It("works with /policies", func() {
		r, err := http.Get(kubernetes.Cluster.GetKuma().GetAPIServerAddress() + "/policies")
		Expect(err).ToNot(HaveOccurred())
		defer r.Body.Close()

		res := map[string][]interface{}{}
		Expect(json.NewDecoder(r.Body).Decode(&res)).To(Succeed())
		Expect(len(res["policies"])).To(BeNumerically(">", 2))
	})

	It("works with /", func() {
		r, err := http.Get(kubernetes.Cluster.GetKuma().GetAPIServerAddress())
		Expect(err).ToNot(HaveOccurred())
		defer r.Body.Close()

		res := map[string]interface{}{}
		Expect(json.NewDecoder(r.Body).Decode(&res)).To(Succeed())
		Expect(res["version"]).ToNot(BeEmpty())
	})

	It("get k8s version of default mesh", func() {
		r, err := http.Get(kubernetes.Cluster.GetKuma().GetAPIServerAddress() + "/meshes/default?format=k8s")
		Expect(err).ToNot(HaveOccurred())
		defer r.Body.Close()

		res := map[string]interface{}{}
		Expect(json.NewDecoder(r.Body).Decode(&res)).To(Succeed())
		Expect(res).To(HaveKey("kind"))
		Expect(res["kind"]).To(Equal("Mesh"))
		Expect(res).To(HaveKey("apiVersion"))
		Expect(res["apiVersion"]).To(Equal("kuma.io/v1alpha1"))
	})

	It("get kubernetes version of default mesh", func() {
		r, err := http.Get(kubernetes.Cluster.GetKuma().GetAPIServerAddress() + "/meshes/default")
		Expect(err).ToNot(HaveOccurred())
		defer r.Body.Close()

		res := map[string]interface{}{}
		Expect(json.NewDecoder(r.Body).Decode(&res)).To(Succeed())
		Expect(res).To(HaveKey("type"))
		Expect(res["type"]).To(Equal("Mesh"))
	})

	type testCase struct {
		path       string
		k8sSecType string
	}
	DescribeTable("gets secret",
		func(given testCase) {
			token, err := kubernetes.Cluster.GetKumactlOptions().RunKumactlAndGetOutput("generate", "user-token",
				"--name", "mesh-system:admin",
				"--group", "mesh-system:admin",
				"--valid-for", "24h",
			)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func(g Gomega) {
				req, err := http.NewRequest("GET", kubernetes.Cluster.GetKuma().GetAPIServerAddress()+given.path, nil)
				g.Expect(err).ToNot(HaveOccurred())
				req.Header.Add("authorization", "Bearer "+token)
				r, err := http.DefaultClient.Do(req)
				g.Expect(err).ToNot(HaveOccurred())
				defer r.Body.Close()

				res := map[string]interface{}{}
				g.Expect(json.NewDecoder(r.Body).Decode(&res)).To(Succeed())
				g.Expect(res).To(HaveKey("kind"))
				g.Expect(res["kind"]).To(Equal("Secret"))
				g.Expect(res).To(HaveKey("apiVersion"))
				g.Expect(res["apiVersion"]).To(Equal("meta.k8s.io/v1"))
				g.Expect(res["type"]).To(Equal(given.k8sSecType))
			}, "30s", "1s").Should(Succeed())
		},
		Entry("global secret", testCase{
			path:       "/global-secrets/inter-cp-ca?format=k8s",
			k8sSecType: "system.kuma.io/global-secret",
		}),
		Entry("mesh secret", testCase{
			path:       "/meshes/default/secrets/dataplane-token-signing-key-default-1?format=k8s",
			k8sSecType: "system.kuma.io/secret",
		}),
	)
}
