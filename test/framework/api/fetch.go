package api

import (
	"fmt"
	"io"
	"net/http"

	"github.com/onsi/gomega"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/test/framework"
)

func FetchResource(g gomega.Gomega, cluster framework.Cluster, out core_model.Resource, mesh string, name string) {
	desc := out.Descriptor()
	path := ""
	switch desc.Scope {
	case core_model.ScopeMesh:
		path += fmt.Sprintf("/meshes/%s/%s/%s", desc.WsPath, mesh, name)
	case core_model.ScopeGlobal:
		path += fmt.Sprintf("/%s/%s", desc.WsPath, name)
	}
	r, err := http.Get(cluster.GetKuma().GetAPIServerAddress() + path)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	defer r.Body.Close()
	g.Expect(r).To(gomega.HaveHTTPStatus(200))

	body, err := io.ReadAll(r.Body)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	restRes, err := rest.JSON.Unmarshal(body, desc)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	_ = out.SetSpec(restRes.GetSpec())
	out.SetMeta(restRes.GetMeta())
}
