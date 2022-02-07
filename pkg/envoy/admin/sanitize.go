package admin

import envoy_admin_v3 "github.com/envoyproxy/go-control-plane/envoy/admin/v3"

func Sanitize(configDump *envoy_admin_v3.ConfigDump) error {
	for _, config := range configDump.Configs {
		if config.MessageIs(&envoy_admin_v3.BootstrapConfigDump{}) {
			bootstrapConfigDump := &envoy_admin_v3.BootstrapConfigDump{}
			if err := config.UnmarshalTo(bootstrapConfigDump); err != nil {
				return err
			}

			for _, grpcService := range bootstrapConfigDump.GetBootstrap().GetDynamicResources().GetAdsConfig().GetGrpcServices() {
				for i, initMeta := range grpcService.InitialMetadata {
					if initMeta.Key == "authorization" {
						grpcService.InitialMetadata[i].Value = "[redacted]"
					}
				}
			}

			for _, grpcService := range bootstrapConfigDump.GetBootstrap().GetHdsConfig().GetGrpcServices() {
				for i, initMeta := range grpcService.InitialMetadata {
					if initMeta.Key == "authorization" {
						grpcService.InitialMetadata[i].Value = "[redacted]"
					}
				}
			}

			if err := config.MarshalFrom(bootstrapConfigDump); err != nil {
				return err
			}
		}
	}
	return nil
}
