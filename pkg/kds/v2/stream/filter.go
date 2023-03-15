package mux

import mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"

type Filter interface {
	InterceptSession(stream mesh_proto.KDSSyncService_GlobalToZoneSyncServer) error
}
