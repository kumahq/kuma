package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
)

type MemoryStoreRecord struct {
	ResourceType string
	Namespace    string
	Name         string
	Spec         string
}
type MemoryStoreRecords = []*MemoryStoreRecord

var _ model.ResourceMeta = &MemoryMeta{}

type MemoryMeta struct {
	Namespace string
	Name      string
	Version   string
}

func (m *MemoryMeta) GetName() string {
	return m.Name
}
func (m *MemoryMeta) GetNamespace() string {
	return m.Namespace
}
func (m *MemoryMeta) GetVersion() string {
	return m.Version
}

var _ store.ResourceStore = &MemoryStore{}

type MemoryStore struct {
	Records MemoryStoreRecords
	mu      sync.RWMutex
}

func (c *MemoryStore) Create(_ context.Context, r model.Resource, fs ...store.CreateOptionsFunc) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	opts := store.NewCreateOptions(fs...)

	// Namespace and Name must be provided via CreateOptions
	if _, record := c.findRecord(string(r.GetType()), opts.Namespace, opts.Name); record != nil {
		return store.ErrorResourceAlreadyExists(r.GetType(), opts.Namespace, opts.Name)
	}

	// convert into storage representation
	record, err := c.marshalRecord(
		string(r.GetType()),
		// Namespace and Name must be provided via CreateOptions
		&MemoryMeta{
			Namespace: opts.Namespace,
			Name:      opts.Name,
		},
		r.GetSpec())
	if err != nil {
		return err
	}

	// persist
	c.Records = append(c.Records, record)
	return nil
}
func (c *MemoryStore) Update(_ context.Context, r model.Resource, fs ...store.UpdateOptionsFunc) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_ = store.NewUpdateOptions(fs...)

	meta, ok := (r.GetMeta()).(*MemoryMeta)
	if !ok {
		return fmt.Errorf("MemoryStore.Update() requires r.GetMeta() to be of type MemoryMeta")
	}

	// Namespace and Name must be provided via r.GetMeta()
	idx, record := c.findRecord(string(r.GetType()), meta.Namespace, meta.Name)
	if record == nil {
		return store.ErrorResourceNotFound(r.GetType(), meta.Namespace, meta.Name)
	}

	record, err := c.marshalRecord(
		string(r.GetType()),
		// Namespace and Name must be provided via r.GetMeta()
		r.GetMeta(), r.GetSpec())
	if err != nil {
		return err
	}

	// persist
	c.Records[idx] = record
	return nil
}
func (c *MemoryStore) Delete(_ context.Context, r model.Resource, fs ...store.DeleteOptionsFunc) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	opts := store.NewDeleteOptions(fs...)

	_, ok := (r.GetMeta()).(*MemoryMeta)
	if r.GetMeta() != nil && !ok {
		return fmt.Errorf("MemoryStore.Delete() requires r.GetMeta() either to be nil or to be of type MemoryMeta")
	}

	// Namespace and Name must be provided via DeleteOptions
	idx, record := c.findRecord(string(r.GetType()), opts.Namespace, opts.Name)
	if record != nil {
		c.Records = append(c.Records[:idx], c.Records[idx+1:]...)
	}
	return nil
}

func (c *MemoryStore) Get(_ context.Context, r model.Resource, fs ...store.GetOptionsFunc) error {
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
func (c *MemoryStore) List(_ context.Context, rs model.ResourceList, fs ...store.ListOptionsFunc) error {
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
		rs.AddItem(r)
	}
	return nil
}

func (c *MemoryStore) findRecord(
	resourceType string, namespace string, name string) (int, *MemoryStoreRecord) {
	for idx, rec := range c.Records {
		if rec.ResourceType == resourceType &&
			rec.Namespace == namespace &&
			rec.Name == name {
			return idx, rec
		}
	}
	return -1, nil
}

func (c *MemoryStore) findRecords(
	resourceType string, namespace string) []*MemoryStoreRecord {
	res := make([]*MemoryStoreRecord, 0)
	for _, rec := range c.Records {
		if rec.ResourceType == resourceType &&
			rec.Namespace == namespace {
			res = append(res, rec)
		}
	}
	return res
}

func (c *MemoryStore) marshalRecord(resourceType string, meta model.ResourceMeta, spec model.ResourceSpec) (*MemoryStoreRecord, error) {
	// convert spec into storage representation
	content, err := json.Marshal(spec)
	if err != nil {
		return nil, err
	}
	return &MemoryStoreRecord{
		ResourceType: resourceType,
		// Namespace and Name must be provided via CreateOptions
		Namespace: meta.GetNamespace(),
		Name:      meta.GetName(),
		Spec:      string(content),
	}, nil
}

func (c *MemoryStore) unmarshalRecord(s *MemoryStoreRecord, r model.Resource) error {
	r.SetMeta(&MemoryMeta{
		Namespace: s.Namespace,
		Name:      s.Name,
	})
	return json.Unmarshal([]byte(s.Spec), r.GetSpec())
}
