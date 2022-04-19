package memory

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/events"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

type resourceKey struct {
	Name         string
	Mesh         string
	ResourceType string
}

type memoryStoreRecord struct {
	resourceKey
	Version          memoryVersion
	Spec             string
	CreationTime     time.Time
	ModificationTime time.Time
	Children         []*resourceKey
}
type memoryStoreRecords = []*memoryStoreRecord

var _ model.ResourceMeta = &memoryMeta{}

type memoryMeta struct {
	Name             string
	Mesh             string
	Version          memoryVersion
	CreationTime     time.Time
	ModificationTime time.Time
}

func (m memoryMeta) GetName() string {
	return m.Name
}

func (m memoryMeta) GetNameExtensions() model.ResourceNameExtensions {
	return model.ResourceNameExtensionsUnsupported
}

func (m memoryMeta) GetMesh() string {
	return m.Mesh
}

func (m memoryMeta) GetVersion() string {
	return m.Version.String()
}

func (m memoryMeta) GetCreationTime() time.Time {
	return m.CreationTime
}

func (m memoryMeta) GetModificationTime() time.Time {
	return m.ModificationTime
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
	records     memoryStoreRecords
	mu          sync.RWMutex
	eventWriter events.Emitter
}

func NewStore() store.ResourceStore {
	return &memoryStore{}
}

func (c *memoryStore) SetEventWriter(writer events.Emitter) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.eventWriter = writer
}

func (c *memoryStore) Create(_ context.Context, r model.Resource, fs ...store.CreateOptionsFunc) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	opts := store.NewCreateOptions(fs...)
	// Name must be provided via CreateOptions
	if _, record := c.findRecord(string(r.Descriptor().Name), opts.Name, opts.Mesh); record != nil {
		return store.ErrorResourceAlreadyExists(r.Descriptor().Name, opts.Name, opts.Mesh)
	}

	meta := memoryMeta{
		Name:             opts.Name,
		Mesh:             opts.Mesh,
		Version:          initialVersion(),
		CreationTime:     opts.CreationTime,
		ModificationTime: opts.CreationTime,
	}

	// fill the meta
	r.SetMeta(meta)

	// convert into storage representation
	record, err := c.marshalRecord(
		string(r.Descriptor().Name),
		meta,
		r.GetSpec())
	if err != nil {
		return err
	}

	if opts.Owner != nil {
		_, ownerRecord := c.findRecord(string(opts.Owner.Descriptor().Name), opts.Owner.GetMeta().GetName(), opts.Owner.GetMeta().GetMesh())
		if ownerRecord == nil {
			return store.ErrorResourceNotFound(opts.Owner.Descriptor().Name, opts.Owner.GetMeta().GetName(), opts.Owner.GetMeta().GetMesh())
		}
		ownerRecord.Children = append(ownerRecord.Children, &record.resourceKey)
	}

	// persist
	c.records = append(c.records, record)
	if c.eventWriter != nil {
		go func() {
			c.eventWriter.Send(events.ResourceChangedEvent{
				Operation: events.Create,
				Type:      r.Descriptor().Name,
				Key:       model.MetaToResourceKey(r.GetMeta()),
			})
		}()
	}
	return nil
}

func (c *memoryStore) Update(_ context.Context, r model.Resource, fs ...store.UpdateOptionsFunc) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	opts := store.NewUpdateOptions(fs...)

	meta, ok := (r.GetMeta()).(memoryMeta)
	if !ok {
		return fmt.Errorf("MemoryStore.Update() requires r.GetMeta() to be of type memoryMeta")
	}

	// Name must be provided via r.GetMeta()
	mesh := r.GetMeta().GetMesh()
	idx, record := c.findRecord(string(r.Descriptor().Name), r.GetMeta().GetName(), mesh)
	if record == nil || meta.Version != record.Version {
		return store.ErrorResourceConflict(r.Descriptor().Name, r.GetMeta().GetName(), r.GetMeta().GetMesh())
	}
	meta.Version = meta.Version.Next()
	meta.ModificationTime = opts.ModificationTime

	record, err := c.marshalRecord(
		string(r.Descriptor().Name),
		meta,
		r.GetSpec())
	if err != nil {
		return err
	}

	// persist
	c.records[idx] = record

	r.SetMeta(meta)
	if c.eventWriter != nil {
		go func() {
			c.eventWriter.Send(events.ResourceChangedEvent{
				Operation: events.Update,
				Type:      r.Descriptor().Name,
				Key:       model.MetaToResourceKey(r.GetMeta()),
			})
		}()
	}
	return nil
}

