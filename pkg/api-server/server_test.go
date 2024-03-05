package api_server_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	api_server "github.com/kumahq/kuma/pkg/api-server"
	config_api_server "github.com/kumahq/kuma/pkg/config/api-server"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/config/multizone"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

var _ = Describe("Resource", func() {
	type testCase struct {
		kdsFlag       model.KDSFlagType
		mode          core.CpMode
		federatedZone bool
		isApiReadOnly bool
		expected      bool
	}
	DescribeTable("on Global", func(given testCase) {
		// given
		cfg := kuma_cp.Config{
			Mode: given.mode,
			ApiServer: &config_api_server.ApiServerConfig{
				ReadOnly: given.isApiReadOnly,
			},
			Multizone: &multizone.MultizoneConfig{
				Zone: &multizone.ZoneConfig{},
			},
		}

		// then
		Expect(api_server.ShouldBeReadOnly(given.kdsFlag, &cfg)).To(Equal(given.expected))
	},
		Entry("shouldn't be read only when kds from global to zone and api is not read only", testCase{
			kdsFlag:       model.GlobalToAllZonesFlag,
			mode:          core.Global,
			isApiReadOnly: false,
			expected:      false,
		}),
		Entry("should be read only when kds from global to zone and api is read only", testCase{
			kdsFlag:       model.GlobalToAllZonesFlag,
			mode:          core.Global,
			isApiReadOnly: true,
			expected:      true,
		}),
		Entry("should be read only when kds from zone to global and api is not read only", testCase{
			kdsFlag:       model.ZoneToGlobalFlag,
			mode:          core.Global,
			isApiReadOnly: false,
			expected:      true,
		}),
		Entry("should be read only when kds from zone to global and api is read only", testCase{
			kdsFlag:       model.GlobalToAllZonesFlag,
			mode:          core.Global,
			isApiReadOnly: true,
			expected:      true,
		}),
		Entry("should be read only when there is no kds and api is read only", testCase{
			kdsFlag:       0,
			mode:          core.Global,
			isApiReadOnly: true,
			expected:      true,
		}),
		Entry("shouldn't be read only when there is no kds and api is not read only", testCase{
			kdsFlag:       0,
			mode:          core.Global,
			isApiReadOnly: false,
			expected:      false,
		}),
		Entry("shouldn't be read only when there are both kds types and api is not read only", testCase{
			kdsFlag:       model.ZoneToGlobalFlag | model.GlobalToAllZonesFlag,
			mode:          core.Global,
			isApiReadOnly: false,
			expected:      false,
		}),
		Entry("should be read only when there are both kds types and api is read only", testCase{
			kdsFlag:       model.ZoneToGlobalFlag | model.GlobalToAllZonesFlag,
			mode:          core.Global,
			isApiReadOnly: true,
			expected:      true,
		}),
	)

	DescribeTable("on Zone", func(given testCase) {
		// given
		cfg := kuma_cp.Config{
			Mode: given.mode,
			ApiServer: &config_api_server.ApiServerConfig{
				ReadOnly: given.isApiReadOnly,
			},
			Multizone: &multizone.MultizoneConfig{
				Zone: &multizone.ZoneConfig{},
			},
		}
		if given.federatedZone {
			cfg.Multizone.Zone.Name = "x"
			cfg.Multizone.Zone.GlobalAddress = "grpcs://global:5685"
		}

		// then
		Expect(api_server.ShouldBeReadOnly(given.kdsFlag, &cfg)).To(Equal(given.expected))
	},
		Entry("should be read only when kds from global to zone and api is read only", testCase{
			kdsFlag:       model.GlobalToAllZonesFlag,
			mode:          core.Zone,
			isApiReadOnly: true,
			expected:      true,
		}),
		Entry("should be read only when kds from global to zone and api is not read only and zone is federated", testCase{
			kdsFlag:       model.GlobalToAllZonesFlag,
			mode:          core.Zone,
			isApiReadOnly: false,
			federatedZone: true,
			expected:      true,
		}),
		Entry("shouldn't be read only when kds from global to zone and api is not read only and zone is not federated", testCase{
			kdsFlag:       model.GlobalToAllZonesFlag,
			mode:          core.Zone,
			isApiReadOnly: false,
			federatedZone: false,
			expected:      false,
		}),
		Entry("should be read only when kds from zone to global and api is read only", testCase{
			kdsFlag:       model.ZoneToGlobalFlag,
			mode:          core.Zone,
			isApiReadOnly: true,
			expected:      true,
		}),
		Entry("should be read only when kds from zone to global and api is not read only", testCase{
			kdsFlag:       model.ZoneToGlobalFlag,
			mode:          core.Zone,
			isApiReadOnly: false,
			expected:      false,
		}),
		Entry("shouldn't be read only when there is no kds and api is not read only", testCase{
			kdsFlag:       0,
			mode:          core.Zone,
			isApiReadOnly: false,
			expected:      false,
		}),
		Entry("should be read only when there is no kds and api is read only", testCase{
			kdsFlag:       0,
			mode:          core.Zone,
			isApiReadOnly: true,
			expected:      true,
		}),
		Entry("shouldn't be read only when there are both kds types and api is not read only", testCase{
			kdsFlag:       model.ZoneToGlobalFlag | model.GlobalToAllZonesFlag,
			mode:          core.Zone,
			isApiReadOnly: false,
			expected:      false,
		}),
		Entry("should be read only when there are both kds types and api is read only", testCase{
			kdsFlag:       model.ZoneToGlobalFlag | model.GlobalToAllZonesFlag,
			mode:          core.Zone,
			isApiReadOnly: true,
			expected:      true,
		}),
	)
})
