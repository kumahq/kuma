package client

import (
	"context"
	"crypto/tls"
	"net/url"
	"time"

	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/golang/protobuf/ptypes"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/registry"
	"github.com/Kong/kuma/pkg/util/proto"
)

type KDSClient interface {
	StartStream() (KDSStream, error)
	Close() error
}

type client struct {
	conn   *grpc.ClientConn
	client mesh_proto.KumaDiscoveryServiceClient
}

var _ KDSClient = &client{}

func New(serverURL string) (KDSClient, error) {
	u, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}
	var dialOpts []grpc.DialOption
	switch u.Scheme {
	case "http":
		dialOpts = append(dialOpts, grpc.WithInsecure())
	case "https":
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: true, // it's acceptable since we don't pass any secrets to the server
		})))
	default:
		return nil, errors.Errorf("unsupported scheme %q. Use one of %s", u.Scheme, []string{"grpc", "grpcs"})
	}
	conn, err := grpc.Dial(u.Host, dialOpts...)
	if err != nil {
		return nil, err
	}
	c := mesh_proto.NewKumaDiscoveryServiceClient(conn)
	return &client{
		conn:   conn,
		client: c,
	}, nil
}

func (c *client) StartStream() (KDSStream, error) {
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.MD{})
	stream, err := c.client.StreamKumaResources(ctx)
	if err != nil {
		return nil, err
	}
	return NewKDSStream(stream), nil
}

func (c *client) Close() error {
	return c.conn.Close()
}

func ToKumaResources(response *envoy.DiscoveryResponse) (model.ResourceList, error) {
	krs := []*mesh_proto.KumaResource{}
	for _, r := range response.Resources {
		kr := &mesh_proto.KumaResource{}
		if err := ptypes.UnmarshalAny(r, kr); err != nil {
			return nil, err
		}
		krs = append(krs, kr)
	}
	return toResources(model.ResourceType(response.TypeUrl), krs)
}

func toResources(resourceType model.ResourceType, krs []*mesh_proto.KumaResource) (model.ResourceList, error) {
	list, err := registry.Global().NewList(resourceType)
	if err != nil {
		return nil, err
	}
	for _, kr := range krs {
		obj, err := registry.Global().NewObject(resourceType)
		if err != nil {
			return nil, err
		}
		err = ptypes.UnmarshalAny(kr.Spec, obj.GetSpec())
		if err != nil {
			return nil, err
		}
		obj.SetMeta(toResourceMeta(kr.Meta))
		if err := list.AddItem(obj); err != nil {
			return nil, err
		}
	}
	return list, nil
}

type resourceMeta struct {
	name             string
	version          string
	mesh             string
	creationTime     *time.Time
	modificationTime *time.Time
}

func (r *resourceMeta) GetName() string {
	return r.name
}

func (r *resourceMeta) GetNameExtensions() model.ResourceNameExtensions {
	return model.ResourceNameExtensionsUnsupported
}

func (r *resourceMeta) GetVersion() string {
	return r.version
}

func (r *resourceMeta) GetMesh() string {
	return r.mesh
}

func (r *resourceMeta) GetCreationTime() time.Time {
	return *r.creationTime
}

func (r *resourceMeta) GetModificationTime() time.Time {
	return *r.modificationTime
}

func toResourceMeta(meta *mesh_proto.KumaResource_Meta) model.ResourceMeta {
	return &resourceMeta{
		name:             meta.Name,
		mesh:             meta.Mesh,
		version:          meta.Version,
		creationTime:     proto.MustTimestampFromProto(meta.CreationTime),
		modificationTime: proto.MustTimestampFromProto(meta.ModificationTime),
	}
}
