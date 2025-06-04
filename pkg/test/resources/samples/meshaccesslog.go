package samples

import (
	"github.com/kumahq/kuma/api/mesh/v1alpha1"
	meshaccesslog_proto "github.com/kumahq/kuma/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
)

func MeshAccessLogFileConf() *builders.MeshAccessLogConfBuilder {
	return builders.MeshAccessLogConf().AddFileBackend(LogFileBackend())
}

func LogFileBackend() *meshaccesslog_proto.FileBackend {
	return &meshaccesslog_proto.FileBackend{
		Path: "/tmp/access.logs",
	}
}

func MeshAccessLogWithFileBackend() *meshaccesslog_proto.MeshAccessLogResource {
	return builders.MeshAccessLog().
		WithTargetRef(builders.TargetRefService("web")).
		AddTo(builders.TargetRefMesh(), MeshAccessLogFileConf()).
		AddTo(builders.TargetRefMesh(), MeshAccessLogFileConf()).
		Build()
}

func MeshAccessLogWithZoneOriginLabel() *meshaccesslog_proto.MeshAccessLogResource {
	return builders.MeshAccessLog().
		WithName("mal-with-origin").
		WithLabels(map[string]string{
			v1alpha1.ResourceOriginLabel: string(v1alpha1.ZoneResourceOrigin),
		}).
		WithTargetRef(builders.TargetRefService("web")).
		AddTo(builders.TargetRefMesh(), MeshAccessLogFileConf()).
		AddTo(builders.TargetRefMesh(), MeshAccessLogFileConf()).
		Build()
}
