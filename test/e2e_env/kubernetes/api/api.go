package api

import (
	"encoding/json"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func Api() {
	It("Works with /policies", func() {
		r, err := http.Get(kubernetes.Cluster.GetKuma().GetAPIServerAddress() + "/policies")
		Expect(err).ToNot(HaveOccurred())
		defer r.Body.Close()

		res := map[string][]interface{}{}
		Expect(json.NewDecoder(r.Body).Decode(&res)).To(Succeed())
		Expect(len(res["policies"])).To(BeNumerically(">", 2))
	})

	It("Works with /", func() {
		r, err := http.Get(kubernetes.Cluster.GetKuma().GetAPIServerAddress())
		Expect(err).ToNot(HaveOccurred())
		defer r.Body.Close()

		res := map[string]interface{}{}
		Expect(json.NewDecoder(r.Body).Decode(&res)).To(Succeed())
		Expect(res["version"]).ToNot(BeEmpty())
	})

	It("Get k8s version of default mesh", func() {
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

	It("Get kubernetes version of default mesh", func() {
		r, err := http.Get(kubernetes.Cluster.GetKuma().GetAPIServerAddress() + "/meshes/default")
		Expect(err).ToNot(HaveOccurred())
		defer r.Body.Close()

		res := map[string]interface{}{}
		Expect(json.NewDecoder(r.Body).Decode(&res)).To(Succeed())
		Expect(res).To(HaveKey("type"))
		Expect(res["type"]).To(Equal("Mesh"))
	})
}
