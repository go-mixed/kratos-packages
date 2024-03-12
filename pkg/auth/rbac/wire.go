package rbac

import "github.com/google/wire"

var RBACProviderSet = wire.NewSet(
	NewCasbin,
)
