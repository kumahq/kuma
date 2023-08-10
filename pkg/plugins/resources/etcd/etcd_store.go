package etcd

import (
	"context"

	"github.com/pkg/errors"
	clientv3 "go.etcd.io/etcd/client/v3"

	config "github.com/kumahq/kuma/pkg/config/plugins/resources/etcd"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/common/etcd"
)

type EtcdStore struct {
	resourceEtcdStore store.ResourceStore
	indexEtcdStore    IndexStore
	client            *clientv3.Client
}

func NewEtcdStore(prefix string, metrics core_metrics.Metrics, config *config.EtcdConfig) (store.ResourceStore, error) {
	client, err := etcd.NewClient(config)
	if err != nil {
		return nil, err
	}

	resourceEtcdStore, err := NewResoucesEtcdStore(prefix, metrics, client)
	if err != nil {
		return nil, err
	}
	etcdIndexStore, err := NewEtcdIndexStore(prefix, metrics, client)
	if err != nil {
		return nil, err
	}

	return EtcdStore{
		resourceEtcdStore: resourceEtcdStore,
		indexEtcdStore:    etcdIndexStore,
		client:            client,
	}, nil
}

func (e EtcdStore) Create(ctx context.Context, m core_model.Resource, optionsFunc ...store.CreateOptionsFunc) error {
	err := e.indexEtcdStore.Create(ctx, m, optionsFunc...)
	if err != nil {
		return errors.Wrap(err, "create etcd index error")
	}
	err = e.resourceEtcdStore.Create(ctx, m, optionsFunc...)
	if err != nil {
		return errors.Wrap(err, "create etcd resource error")
	}

	return nil
}

func (e EtcdStore) Update(ctx context.Context, m core_model.Resource, optionsFunc ...store.UpdateOptionsFunc) error {
	err := e.indexEtcdStore.Update(ctx, m, optionsFunc...)
	if err != nil {
		return errors.Wrap(err, "update etcd resource error")
	}
	err = e.resourceEtcdStore.Update(ctx, m, optionsFunc...)
	if err != nil {
		return errors.Wrap(err, "update etcd resource error")
	}
	return nil
}

func (e EtcdStore) Delete(ctx context.Context, m core_model.Resource, optionsFunc ...store.DeleteOptionsFunc) error {
	opts := store.NewDeleteOptions(optionsFunc...)

	indexResourceList := NewIndexResourceList()
	err := e.indexEtcdStore.List(ctx, indexResourceList, store.WithListIndexOptions(m.Descriptor().Name, opts.Mesh, opts.Name))
	if err != nil {
		return errors.Wrap(err, "delete etcd resource error")
	}

	items := indexResourceList.GetItems()
	for _, item := range items {
		meta := item.GetMeta()
		if etcdResourceMetaObject, ok := meta.(*etcdResourceMetaObject); ok {
			if err := e.Delete(ctx, item, store.DeleteByKey(etcdResourceMetaObject.Name, etcdResourceMetaObject.Mesh)); err != nil {
				return err
			}
		}
	}

	err = e.resourceEtcdStore.Get(ctx, m, store.GetByKey(opts.Name, opts.Mesh))
	if err != nil {
		return err
	}

	if meta, ok := m.GetMeta().(GetOwner); ok && meta.GetOwner() != nil {
		err = e.indexEtcdStore.Delete(ctx, m, store.WithDeleteIndexOptions(m.Descriptor().Name, opts.Mesh, opts.Name),
			store.WithDeleteIndexOptions(meta.GetOwner().Type, meta.GetOwner().Mesh, meta.GetOwner().Name))
		if err != nil {
			return err
		}
	}

	err = e.resourceEtcdStore.Delete(ctx, m, optionsFunc...)
	if err != nil {
		return err
	}
	return nil
}

func (e EtcdStore) Get(ctx context.Context, m core_model.Resource, optionsFunc ...store.GetOptionsFunc) error {
	err := e.resourceEtcdStore.Get(ctx, m, optionsFunc...)
	if err != nil {
		return err
	}
	return nil
}

func (e EtcdStore) List(ctx context.Context, list core_model.ResourceList, optionsFunc ...store.ListOptionsFunc) error {
	err := e.resourceEtcdStore.List(ctx, list, optionsFunc...)
	if err != nil {
		return errors.Wrap(err, "list etcd resource error")
	}
	return nil
}
