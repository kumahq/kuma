package samples

import (
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
)

func DataplaneInsightBackendBuilder() *builders.DataplaneInsightBuilder {
	return builders.DataplaneInsight()
}

func DataplaneInsight() *mesh.DataplaneInsightResource {
	return DataplaneInsightBackendBuilder().Build()
}