func (c *memoryStore) Delete(ctx context.Context, r model.Resource, fs ...store.DeleteOptionsFunc) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.delete(r, fs...)
}

func (c *memoryStore) delete(r model.Resource, fs ...store.DeleteOptionsFunc) error {
	opts := store.NewDeleteOptions(fs...)

	_, ok := (r.GetMeta()).(memoryMeta)
	if r.GetMeta() != nil && !ok {
		return fmt.Errorf("MemoryStore.Delete() requires r.GetMeta() either to be nil or to be of type memoryMeta")
	}

	// Name must be provided via DeleteOptions
	idx, record := c.findRecord(string(r.Descriptor().Name), opts.Name, opts.Mesh)
	if record == nil {
		return store.ErrorResourceNotFound(r.Descriptor().Name, opts.Name, opts.Mesh)
	}
	for _, child := range record.Children {
		_, childRecord := c.findRecord(child.ResourceType, child.Name, child.Mesh)
		if childRecord == nil {
			continue // resource was already deleted
		}
		obj, err := registry.Global().NewObject(model.ResourceType(child.ResourceType))
		if err != nil {
			return fmt.Errorf("MemoryStore.Delete() couldn't unmarshal child resource")
		}
		if err := c.unmarshalRecord(childRecord, obj); err != nil {
			return fmt.Errorf("MemoryStore.Delete() couldn't unmarshal child resource")
		}
		if err := c.delete(obj, store.DeleteByKey(childRecord.Name, childRecord.Mesh)); err != nil {
			return fmt.Errorf("MemoryStore.Delete() couldn't delete linked child resource")
		}
	}
	c.records = append(c.records[:idx], c.records[idx+1:]...)
	if c.eventWriter != nil {
		go func() {
			c.eventWriter.Send(events.ResourceChangedEvent{
				Operation: events.Delete,
				Type:      r.Descriptor().Name,
				Key: model.ResourceKey{
					Mesh: opts.Mesh,
					Name: opts.Name,
				},
			})
		}()
	}
	return nil
}

func (c *memoryStore) Get(_ context.Context, r model.Resource, fs ...store.GetOptionsFunc) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	opts := store.NewGetOptions(fs...)
	// Name must be provided via GetOptions
	_, record := c.findRecord(string(r.Descriptor().Name), opts.Name, opts.Mesh)
	if record == nil {
		return store.ErrorResourceNotFound(r.Descriptor().Name, opts.Name, opts.Mesh)
	}
	if opts.Version != "" && opts.Version != record.Version.String() {
		return store.ErrorResourcePreconditionFailed(r.Descriptor().Name, opts.Name, opts.Mesh)
	}
	return c.unmarshalRecord(record, r)
}

func (c *memoryStore) List(_ context.Context, rs model.ResourceList, fs ...store.ListOptionsFunc) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	opts := store.NewListOptions(fs...)

	records := c.findRecords(string(rs.GetItemType()), opts.Mesh)

	for i := 0; i < len(records); i++ {
		r := rs.NewItem()
		if err := c.unmarshalRecord(records[i], r); err != nil {
			return err
		}
		_ = rs.AddItem(r)
	}

	rs.GetPagination().SetTotal(uint32(len(records)))

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
		resourceKey: resourceKey{
			ResourceType: resourceType,
			// Name must be provided via CreateOptions
			Name: meta.Name,
			Mesh: meta.Mesh,
		},
		Version:          meta.Version,
		Spec:             string(content),
		CreationTime:     meta.CreationTime,
		ModificationTime: meta.ModificationTime,
	}, nil
}

func (c *memoryStore) unmarshalRecord(s *memoryStoreRecord, r model.Resource) error {
	r.SetMeta(memoryMeta{
		Name:             s.Name,
		Mesh:             s.Mesh,
		Version:          s.Version,
		CreationTime:     s.CreationTime,
		ModificationTime: s.ModificationTime,
	})
	return util_proto.FromJSON([]byte(s.Spec), r.GetSpec())
}
