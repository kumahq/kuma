package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
)

type memoryStoreRecord struct {
	ResourceType string
	Namespace    string
	Name         string
	Spec         string
}
type memoryStoreRecords = []*memoryStoreRecord

var _ model.ResourceMeta = &memoryMeta{}

type memoryMeta struct {
	Namespace string
	Name      string
}

func (m *memoryMeta) GetName() string {
	return m.Name
}
func (m *memoryMeta) GetNamespace() string {
	return m.Namespace
}
func (m *memoryMeta) GetVersion() string {
	return "0"
}

var _ store.ResourceStore = &memoryStore{}

type memoryStore struct {
	records memoryStoreRecords
	mu      sync.RWMutex
}

func NewStore() (store.ResourceStore) {
	return &memoryStore{}
}

func (c *memoryStore) Create(_ context.Context, r model.Resource, fs ...store.CreateOptionsFunc) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	opts := store.NewCreateOptions(fs...)

	// Namespace and Name must be provided via CreateOptions
	if _, record := c.findRecord(string(r.GetType()), opts.Namespace, opts.Name); record != nil {
		return store.ErrorResourceAlreadyExists(r.GetType(), opts.Namespace, opts.Name)
	}

	// fill the meta
	r.SetMeta(&memoryMeta{
		Name: opts.Name,
		Namespace: opts.Namespace,
	})

	// convert into storage representation
	record, err := c.marshalRecord(
		string(r.GetType()),
		// Namespace and Name must be provided via CreateOptions
		&memoryMeta{
			Namespace: opts.Namespace,
			Name:      opts.Name,
		},
		r.GetSpec())
	if err != nil {
		return err
	}

	// persist
	c.records = append(c.records, record)
	return nil
}
func (c *memoryStore) Update(_ context.Context, r model.Resource, fs ...store.UpdateOptionsFunc) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_ = store.NewUpdateOptions(fs...)

	_, ok := (r.GetMeta()).(*memoryMeta)
	if !ok {
		return fmt.Errorf("MemoryStore.Update() requires r.GetMeta() to be of type memoryMeta")
	}

	// Namespace and Name must be provided via r.GetMeta()
	idx, record := c.findRecord(string(r.GetType()), r.GetMeta().GetNamespace(), r.GetMeta().GetName())
	if record == nil {
		return store.ErrorResourceNotFound(r.GetType(), r.GetMeta().GetNamespace(), r.GetMeta().GetName())
	}

	record, err := c.marshalRecord(
		string(r.GetType()),
		// Namespace and Name must be provided via r.GetMeta()
		r.GetMeta(), r.GetSpec())
	if err != nil {
		return err
	}

	// persist
	c.records[idx] = record
	return nil
}
func (c *memoryStore) Delete(_ context.Context, r model.Resource, fs ...store.DeleteOptionsFunc) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	opts := store.NewDeleteOptions(fs...)

	_, ok := (r.GetMeta()).(*memoryMeta)
	if r.GetMeta() != nil && !ok {
		return fmt.Errorf("MemoryStore.Delete() requires r.GetMeta() either to be nil or to be of type memoryMeta")
	}

	// Namespace and Name must be provided via DeleteOptions
	idx, record := c.findRecord(string(r.GetType()), opts.Namespace, opts.Name)
	if record != nil {
		c.records = append(c.records[:idx], c.records[idx+1:]...)
	}
	return nil
}

func (c *memoryStore) Get(_ context.Context, r model.Resource, fs ...store.GetOptionsFunc) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	opts := store.NewGetOptions(fs...)

	// Namespace and Name must be provided via GetOptions
	_, record := c.findRecord(string(r.GetType()), opts.Namespace, opts.Name)
	if record == nil {
		return store.ErrorResourceNotFound(r.GetType(), opts.Namespace, opts.Name)
	}
	return c.unmarshalRecord(record, r)
}
func (c *memoryStore) List(_ context.Context, rs model.ResourceList, fs ...store.ListOptionsFunc) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	opts := store.NewListOptions(fs...)

	// Namespace must be provided via ListOptions
	records := c.findRecords(string(rs.GetItemType()), opts.Namespace)
	for _, record := range records {
		r := rs.NewItem()
		if err := c.unmarshalRecord(record, r); err != nil {
			return err
		}
		_ = rs.AddItem(r)
	}
	return nil
}

func (c *memoryStore) findRecord(
	resourceType string, namespace string, name string) (int, *memoryStoreRecord) {
	for idx, rec := range c.records {
		if rec.ResourceType == resourceType &&
			rec.Namespace == namespace &&
			rec.Name == name {
			return idx, rec
		}
	}
	return -1, nil
}

func (c *memoryStore) findRecords(
	resourceType string, namespace string) []*memoryStoreRecord {
	res := make([]*memoryStoreRecord, 0)
	for _, rec := range c.records {
		if rec.ResourceType == resourceType &&
			rec.Namespace == namespace {
			res = append(res, rec)
		}
	}
	return res
}

func (c *memoryStore) marshalRecord(resourceType string, meta model.ResourceMeta, spec model.ResourceSpec) (*memoryStoreRecord, error) {
	// convert spec into storage representation
	content, err := json.Marshal(spec)
	if err != nil {
		return nil, err
	}
	return &memoryStoreRecord{
		ResourceType: resourceType,
		// Namespace and Name must be provided via CreateOptions
		Namespace: meta.GetNamespace(),
		Name:      meta.GetName(),
		Spec:      string(content),
	}, nil
}

func (c *memoryStore) unmarshalRecord(s *memoryStoreRecord, r model.Resource) error {
	r.SetMeta(&memoryMeta{
		Namespace: s.Namespace,
		Name:      s.Name,
	})
	return json.Unmarshal([]byte(s.Spec), r.GetSpec())
}
