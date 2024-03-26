package admin

import (
	"strings"

	envoy_admin_v3 "github.com/envoyproxy/go-control-plane/envoy/admin/v3"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

func Sanitize(configDump []byte) ([]byte, error) {
	toReplace := []string{}
	cd := &envoy_admin_v3.ConfigDump{}
	if err := util_proto.FromJSON(configDump, cd); err != nil {
		return nil, err
	}
	for _, config := range cd.Configs {
		if config.MessageIs(&envoy_admin_v3.BootstrapConfigDump{}) {
			bootstrapConfigDump := &envoy_admin_v3.BootstrapConfigDump{}
			if err := config.UnmarshalTo(bootstrapConfigDump); err != nil {
				return nil, err
			}

			for _, grpcService := range bootstrapConfigDump.GetBootstrap().GetDynamicResources().GetAdsConfig().GetGrpcServices() {
				for i, initMeta := range grpcService.InitialMetadata {
					if initMeta.Key == "authorization" {
						toReplace = append(toReplace, grpcService.InitialMetadata[i].Value, "[redacted]")
					}
				}
			}

			for _, grpcService := range bootstrapConfigDump.GetBootstrap().GetHdsConfig().GetGrpcServices() {
				for i, initMeta := range grpcService.InitialMetadata {
					if initMeta.Key == "authorization" {
						toReplace = append(toReplace, grpcService.InitialMetadata[i].Value, "[redacted]")
					}
				}
			}
		}
	}
	return []byte(strings.NewReplacer(toReplace...).Replace(string(configDump))), nil
}
