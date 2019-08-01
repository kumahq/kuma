package remote

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model/rest"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"

	"github.com/pkg/errors"
)

var _ model.ResourceMeta = &remoteMeta{}

type remoteMeta struct {
	Namespace string
	Name      string
	Mesh      string
	Version   string
}

func (m remoteMeta) GetName() string {
	return m.Name
}
func (m remoteMeta) GetNamespace() string {
	return m.Namespace
}
func (m remoteMeta) GetMesh() string {
	return m.Mesh
}
func (m remoteMeta) GetVersion() string {
	return m.Version
}

func NewStore(client http.Client, api rest.Api) store.ResourceStore {
	return &remoteStore{
		client: client,
		api:    api,
	}
}

var _ store.ResourceStore = &remoteStore{}

type remoteStore struct {
	client http.Client
	api    rest.Api
}

func (s *remoteStore) Create(context.Context, model.Resource, ...store.CreateOptionsFunc) error {
	return errors.Errorf("not implemented yet")
}
func (s *remoteStore) Update(context.Context, model.Resource, ...store.UpdateOptionsFunc) error {
	return errors.Errorf("not implemented yet")
}
func (s *remoteStore) Delete(context.Context, model.Resource, ...store.DeleteOptionsFunc) error {
	return errors.Errorf("not implemented yet")
}
func (s *remoteStore) Get(context.Context, model.Resource, ...store.GetOptionsFunc) error {
	return errors.Errorf("not implemented yet")
}
func (s *remoteStore) List(ctx context.Context, rs model.ResourceList, fs ...store.ListOptionsFunc) error {
	resourceApi, err := s.api.GetResourceApi(rs.GetItemType())
	if err != nil {
		return errors.Wrapf(err, "failed to construct URI to fetch a list of %q", rs.GetItemType())
	}
	opts := store.NewListOptions(fs...)
	req, err := http.NewRequest("GET", fmt.Sprintf("/meshes/%s/%s", opts.Mesh, resourceApi.CollectionPath), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := s.client.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	rsr := &rest.ResourceListReceiver{
		ResourceRegistry: &model.SimpleResourceRegistry{
			ResourceTypes: map[model.ResourceType]model.Resource{
				(rs.GetItemType()): rs.NewItem(),
			},
		},
	}
	if err := json.Unmarshal(b, rsr); err != nil {
		return err
	}
	for _, ri := range rsr.ResourceList.Items {
		r := rs.NewItem()
		r.SetSpec(ri.Spec)
		r.SetMeta(&remoteMeta{
			Namespace: "",
			Name:      ri.Meta.Name,
			Mesh:      ri.Meta.Mesh,
			Version:   "",
		})
		_ = rs.AddItem(r)
	}
	return nil
}
