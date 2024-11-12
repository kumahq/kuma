package remote

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"maps"
	"net/http"
	"strconv"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	rest_v1alpha1 "github.com/kumahq/kuma/pkg/core/resources/model/rest/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/rest/errors/types"
	util_http "github.com/kumahq/kuma/pkg/util/http"
)

var log = core.Log.WithName("store-remote")

func NewStore(client util_http.Client, api rest.Api) store.ResourceStore {
	return &remoteStore{
		client: client,
		api:    api,
	}
}

var _ store.ResourceStore = &remoteStore{}

type remoteStore struct {
	client util_http.Client
	api    rest.Api
}

func (s *remoteStore) Create(ctx context.Context, res model.Resource, fs ...store.CreateOptionsFunc) error {
	opts := store.NewCreateOptions(fs...)
	meta := rest_v1alpha1.ResourceMeta{
<<<<<<< HEAD
		Type: string(res.Descriptor().Name),
		Name: opts.Name,
		Mesh: opts.Mesh,
=======
		Type:   string(res.Descriptor().Name),
		Name:   opts.Name,
		Mesh:   opts.Mesh,
		Labels: maps.Clone(opts.Labels),
>>>>>>> c3d7187c7 (fix(kuma-cp): avoid concurrent access on resource meta (#11997))
	}
	if err := s.upsert(ctx, res, meta); err != nil {
		return err
	}
	return nil
}

func (s *remoteStore) Update(ctx context.Context, res model.Resource, fs ...store.UpdateOptionsFunc) error {
	meta := rest_v1alpha1.ResourceMeta{
<<<<<<< HEAD
		Type: string(res.Descriptor().Name),
		Name: res.GetMeta().GetName(),
		Mesh: res.GetMeta().GetMesh(),
=======
		Type:   string(res.Descriptor().Name),
		Name:   res.GetMeta().GetName(),
		Mesh:   res.GetMeta().GetMesh(),
		Labels: maps.Clone(opts.Labels),
>>>>>>> c3d7187c7 (fix(kuma-cp): avoid concurrent access on resource meta (#11997))
	}
	if err := s.upsert(ctx, res, meta); err != nil {
		return err
	}
	return nil
}

func (s *remoteStore) upsert(ctx context.Context, res model.Resource, meta rest_v1alpha1.ResourceMeta) error {
	resourceApi, err := s.api.GetResourceApi(res.Descriptor().Name)
	if err != nil {
		return errors.Wrapf(err, "failed to construct URI to update a %q", res.Descriptor().Name)
	}
	resCopy := res.Descriptor().NewObject()
	resCopy.SetMeta(meta)
	if err := resCopy.SetSpec(res.GetSpec()); err != nil {
		return err
	}
	restRes := rest.From.Resource(resCopy)
	b, err := json.Marshal(restRes)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("PUT", resourceApi.Item(meta.Mesh, meta.Name), bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("content-type", "application/json")
	statusCode, b, err := s.doRequest(ctx, req)
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK && statusCode != http.StatusCreated {
		if statusCode == http.StatusMethodNotAllowed {
			return errors.Errorf("%s", string(b))
		} else {
			return errors.Errorf("(%d): %s", statusCode, string(b))
		}
	}
	res.SetMeta(meta)
	return nil
}

func (s *remoteStore) Delete(ctx context.Context, res model.Resource, fs ...store.DeleteOptionsFunc) error {
	opts := store.NewDeleteOptions(fs...)
	resourceApi, err := s.api.GetResourceApi(res.Descriptor().Name)
	if err != nil {
		return errors.Wrapf(err, "failed to construct URI to delete a %q", res.Descriptor().Name)
	}
	req, err := http.NewRequest("DELETE", resourceApi.Item(opts.Mesh, opts.Name), nil)
	if err != nil {
		return err
	}
	statusCode, b, err := s.doRequest(ctx, req)
	if err != nil {
		if statusCode == 404 {
			return store.ErrorResourceNotFound(res.Descriptor().Name, opts.Name, opts.Mesh)
		}
		return err
	}
	if statusCode != http.StatusOK {
		if statusCode == http.StatusMethodNotAllowed {
			return errors.Errorf("%s", string(b))
		} else {
			return errors.Errorf("(%d): %s", statusCode, string(b))
		}
	}
	return nil
}

func (s *remoteStore) Get(ctx context.Context, res model.Resource, fs ...store.GetOptionsFunc) error {
	resourceApi, err := s.api.GetResourceApi(res.Descriptor().Name)
	if err != nil {
		return errors.Wrapf(err, "failed to construct URI to fetch a %q", res.Descriptor().Name)
	}
	opts := store.NewGetOptions(fs...)
	req, err := http.NewRequest("GET", resourceApi.Item(opts.Mesh, opts.Name), nil)
	if err != nil {
		return err
	}
	statusCode, b, err := s.doRequest(ctx, req)
	if statusCode == 404 {
		return store.ErrorResourceNotFound(res.Descriptor().Name, opts.Name, opts.Mesh)
	}
	if err != nil {
		return err
	}
	if statusCode != 200 {
		return errors.Errorf("(%d): %s", statusCode, string(b))
	}
	restRes, err := rest.JSON.Unmarshal(b, res.Descriptor())
	if err != nil {
		return err
	}
	res.SetMeta(restRes.GetMeta())
	return res.SetSpec(restRes.GetSpec())
}

func (s *remoteStore) List(ctx context.Context, rs model.ResourceList, fs ...store.ListOptionsFunc) error {
	resourceApi, err := s.api.GetResourceApi(rs.GetItemType())
	if err != nil {
		return errors.Wrapf(err, "failed to construct URI to fetch a list of %q", rs.GetItemType())
	}
	opts := store.NewListOptions(fs...)
	req, err := http.NewRequest("GET", resourceApi.List(opts.Mesh), nil)
	if err != nil {
		return err
	}
	query := req.URL.Query()
	if opts.PageOffset != "" {
		query.Add("offset", opts.PageOffset)
	}
	if opts.PageSize != 0 {
		query.Add("size", strconv.Itoa(opts.PageSize))
	}
	req.URL.RawQuery = query.Encode()

	log.V(1).Info("doing request to control-plane", "method", req.Method, "url", req.URL.String())
	statusCode, b, err := s.doRequest(ctx, req)
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK {
		return errors.Errorf("(%d): %s", statusCode, string(b))
	}
	return rest.JSON.UnmarshalListToCore(b, rs)
}

// execute a request. Returns status code, body, error
func (s *remoteStore) doRequest(ctx context.Context, req *http.Request) (int, []byte, error) {
	req.Header.Set("Accept", "application/json")
	resp, err := s.client.Do(req.WithContext(ctx))
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, err
	}
	if resp.StatusCode/100 >= 4 {
		kumaErr := types.Error{}
		if err := json.Unmarshal(b, &kumaErr); err == nil {
			if kumaErr.Title != "" {
				return resp.StatusCode, b, &kumaErr
			}
		}
	}
	return resp.StatusCode, b, nil
}
