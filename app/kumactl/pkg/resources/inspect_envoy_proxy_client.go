package resources

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	util_http "github.com/kumahq/kuma/pkg/util/http"
)

type InspectEnvoyProxyClient interface {
	ConfigDump(ctx context.Context, rk core_model.ResourceKey) ([]byte, error)
	Stats(ctx context.Context, rk core_model.ResourceKey) ([]byte, error)
	Clusters(ctx context.Context, rk core_model.ResourceKey) ([]byte, error)
}

func NewInspectEnvoyProxyClient(resDesc core_model.ResourceTypeDescriptor, client util_http.Client) InspectEnvoyProxyClient {
	return &httpInspectEnvoyProxyClient{
		resDesc: resDesc,
		client:  client,
	}
}

type httpInspectEnvoyProxyClient struct {
	client  util_http.Client
	resDesc core_model.ResourceTypeDescriptor
}

var _ InspectEnvoyProxyClient = &httpInspectEnvoyProxyClient{}

func (h *httpInspectEnvoyProxyClient) ConfigDump(ctx context.Context, rk core_model.ResourceKey) ([]byte, error) {
	return h.executeInspectRequest(ctx, rk, "xds")
}

func (h *httpInspectEnvoyProxyClient) Stats(ctx context.Context, rk core_model.ResourceKey) ([]byte, error) {
	return h.executeInspectRequest(ctx, rk, "stats")
}

func (h *httpInspectEnvoyProxyClient) Clusters(ctx context.Context, rk core_model.ResourceKey) ([]byte, error) {
	return h.executeInspectRequest(ctx, rk, "clusters")
}

func (h *httpInspectEnvoyProxyClient) executeInspectRequest(ctx context.Context, rk core_model.ResourceKey, inspectionPath string) ([]byte, error) {
	resUrl, err := h.buildURL(rk, inspectionPath)
	if err != nil {
		return nil, errors.Wrap(err, "could not construct the url")
	}
	req, err := http.NewRequest("GET", resUrl.String(), nil)
	if err != nil {
		return nil, err
	}
	statusCode, b, err := doRequest(h.client, ctx, req)
	if err != nil {
		return nil, err
	}
	if statusCode != 200 {
		return nil, errors.Errorf("(%d): %s", statusCode, string(b))
	}
	return b, nil
}

func (h *httpInspectEnvoyProxyClient) buildURL(rk core_model.ResourceKey, inspectionPath string) (*url.URL, error) {
	var prefix string
	if h.resDesc.Scope == core_model.ScopeMesh {
		prefix = fmt.Sprintf("/meshes/%s", rk.Mesh)
	}
	plural := h.resDesc.WsPath
	// todo(lobkovilya): rename mesh.ZoneIngressResourceTypeDescriptor.WsPath to "zoneingresses" and use it here
	if h.resDesc.Name == mesh.ZoneIngressType {
		plural = "zoneingresses"
	}
	return url.Parse(fmt.Sprintf("%s/%s/%s/%s", prefix, plural, rk.Name, inspectionPath))
}
