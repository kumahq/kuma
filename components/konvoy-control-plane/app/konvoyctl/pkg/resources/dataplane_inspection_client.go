package resources

import (
	"context"
	"fmt"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/mesh"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/remote"
	konvoy_http "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/util/http"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/url"
)

type DataplaneInspectionClient interface {
	List(ctx context.Context, meshName string, tags map[string]string) (*mesh.DataplaneInspectionResourceList, error)
}

func NewDataplaneInspectionClient(apiServerUrl string) (DataplaneInspectionClient, error) {
	client, err := apiServerClient(apiServerUrl)
	if err != nil {
		return nil, err
	}
	return &httpDataplaneInspectionClient{
		Client: client,
	}, nil
}

type httpDataplaneInspectionClient struct {
	Client konvoy_http.Client
}

func (d *httpDataplaneInspectionClient) List(ctx context.Context, meshName string, tags map[string]string) (*mesh.DataplaneInspectionResourceList, error) {
	resUrl, err := constructUrl(meshName, tags)
	if err != nil {
		return nil, errors.Wrap(err, "could not construct the url")
	}
	req, err := http.NewRequest("GET", resUrl.String(), nil)
	if err != nil {
		return nil, err
	}
	statusCode, b, err := d.doRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK {
		return nil, errors.Errorf("unexpected status code: %d", statusCode)
	}
	inspections := mesh.DataplaneInspectionResourceList{}
	if err := remote.UnmarshalList(b, &inspections); err != nil {
		return nil, err
	}
	return &inspections, nil
}

func constructUrl(meshName string, tags map[string]string) (*url.URL, error) {
	result, err := url.Parse(fmt.Sprintf("/meshes/%s/dataplane-inspections", meshName))
	if err != nil {
		return nil, err
	}
	query := result.Query()
	for tag, value := range tags {
		query.Add("tag", fmt.Sprintf("%s:%s", tag, value))
	}
	result.RawQuery = query.Encode()
	return result, err
}

func (d *httpDataplaneInspectionClient) doRequest(ctx context.Context, req *http.Request) (int, []byte, error) {
	resp, err := d.Client.Do(req.WithContext(ctx))
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, err
	}
	return resp.StatusCode, b, nil
}
