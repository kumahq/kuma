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
