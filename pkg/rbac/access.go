package rbac

import (
	resources_rbac "github.com/kumahq/kuma/pkg/core/resources/rbac"
	tokens_rbac "github.com/kumahq/kuma/pkg/tokens/builtin/rbac"
)

type RBACAccess struct {
	ResourceAccess               resources_rbac.ResourceAccess
	GenerateDataplaneTokenAccess tokens_rbac.GenerateDataplaneTokenAccess
}
