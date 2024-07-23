package samples

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
)

func DataplaneInsightBackendBuilder() *builders.DataplaneInsightBuilder {
	return builders.DataplaneInsight()
}

func DataplaneInsight() *mesh.DataplaneInsightResource {
	return DataplaneInsightBackendBuilder().Build()
}
