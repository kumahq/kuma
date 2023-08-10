package etcd

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	clientv3 "go.etcd.io/etcd/client/v3"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/common/etcd"
)

type IndexStore interface {
	Create(context.Context, core_model.Resource, ...store.CreateOptionsFunc) error
	Update(context.Context, core_model.Resource, ...store.UpdateOptionsFunc) error
	Delete(context.Context, core_model.Resource, ...store.DeleteIndexOptionsFunc) error
	List(context.Context, core_model.ResourceList, ...store.ListIndexOptionsFunc) error
}

type EtcdIndexStore struct {
	prefix  string
	client  *clientv3.Client
	metrics core_metrics.Metrics
}

func (e EtcdIndexStore) Create(ctx context.Context, resource core_model.Resource, optionsFunc ...store.CreateOptionsFunc) error {
	opts := store.NewCreateOptions(optionsFunc...)
	if opts.Owner != nil {
		key := etcd.Key{
			Typ:  resource.Descriptor().Name,
			Mesh: opts.Mesh,
			Name: opts.Name,
		}
		ownerKey := etcd.Key{
			Typ:  opts.Owner.Descriptor().Name,
			Mesh: opts.Owner.GetMeta().GetMesh(),
			Name: opts.Owner.GetMeta().GetName(),
		}
		indexKey := etcd.NewEtcdIndexKey(e.prefix, &key, &ownerKey)

		resourceMetaWrap := store.ResouceMetaWrap(store.WithMeshName(opts.Mesh), store.WithName(opts.Name), store.WithVersion("0"),
			store.WithCreationTime(opts.CreationTime), store.WithModificationTime(opts.CreationTime))
		etcdResourceMeta := ToEtcdResourceMetaObject(resourceMetaWrap, nil)
		object := etcdResourceMeta.(*etcdResourceMetaObject)
		metaBytes, err := core_model.ToJSON(indexResourceObject{EtcdResourceMetaObject: *object, Type: resource.Descriptor().Name})
		if err != nil {
			return err
		}

		_, err = e.client.Put(ctx, indexKey.String(), string(metaBytes))
		if err != nil {
			return err
		}
	}
	return nil
}

func (e EtcdIndexStore) Update(ctx context.Context, resource core_model.Resource, optionsFunc ...store.UpdateOptionsFunc) error {
	return nil
}

func (e EtcdIndexStore) Delete(ctx context.Context, resource core_model.Resource, optionsFunc ...store.DeleteIndexOptionsFunc) error {
	opts := store.NewDeleteIndexOptions(optionsFunc...)
	key := etcd.Key{
		Typ:  opts.Type,
		Mesh: opts.Mesh,
		Name: opts.Name,
	}
	var ownerkey *etcd.Key
	if len(string(opts.Owner.Type)) != 0 || len(opts.Owner.Mesh) != 0 || len(opts.Owner.Name) != 0 {
		ownerkey = &etcd.Key{
			Typ:  opts.Owner.Type,
			Mesh: opts.Owner.Mesh,
			Name: opts.Owner.Name,
		}
	}
	indexKey := etcd.NewEtcdIndexKey(e.prefix, &key, ownerkey)
	_, err := e.client.Delete(ctx, indexKey.String(), clientv3.WithPrefix())
	if err != nil {
		return err
	}
	return nil
}

func (e EtcdIndexStore) List(ctx context.Context, list core_model.ResourceList, optionsFunc ...store.ListIndexOptionsFunc) error {
	opts := store.NewListIndexOptions(optionsFunc...)
	key := etcd.Key{
		Typ:  opts.Type,
		Mesh: opts.Mesh,
		Name: opts.Name,
	}
	indexKey := etcd.NewEtcdIndexKey(e.prefix, &key, nil).String()

	response, err := e.client.Get(ctx, indexKey, clientv3.WithPrefix())
	if err != nil {
		return errors.Wrap(err, "index list error")
	}
	var count uint32
	for _, kv := range response.Kvs {
		value := kv.Value
		var indexResourceObject indexResourceObject
		err = json.Unmarshal(value, &indexResourceObject)
		if err != nil {
			return errors.Wrap(err, "json.Unmarshal error")
		}
		item := list.NewItem()
		item.SetMeta(&indexResourceObject.EtcdResourceMetaObject)
		if i, ok := item.(*IndexResource); ok {
			i.SetType(indexResourceObject.Type)
			i.SetIndexKey(string(kv.Key))
		}

		err = list.AddItem(item)
		if err != nil {
			return err
		}
		count++
	}
	list.GetPagination().SetTotal(count)

	return nil
}

func NewEtcdIndexStore(prefix string, metrics core_metrics.Metrics, client *clientv3.Client) (*EtcdIndexStore, error) {
	return &EtcdIndexStore{
		metrics: metrics,
		client:  client,
		prefix:  prefix,
	}, nil
}
