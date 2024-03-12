package rbac

import (
	_ "embed"
	"fmt"
	"github.com/casbin/casbin/v2"
	casbinModel "github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/samber/lo"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/auth"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/log"
	"gorm.io/gorm"
)

// 将rbac_model.conf中的内容绑定到golang的资源文件
//
//go:embed rbac_model.conf
var rbacModel string

type Casbin struct {
	*casbin.Enforcer

	logger *log.Helper
}

// NewCasbin 启用casbin
func NewCasbin(
	db *gorm.DB,
	logger log.Logger,
) *Casbin {
	gormadapter.TurnOffAutoMigrate(db)
	// 关闭了自动迁移，下面函数不会返回错误
	adapter, _ := gormadapter.NewFilteredAdapterByDB(db, "", (&PolicyModel{}).TableName())

	model, err := casbinModel.NewModelFromString(rbacModel)
	if err != nil {
		panic("casbin rbac_model.conf error")
	}
	// 使用NewFilteredAdapterByDB表示自己loadPolicy，所以NewEnforcer方法不会返回错误
	e, _ := casbin.NewEnforcer(model, adapter)

	e.SetLogger(&casbinLogger{
		logger: log.NewModuleHelper(logger.Clone().AddStack(1), "rbac/casbin"),
	})

	return &Casbin{
		Enforcer: e,
		logger:   log.NewModuleHelper(logger, "rbac"),
	}
}

func (c *Casbin) getUserName(guard auth.IGuard) string {
	return fmt.Sprintf("%s:%d", guard.GetGuardName(), guard.GetAuthorizationID())
}

// AttachRolesForUser 为用户附加角色(会先删除用户的所有角色)
func (c *Casbin) AttachRolesForUser(guard auth.IGuard, roles ...string) error {
	sub := c.getUserName(guard)

	_, _ = c.DeleteRolesForUser(sub)
	_, err := c.AddRolesForUser(sub, roles)
	return err
}

// DetachRolesForUser 为用户删除角色，如果没有指定角色，则删除所有角色
func (c *Casbin) DetachRolesForUser(guard auth.IGuard, roles ...string) error {
	sub := c.getUserName(guard)

	if len(roles) == 0 {
		_, _ = c.DeleteRolesForUser(sub)
	}
	for _, role := range roles {
		_, _ = c.DeleteRoleForUser(sub, role)
	}
	return nil
}

// ClearRolePolicies 清除角色的所有策略，但是不会清除用户的角色
func (c *Casbin) ClearRolePolicies(role string) error {
	_, err := c.RemoveFilteredPolicy(0, role)
	return err
}

// ClearUserPolicies 清除用户的所有策略
func (c *Casbin) ClearUserPolicies(guard auth.IGuard) error {
	sub := c.getUserName(guard)
	_, err := c.RemoveFilteredPolicy(0, sub)
	return err
}

// GetUserRoles 获取用户的角色，如果有多个用户，则返回所有用户的角色（去重）
func (c *Casbin) GetUserRoles(guards ...auth.IGuard) ([]string, error) {
	if err := c.LoadUserPolicies(guards...); err != nil {
		return nil, err
	}

	var roles []string
	for _, user := range lo.Map(guards, func(guard auth.IGuard, _ int) string {
		return c.getUserName(guard)
	}) {
		r, err := c.GetRolesForUser(user)
		if err != nil {
			return nil, err
		}
		roles = append(roles, r...)
	}

	return lo.Uniq(roles), nil
}

// LoadRolePolicies 加载角色的策略。注意：每次加载都会清理所有角色的策略，然后按下面的条件加载用户组的策略
func (c *Casbin) LoadRolePolicies(roles ...string) error {
	return c.LoadFilteredPolicy(gormadapter.Filter{Ptype: []string{"p"}, V0: roles})
}

// LoadUserPolicies 加载用户（以及用户所属的角色）的策略。注意：每次加载都会清理所有用户、角色的策略，然后按下面的条件加载用户、角色的策略
func (c *Casbin) LoadUserPolicies(guards ...auth.IGuard) error {
	users := lo.Map(guards, func(guard auth.IGuard, _ int) string {
		return c.getUserName(guard)
	})

	// 先查询用户的角色
	if err := c.LoadFilteredPolicy(gormadapter.Filter{Ptype: []string{"g"}, V0: users}); err != nil {
		return err
	}

	roles, err := c.GetUserRoles(guards...)
	if err != nil {
		return err
	}

	return c.LoadFilteredPolicy([]gormadapter.Filter{
		{Ptype: []string{"p"}, V0: roles},
		{Ptype: []string{"g", "p"}, V0: users},
	})
}

// EnforceUser 检查用户是否有权限
func (c *Casbin) EnforceUser(guard auth.IGuard, obj string, act string) (bool, error) {
	if err := c.LoadUserPolicies(guard); err != nil {
		return false, err
	}

	return c.Enforce(c.getUserName(guard), obj, act)
}

// EnforceRole 检查角色是否有权限
func (c *Casbin) EnforceRole(role string, obj string, act string) (bool, error) {
	if err := c.Enforcer.LoadFilteredPolicy(gormadapter.Filter{Ptype: []string{"p"}, V0: []string{role}}); err != nil {
		return false, err
	}
	return c.Enforce(role, obj, act)
}
