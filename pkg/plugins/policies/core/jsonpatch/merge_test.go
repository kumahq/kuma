package jsonpatch_test

import (
	accesslogv3 "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"
	grpcv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/grpc/v3"
	open_telemetryv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/open_telemetry/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/jsonpatch"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("Json Patch merge", func() {
	hcm := &envoy_hcm.HttpConnectionManager{
		HttpFilters: []*envoy_hcm.HttpFilter{},
		AccessLog: []*accesslogv3.AccessLog{
			{
				Name: "envoy.access_loggers.open_telemetry",
				ConfigType: &accesslogv3.AccessLog_TypedConfig{
					TypedConfig: util_proto.MustMarshalAny(&open_telemetryv3.OpenTelemetryAccessLogConfig{
						CommonConfig: &grpcv3.CommonGrpcAccessLogConfig{
							LogName: "a",
						},
					}),
				},
			},
		},
	}

	It("should merge for camelCase", func() {
		// given
		patches := []common_api.JsonPatchBlock{
			{
				Op:    "replace",
				Path:  pointer.To("/accessLog/0/typedConfig/commonConfig/logName"),
				Value: []byte(`"y"`),
			},
		}

		// when
		hcmAny := util_proto.MustMarshalAny(hcm)
		mergedHcmAny, err := jsonpatch.MergeJsonPatchAny(hcmAny, patches)

		// then
		Expect(err).ToNot(HaveOccurred())
		mergedHcm := &envoy_hcm.HttpConnectionManager{}
		err = util_proto.UnmarshalAnyTo(mergedHcmAny, mergedHcm)
		Expect(err).ToNot(HaveOccurred())
		mergedOtel := &open_telemetryv3.OpenTelemetryAccessLogConfig{}
		err = util_proto.UnmarshalAnyTo(mergedHcm.AccessLog[0].GetTypedConfig(), mergedOtel)
		Expect(err).ToNot(HaveOccurred())
		Expect(mergedOtel.CommonConfig.LogName).To(Equal("y"))
	})

	It("should not merge for snake_case", func() {
		// given
		patches := []common_api.JsonPatchBlock{
			{
				Op:    "replace",
				Path:  pointer.To("/access_log/0/typed_config/common_config/log_name"),
				Value: []byte(`"y"`),
			},
		}

		// when
		_, err := jsonpatch.MergeJsonPatchAny(util_proto.MustMarshalAny(hcm), patches)

		// then
		Expect(err).To(HaveOccurred())
	})
})
