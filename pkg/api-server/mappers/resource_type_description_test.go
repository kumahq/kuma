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
			HasFromTargetRef:  false,
			HasRulesTargetRef: true,
		},
	}, false, false)

	require.Len(t, response.Resources, 1)
	require.NotNil(t, response.Resources[0].Policy)
	require.False(t, response.Resources[0].Policy.HasFromTargetRef)
	require.True(t, response.Resources[0].Policy.HasRulesTargetRef)
	require.True(t, response.Resources[0].Policy.HasToTargetRef)
}
