package api

import (
	"fmt"
	"io"
	"net/http"

	"github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/kri"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/test/framework"
)

// fetchResourceFromPath performs the HTTP GET and unmarshalling for a given API path.
func fetchResourceFromPath(g gomega.Gomega, cluster framework.Cluster, out core_model.Resource, path string) int {
	r, err := http.Get(cluster.GetKuma().GetAPIServerAddress() + path)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	defer func() { _ = r.Body.Close() }()
	g.Expect(r).To(gomega.HaveHTTPStatus(200))

	body, err := io.ReadAll(r.Body)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	restRes, err := rest.JSON.Unmarshal(body, out.Descriptor())
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(out.SetSpec(restRes.GetSpec())).ToNot(gomega.HaveOccurred())
	out.SetMeta(restRes.GetMeta())

	return r.StatusCode
}

func FetchResource(g gomega.Gomega, cluster framework.Cluster, out core_model.Resource, mesh string, name string) {
	desc := out.Descriptor()
	path := ""
	switch desc.Scope {
	case core_model.ScopeMesh:
		path += fmt.Sprintf("/meshes/%s/%s/%s", desc.WsPath, mesh, name)
	case core_model.ScopeGlobal:
		path += fmt.Sprintf("/%s/%s", desc.WsPath, name)
	}
	fetchResourceFromPath(g, cluster, out, path)
}

func FetchResourceByKri(g gomega.Gomega, cluster framework.Cluster, out core_model.Resource, kri kri.Identifier) int {
	path := "/_kri/" + kri.String()
	return fetchResourceFromPath(g, cluster, out, path)
}
