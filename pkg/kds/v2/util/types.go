package util

import (
	std_errors "errors"

	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
)

type UpstreamResponse struct {
	ControlPlaneId      string
	Type                core_model.ResourceType
	AddedResources      core_model.ResourceList
	InvalidResourcesKey []core_model.ResourceKey
	RemovedResourcesKey []core_model.ResourceKey
	IsInitialRequest    bool
}

func (u *UpstreamResponse) Validate() error {
	if u.AddedResources == nil {
		return nil
	}
	var err error
	for _, res := range u.AddedResources.GetItems() {
		if validationErr := core_model.Validate(res); validationErr != nil {
			err = std_errors.Join(err, validationErr)
			u.InvalidResourcesKey = append(u.InvalidResourcesKey, core_model.MetaToResourceKey(res.GetMeta()))
		}
	}
	return err
}

type Callbacks struct {
	OnResourcesReceived func(upstream UpstreamResponse) (error, error)
}
