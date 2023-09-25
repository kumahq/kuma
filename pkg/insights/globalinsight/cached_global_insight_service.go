package globalinsight

import (
	"context"
	"time"

	"github.com/patrickmn/go-cache"

	api_types "github.com/kumahq/kuma/api/openapi/types"
	"github.com/kumahq/kuma/pkg/multitenant"
)

type cachedGlobalInsightService struct {
	globalInsightService GlobalInsightService
	cache                *cache.Cache
	tenants              multitenant.Tenants
}

var _ GlobalInsightService = &cachedGlobalInsightService{}

func NewCachedGlobalInsightService(globalInsightService GlobalInsightService, tenants multitenant.Tenants, expirationTime time.Duration) GlobalInsightService {
	return &cachedGlobalInsightService{
		globalInsightService: globalInsightService,
		cache:                cache.New(expirationTime, time.Duration(int64(float64(expirationTime)*0.9))),
		tenants:              tenants,
	}
}

func (gis *cachedGlobalInsightService) GetGlobalInsight(ctx context.Context) (*api_types.GlobalInsight, error) {
	tenantID, err := gis.tenants.GetID(ctx)
	if err != nil {
		return nil, err
	}

	obj, found := gis.cache.Get(tenantID)
	if !found {
		insight, err := gis.globalInsightService.GetGlobalInsight(ctx)
		if err != nil {
			return nil, err
		}
		gis.cache.SetDefault(tenantID, insight)
		return insight, nil
	}

	return obj.(*api_types.GlobalInsight), nil
}
