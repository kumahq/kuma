package api

import (
	"encoding/json"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/universal/env"
)

func Api() {
	It("Works with /policies", func() {
		r, err := http.Get(env.Cluster.GetKuma().GetAPIServerAddress() + "/policies")
		Expect(err).ToNot(HaveOccurred())
		defer r.Body.Close()

		res := map[string][]interface{}{}
		Expect(json.NewDecoder(r.Body).Decode(&res)).To(Succeed())
		Expect(len(res["policies"])).To(BeNumerically(">", 2))
	})

	It("Works with /", func() {
		r, err := http.Get(env.Cluster.GetKuma().GetAPIServerAddress())
		Expect(err).ToNot(HaveOccurred())
		defer r.Body.Close()

		res := map[string]interface{}{}
		Expect(json.NewDecoder(r.Body).Decode(&res)).To(Succeed())
		Expect(res["version"]).ToNot(BeEmpty())
	})
}
