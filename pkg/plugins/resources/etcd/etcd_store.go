package etcd

import (
	"context"
	"encoding/json"
	config "github.com/kumahq/kuma/pkg/config/plugins/resources/etcd"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/common/etcd"
	"github.com/pkg/errors"
	clientv3 "go.etcd.io/etcd/client/v3"
	"strconv"
)

type EtcdStore struct {
	prefix  string
	client  *clientv3.Client
	metrics core_metrics.Metrics
}

func (e EtcdStore) Create(ctx context.Context, resource core_model.Resource, optionsFunc ...store.CreateOptionsFunc) error {
	opts := store.NewCreateOptions(optionsFunc...)
	specBytes, err := core_model.ToJSON(resource.GetSpec())
	if err != nil {
		return errors.Wrap(err, "failed to convert spec to json")
	}

	resourceMetaWrap := store.ResouceMetaWrap(store.WithMeshName(opts.Mesh), store.WithName(opts.Name), store.WithVersion("0"),
		store.WithCreationTime(opts.CreationTime), store.WithModificationTime(opts.CreationTime))
	etcdResourceMeta := ToEtcdResourceMetaObject(resourceMetaWrap)
	metaBytes, err := core_model.ToJSON(etcdResourceMeta)
	if err != nil {
		return errors.Wrap(err, "failed to convert spec to json")
	}
	value := resourceObject{
		MetaData: metaBytes,
		SpecData: specBytes,
	}
	if opts.Owner != nil {
		value.Owner.Name = opts.Owner.GetMeta().GetName()
		value.Owner.Mesh = opts.Owner.GetMeta().GetMesh()
		value.Owner.Type = opts.Owner.Descriptor().Name
	}

	bytes, err := json.Marshal(value)
	if err != nil {
		return errors.Wrap(err, "json.Marshal error")
	}
	key := etcd.NewEtcdKey(e.prefix, resource.Descriptor().Name, opts.Mesh, opts.Name).String()
	_, err = e.client.Put(ctx, key, string(bytes))
	if err != nil {
		return errors.Wrap(err, "etcd put error")
	}
	// update resource's meta with new version
	resource.SetMeta(resourceMetaWrap)

	return nil
}

func (e EtcdStore) Update(ctx context.Context, resource core_model.Resource, optionsFunc ...store.UpdateOptionsFunc) error {
	opts := store.NewUpdateOptions(optionsFunc...)

	specBytes, err := core_model.ToJSON(resource.GetSpec())
	if err != nil {
		return errors.Wrap(err, "failed to convert spec to json")
	}

	version, err := strconv.Atoi(resource.GetMeta().GetVersion())
	if err != nil {
		return errors.Wrap(err, "failed to convert meta version to int")
	}
	newVersion := version + 1
	resourceMeta := resource.GetMeta()
	resourceMetaWrap := store.ResouceMetaWrap(store.WithResourceMeta(resourceMeta), store.WithVersion(strconv.Itoa(newVersion)), store.WithModificationTime(opts.ModificationTime))
	key := etcd.NewEtcdKey(e.prefix, resource.Descriptor().Name, resource.GetMeta().GetMesh(), resource.GetMeta().GetName()).String()
	etcdResourceMeta := ToEtcdResourceMetaObject(resourceMetaWrap)
	metaBytes, err := core_model.ToJSON(etcdResourceMeta)
	if err != nil {
		return errors.Wrap(err, "failed to convert spec to json")
	}
	value := resourceObject{
		MetaData: metaBytes,
		SpecData: specBytes,
	}
	bytes, err := json.Marshal(value)
	if err != nil {
		return errors.Wrap(err, "json.Marshal error")
	}
	_, err = e.client.Put(ctx, key, string(bytes))
	if err != nil {
		return errors.Wrap(err, "etcd put error")
	}
	// update resource's meta with new version
	resource.SetMeta(resourceMetaWrap)

	return nil
}

