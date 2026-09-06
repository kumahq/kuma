package mappers

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
)

func TestMapResourceTypeDescriptionPreservesRulesTargetRefPolicies(t *testing.T) {
	t.Parallel()

	response := MapResourceTypeDescription([]model.ResourceTypeDescriptor{
		{
			Name:              "MeshCircuitBreaker",
			WsPath:            "meshcircuitbreakers",
			IsPolicy:          true,
			IsTargetRefBased:  true,
			HasToTargetRef:    true,
			HasRulesTargetRef: true,
		},
	}, false, false, false)

	require.Len(t, response.Resources, 1)
	require.NotNil(t, response.Resources[0].Policy)
	require.False(t, response.Resources[0].Policy.HasFromTargetRef)
	require.False(t, response.Resources[0].Policy.IsFromAsRules)
	require.True(t, response.Resources[0].Policy.HasRulesTargetRef)
	require.True(t, response.Resources[0].Policy.HasToTargetRef)
}

func TestMapResourceTypeDescriptionReadOnly(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		readOnly      bool
		isGlobal      bool
		federatedZone bool
		descriptor    model.ResourceTypeDescriptor
		expected      bool
	}{
		{
			name:       "standalone Zone can write resources without KDS ownership",
			descriptor: model.ResourceTypeDescriptor{},
		},
		{
			name:     "Global can write resources provided by Global",
			isGlobal: true,
			descriptor: model.ResourceTypeDescriptor{
				KDSFlags: model.GlobalToZonesFlag,
			},
		},
		{
			name:     "Global cannot write resources provided by Zone",
			isGlobal: true,
			descriptor: model.ResourceTypeDescriptor{
				KDSFlags: model.ZoneToGlobalFlag,
			},
			expected: true,
		},
		{
			name:          "federated Zone can write resources provided by Zone",
			federatedZone: true,
			descriptor: model.ResourceTypeDescriptor{
				KDSFlags: model.ZoneToGlobalFlag,
			},
		},
		{
			name:          "federated Zone cannot write resources provided by Global",
			federatedZone: true,
			descriptor: model.ResourceTypeDescriptor{
				KDSFlags: model.GlobalToZonesFlag,
			},
			expected: true,
		},
		{
			name: "descriptor read-only wins",
			descriptor: model.ResourceTypeDescriptor{
				ReadOnly: true,
			},
			expected: true,
		},
		{
			name:     "API read-only wins",
			readOnly: true,
			descriptor: model.ResourceTypeDescriptor{
				KDSFlags: model.GlobalToZonesFlag | model.ZoneToGlobalFlag,
			},
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			test.descriptor.Name = "TestResource"

			response := MapResourceTypeDescription(
				[]model.ResourceTypeDescriptor{test.descriptor},
				test.readOnly,
				test.isGlobal,
				test.federatedZone,
			)

			require.Len(t, response.Resources, 1)
			require.Equal(t, test.expected, response.Resources[0].ReadOnly)
		})
	}
}

func TestMapResourceTypeDescriptionFederationIsIndependentFromReadOnly(t *testing.T) {
	t.Parallel()

	descriptor := model.ResourceTypeDescriptor{
		Name:     "GlobalProvidedResource",
		KDSFlags: model.GlobalToZonesFlag,
	}

	response := MapResourceTypeDescription([]model.ResourceTypeDescriptor{descriptor}, true, false, true)

	require.Len(t, response.Resources, 1)
	require.True(t, response.Resources[0].ReadOnly)
	require.True(t, response.Resources[0].IncludeInFederation)
}
