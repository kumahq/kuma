package samples

import (
	"github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	meshaccesslog_proto "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
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
		WithTargetRef(builders.TargetRefDataplaneLabels("kuma.io/display-name", "web")).
		AddTo(builders.TargetRefMesh(), MeshAccessLogFileConf()).
		AddTo(builders.TargetRefMesh(), MeshAccessLogFileConf()).
		Build()
}

func MeshAccessLogWithZoneLabels() *meshaccesslog_proto.MeshAccessLogResource {
	return builders.MeshAccessLog().
		WithName("mal-with-origin").
		WithLabels(map[string]string{
			v1alpha1.ResourceOriginLabel: string(v1alpha1.ZoneResourceOrigin),
			v1alpha1.ZoneTag:             "zone-1",
			v1alpha1.EnvTag:              v1alpha1.KubernetesEnvironment,
			v1alpha1.KubeNamespaceTag:    "kuma-system",
			v1alpha1.DisplayName:         "mal-with-origin",
			v1alpha1.PolicyRoleLabel:     string(v1alpha1.SystemPolicyRole),
			"kuma.io/mesh":               "default",
			"team":                       "payments",
		}).
		WithTargetRef(builders.TargetRefDataplaneLabels("kuma.io/display-name", "web")).
		AddTo(builders.TargetRefMesh(), MeshAccessLogFileConf()).
		AddTo(builders.TargetRefMesh(), MeshAccessLogFileConf()).
		Build()
}