func (e EtcdStore) Delete(ctx context.Context, resource core_model.Resource, optionsFunc ...store.DeleteOptionsFunc) error {
	opts := store.NewDeleteOptions(optionsFunc...)
	key := etcd.NewEtcdKey(e.prefix, resource.Descriptor().Name, opts.Mesh, opts.Name).String()
	response, err := e.client.Delete(context.Background(), key)
	if err != nil {
		return errors.Wrap(err, "etcd delete error")
	}
	if response.Deleted == 0 {
		return store.ErrorResourceNotFound(resource.Descriptor().Name, opts.Name, opts.Mesh)

	}

	return nil
}

func (e EtcdStore) Get(ctx context.Context, resource core_model.Resource, optionsFunc ...store.GetOptionsFunc) error {
	opts := store.NewGetOptions(optionsFunc...)
	key := etcd.NewEtcdKey(e.prefix, resource.Descriptor().Name, opts.Mesh, opts.Name).String()

	response, err := e.client.Get(context.Background(), key)
	if err != nil {
		return errors.Wrap(err, "etcd get error")
	}
	if response.Count == 0 {
		return store.ErrorResourceNotFound(resource.Descriptor().Name, opts.Name, opts.Mesh)
	}

	if response.Count != 1 {
		return errors.New("get count is not one")
	}

	value := response.Kvs[0].Value
	var resourceObject resourceObject
	err = json.Unmarshal(value, &resourceObject)
	if err != nil {
		return errors.Wrap(err, "json.Unmarshal error")
	}

	if err := core_model.FromJSON(resourceObject.SpecData, resource.GetSpec()); err != nil {
		return errors.Wrap(err, "failed to convert json to spec")
	}
	etcdResourceMeta := ToEtcdResourceMetaObject(resource.GetMeta())
	if err := core_model.FromJSON(resourceObject.MetaData, etcdResourceMeta); err != nil {
		return errors.Wrap(err, "failed to convert json to meta")
	}

	resource.SetMeta(etcdResourceMeta)

	if opts.Version != "" && resource.GetMeta().GetVersion() != opts.Version {
		return store.ErrorResourcePreconditionFailed(resource.Descriptor().Name, opts.Name, opts.Mesh)
	}

	return nil
}

func (e EtcdStore) List(ctx context.Context, list core_model.ResourceList, optionsFunc ...store.ListOptionsFunc) error {
	opts := store.NewListOptions(optionsFunc...)
	key := etcd.NewEtcdKey(e.prefix, list.GetItemType(), opts.Mesh, "").Prefix()

	//start := time.Now()
	response, err := e.client.Get(context.Background(), key, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	//log.WithName("########").Info("get etcd", "cost", time.Since(start), "key", key)
	if err != nil {
		return errors.Wrap(err, "etcd get error")
	}
	for _, kv := range response.Kvs {
		value := kv.Value

		var resourceObject resourceObject
		err = json.Unmarshal(value, &resourceObject)
		if err != nil {
			return errors.Wrap(err, "json.Unmarshal error")
		}
		item := list.NewItem()
		if err := core_model.FromJSON(resourceObject.SpecData, item.GetSpec()); err != nil {
			return errors.Wrap(err, "failed to convert json to spec")
		}
		etcdResourceMeta := ToEtcdResourceMetaObject(item.GetMeta())
		if err := core_model.FromJSON(resourceObject.MetaData, etcdResourceMeta); err != nil {
			return errors.Wrap(err, "failed to convert json to meta")
		}

		item.SetMeta(etcdResourceMeta)

		if err := list.AddItem(item); err != nil {
			return errors.Wrap(err, "list.AddItem error")
		}
	}
	list.GetPagination().SetTotal(uint32(response.Count))
	return nil
}

func newEtcdStore(prefix string, metrics core_metrics.Metrics, config *config.EtcdConfig) (*EtcdStore, error) {
	client, err := clientv3.NewFromURLs(config.Endpoints)
	if err != nil {
		return nil, err
	}
	return &EtcdStore{
		metrics: metrics,
		client:  client,
		prefix:  prefix,
	}, nil
}
