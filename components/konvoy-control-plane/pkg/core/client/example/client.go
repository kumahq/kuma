package example

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/client"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/model"
)

type ExampleStorageRecord struct {
	ResourceType string
	Namespace    string
	Name         string
	Spec         string
}
type ExampleStorageRecords = []*ExampleStorageRecord

var _ model.ResourceMeta = &ExampleMeta{}

type ExampleMeta struct {
	Namespace string
	Name      string
	Version   string
}

func (m *ExampleMeta) GetName() string {
	return m.Name
}
func (m *ExampleMeta) GetNamespace() string {
	return m.Namespace
}
func (m *ExampleMeta) GetVersion() string {
	return m.Version
}

var _ client.ResourceClient = &ExampleResourceClient{}

type ExampleResourceClient struct {
	PersistedRecords ExampleStorageRecords
}

func (c *ExampleResourceClient) Create(_ context.Context, r model.Resource, fs ...client.CreateOptionsFunc) error {
	opts := client.NewCreateOptions(fs...)
	// Namespace and Name must be provided via CreateOptions
	if _, record := c.findPersistedRecord(string(r.GetType()), opts.Namespace, opts.Name); record != nil {
		return client.ErrorResourceAlreadyExists(r.GetType(), opts.Namespace, opts.Name)
	}
	// convert spec into storage representation
	spec, err := json.Marshal(r.GetSpec())
	if err != nil {
		return err
	}
	// persist
	c.PersistedRecords = append(c.PersistedRecords,
		&ExampleStorageRecord{
			ResourceType: string(r.GetType()),
			// Namespace and Name must be provided via CreateOptions
			Namespace: opts.Namespace,
			Name:      opts.Name,
			Spec:      string(spec),
		})
	return nil
}
func (c *ExampleResourceClient) Update(_ context.Context, r model.Resource, fs ...client.UpdateOptionsFunc) error {
	_ = client.NewUpdateOptions(fs...)
	// type cast r.GetMeta()
	meta, ok := (r.GetMeta()).(*ExampleMeta)
	if !ok {
		return fmt.Errorf("ExampleResourceClient.Update() requires r.GetMeta() to be of type ExampleMeta")
	}
	// Namespace and Name must be provided via r.GetMeta()
	idx, record := c.findPersistedRecord(string(r.GetType()), meta.Namespace, meta.Name)
	if record == nil {
		return client.ErrorResourceNotFound(r.GetType(), meta.Namespace, meta.Name)
	}
	// convert spec into storage representation
	spec, err := json.Marshal(r.GetSpec())
	if err != nil {
		return err
	}
	// persist
	c.PersistedRecords[idx] = &ExampleStorageRecord{
		ResourceType: string(r.GetType()),
		// Namespace and Name must be provided via CreateOptions
		Namespace: meta.Namespace,
		Name:      meta.Name,
		Spec:      string(spec),
	}
	return nil
}
func (c *ExampleResourceClient) Delete(_ context.Context, r model.Resource, fs ...client.DeleteOptionsFunc) error {
	opts := client.NewDeleteOptions(fs...)
	// type cast r.GetMeta()
	_, ok := (r.GetMeta()).(*ExampleMeta)
	if r.GetMeta() != nil && !ok {
		return fmt.Errorf("ExampleResourceClient.Delete() requires r.GetMeta() either to be nil or to be of type ExampleMeta")
	}
	// Namespace and Name must be provided via DeleteOptions
	idx, record := c.findPersistedRecord(string(r.GetType()), opts.Namespace, opts.Name)
	if record != nil {
		c.PersistedRecords = append(c.PersistedRecords[:idx], c.PersistedRecords[idx+1:]...)
	}
	return nil
}

func (c *ExampleResourceClient) Get(_ context.Context, r model.Resource, fs ...client.GetOptionsFunc) error {
	opts := client.NewGetOptions(fs...)
	// Namespace and Name must be provided via GetOptions
	_, record := c.findPersistedRecord(string(r.GetType()), opts.Namespace, opts.Name)
	if record == nil {
		return client.ErrorResourceNotFound(r.GetType(), opts.Namespace, opts.Name)
	}
	return c.convertPersistedRecord(record, r)
}
func (c *ExampleResourceClient) List(_ context.Context, rs model.ResourceList, fs ...client.ListOptionsFunc) error {
	opts := client.NewListOptions(fs...)
	// Namespace must be provided via ListOptions
	records := c.findPersistedRecords(string(rs.GetItemType()), opts.Namespace)
	for _, record := range records {
		r := rs.NewItem()
		c.convertPersistedRecord(record, r)
		rs.AddItem(r)
	}
	return nil
}

func (c *ExampleResourceClient) findPersistedRecord(
	resourceType string, namespace string, name string) (int, *ExampleStorageRecord) {
	for idx, rec := range c.PersistedRecords {
		if rec.ResourceType == resourceType &&
			rec.Namespace == namespace &&
			rec.Name == name {
			return idx, rec
		}
	}
	return -1, nil
}

func (c *ExampleResourceClient) findPersistedRecords(
	resourceType string, namespace string) []*ExampleStorageRecord {
	res := make([]*ExampleStorageRecord, 0)
	for _, rec := range c.PersistedRecords {
		if rec.ResourceType == resourceType &&
			rec.Namespace == namespace {
			res = append(res, rec)
		}
	}
	return res
}

func (c *ExampleResourceClient) convertPersistedRecord(s *ExampleStorageRecord, r model.Resource) error {
	r.SetMeta(&ExampleMeta{
		Namespace: s.Namespace,
		Name:      s.Name,
	})
	return json.Unmarshal([]byte(s.Spec), r.GetSpec())
}
