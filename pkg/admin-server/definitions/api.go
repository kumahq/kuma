package definitions

import (
	"github.com/pkg/errors"

	system "github.com/Kong/kuma/pkg/core/resources/apis/system"
	"github.com/Kong/kuma/pkg/core/resources/model"
	core_rest "github.com/Kong/kuma/pkg/core/resources/model/rest"
)

func AllApis() core_rest.Api {
	return &adminApis{}
}

var _ core_rest.Api = &adminApis{}

type adminApis struct {
}

func (a *adminApis) GetResourceApi(typ model.ResourceType) (core_rest.ResourceApi, error) {
	if typ == system.SecretType {
		return core_rest.NewResourceApi(typ, "secrets"), nil
	}
	return nil, errors.Errorf("unknown resource type: %q", typ)
}
