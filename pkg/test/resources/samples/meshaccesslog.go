package samples

import (
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
