package memory

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/store"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
)

type memoryStoreRecord struct {
	ResourceType string
	Name         string
	Mesh         string
	Version      memoryVersion
	Spec         string
}
type memoryStoreRecords = []*memoryStoreRecord

var _ model.ResourceMeta = &memoryMeta{}

type memoryMeta struct {
	Name    string
	Mesh    string
	Version memoryVersion
}

func (m memoryMeta) GetName() string {
	return m.Name
}
func (m memoryMeta) GetMesh() string {
	return m.Mesh
}
func (m memoryMeta) GetVersion() string {
	return m.Version.String()
}

type memoryVersion uint64

func initialVersion() memoryVersion {
	return memoryVersion(1)
}

func (v memoryVersion) Next() memoryVersion {
	return memoryVersion(uint64(v) + 1)
}

func (v memoryVersion) String() string {
	return strconv.FormatUint(uint64(v), 10)
}

var _ store.ResourceStore = &memoryStore{}

type memoryStore struct {
	records memoryStoreRecords
	mu      sync.RWMutex
}

func NewStore() store.ResourceStore {
	return &memoryStore{}
}

func (c *memoryStore) Create(_ context.Context, r model.Resource, fs ...store.CreateOptionsFunc) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	opts := store.NewCreateOptions(fs...)

	// Name must be provided via CreateOptions
	if _, record := c.findRecord(string(r.GetType()), opts.Name, opts.Mesh); record != nil {
		return store.ErrorResourceAlreadyExists(r.GetType(), opts.Name, opts.Mesh)
	}

	meta := memoryMeta{
		Name:    opts.Name,
		Mesh:    opts.Mesh,
		Version: initialVersion(),
	}

	// fill the meta
	r.SetMeta(meta)

	// convert into storage representation
	record, err := c.marshalRecord(
		string(r.GetType()),
		meta,
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

	meta, ok := (r.GetMeta()).(memoryMeta)
	if !ok {
		return fmt.Errorf("MemoryStore.Update() requires r.GetMeta() to be of type memoryMeta")
	}

	// Name must be provided via r.GetMeta()
	mesh := r.GetMeta().GetMesh()
	idx, record := c.findRecord(string(r.GetType()), r.GetMeta().GetName(), mesh)
	if record == nil || meta.Version != record.Version {
		return store.ErrorResourceConflict(r.GetType(), r.GetMeta().GetName(), r.GetMeta().GetMesh())
	}
	meta.Version = meta.Version.Next()

	record, err := c.marshalRecord(
		string(r.GetType()),
		meta,
		r.GetSpec())
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

	_, ok := (r.GetMeta()).(memoryMeta)
	if r.GetMeta() != nil && !ok {
		return fmt.Errorf("MemoryStore.Delete() requires r.GetMeta() either to be nil or to be of type memoryMeta")
	}

	// Name must be provided via DeleteOptions
	idx, record := c.findRecord(string(r.GetType()), opts.Name, opts.Mesh)
	if record == nil {
		return store.ErrorResourceNotFound(r.GetType(), opts.Name, opts.Mesh)
	}
	c.records = append(c.records[:idx], c.records[idx+1:]...)
	return nil
}

func (c *memoryStore) Get(_ context.Context, r model.Resource, fs ...store.GetOptionsFunc) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	opts := store.NewGetOptions(fs...)

	// Name must be provided via GetOptions
	_, record := c.findRecord(string(r.GetType()), opts.Name, opts.Mesh)
	if record == nil {
		return store.ErrorResourceNotFound(r.GetType(), opts.Name, opts.Mesh)
	}
	if opts.Version != "" && opts.Version != record.Version.String() {
		return store.ErrorResourceNotFound(r.GetType(), opts.Name, opts.Mesh)
	}
	return c.unmarshalRecord(record, r)
}
func (c *memoryStore) List(_ context.Context, rs model.ResourceList, fs ...store.ListOptionsFunc) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	opts := store.NewListOptions(fs...)

	records := c.findRecords(string(rs.GetItemType()), opts.Mesh)
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
	resourceType string, name string, mesh string) (int, *memoryStoreRecord) {
	for idx, rec := range c.records {
		if rec.ResourceType == resourceType &&
			rec.Name == name &&
			rec.Mesh == mesh {
			return idx, rec
		}
	}
	return -1, nil
}

func (c *memoryStore) findRecords(
	resourceType string, mesh string) []*memoryStoreRecord {
	res := make([]*memoryStoreRecord, 0)
	for _, rec := range c.records {
		if rec.ResourceType == resourceType &&
			(mesh == "" || rec.Mesh == mesh) {
			res = append(res, rec)
		}
	}
	return res
}

func (c *memoryStore) marshalRecord(resourceType string, meta memoryMeta, spec model.ResourceSpec) (*memoryStoreRecord, error) {
	// convert spec into storage representation
	content, err := util_proto.ToJSON(spec)
	if err != nil {
		return nil, err
	}
	return &memoryStoreRecord{
		ResourceType: resourceType,
		// Name must be provided via CreateOptions
		Name:    meta.Name,
		Mesh:    meta.Mesh,
		Version: meta.Version,
		Spec:    string(content),
	}, nil
}

func (c *memoryStore) unmarshalRecord(s *memoryStoreRecord, r model.Resource) error {
	r.SetMeta(memoryMeta{
		Name:    s.Name,
		Mesh:    s.Mesh,
		Version: s.Version,
	})
	return util_proto.FromJSON([]byte(s.Spec), r.GetSpec())
}
