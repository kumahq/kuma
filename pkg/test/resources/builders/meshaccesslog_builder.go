package builders

import (
	common_proto "github.com/kumahq/kuma/api/common/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	meshaccesslog_proto "github.com/kumahq/kuma/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

type MeshAccessLogBuilder struct {
	res *meshaccesslog_proto.MeshAccessLogResource
}

func MeshAccessLog() *MeshAccessLogBuilder {
	return &MeshAccessLogBuilder{
		res: &meshaccesslog_proto.MeshAccessLogResource{
			Meta: &test_model.ResourceMeta{
				Mesh: core_model.DefaultMesh,
				Name: "mal-1",
			},
			Spec: &meshaccesslog_proto.MeshAccessLog{},
		},
	}
}

func (m *MeshAccessLogBuilder) WithTargetRef(targetRef *common_proto.TargetRef) *MeshAccessLogBuilder {
	m.res.Spec.TargetRef = targetRef
	return m
}

func (m *MeshAccessLogBuilder) AddFrom(targetRef *common_proto.TargetRef, conf *MeshAccessLogConfBuilder) *MeshAccessLogBuilder {
	m.res.Spec.From = append(m.res.Spec.From, &meshaccesslog_proto.MeshAccessLog_From{
		TargetRef: targetRef,
		Default:   conf.res,
	})
	return m
}

func (m *MeshAccessLogBuilder) AddTo(targetRef *common_proto.TargetRef, conf *MeshAccessLogConfBuilder) *MeshAccessLogBuilder {
	m.res.Spec.To = append(m.res.Spec.To, &meshaccesslog_proto.MeshAccessLog_To{
		TargetRef: targetRef,
		Default:   conf.res,
	})
	return m
}

func (m *MeshAccessLogBuilder) Build() *meshaccesslog_proto.MeshAccessLogResource {
	if err := m.res.Validate(); err != nil {
		panic(err)
	}
	return m.res
}

func MeshAccessLogConf() *MeshAccessLogConfBuilder {
	return &MeshAccessLogConfBuilder{
		res: &meshaccesslog_proto.MeshAccessLog_Conf{},
	}
}

type MeshAccessLogConfBuilder struct {
	res *meshaccesslog_proto.MeshAccessLog_Conf
}

func (m *MeshAccessLogConfBuilder) AddFileBackend(fileBackend *meshaccesslog_proto.MeshAccessLog_FileBackend) *MeshAccessLogConfBuilder {
	m.res.Backends = append(m.res.Backends, &meshaccesslog_proto.MeshAccessLog_Backend{
		File: fileBackend,
	})
	return m
}
