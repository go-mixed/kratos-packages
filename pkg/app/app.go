package app

import (
	"context"
	"github.com/go-kratos/kratos/v2"
	"github.com/google/uuid"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/auth"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/sign"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/trace"
)

// Version 修改版本执行： go build -ldflags "-X gopkg.in/go-mixed/kratos-packages.v2/pkg/app.Version=x.y.z"
var Version string = ""

var created bool

// App
// Kratos也有这些参数，但是必须在Run时，ctx才会传入相关Server.Start中。而很多函数在初始化阶段就需要ctx。
// 以及ctx有重建的情况（比如WS的Pub/Sub和context序列化），会丢失*kratos.App的ctx包裹
// 所以新建一个全局唯一struct，可以随处调用
type App struct {
	id       string
	name     string
	version  string
	metadata map[string]string
	endpoint []string
	ctx      context.Context
}

var _ kratos.AppInfo = (*App)(nil)

// NewApp 实例化app。只能实例化一次，否则会panic
func NewApp(
	name string,

) *App {
	if created {
		panic("app can only be created once")
	}
	created = true
	app := &App{
		id:      uuid.New().String(),
		name:    name,
		version: Version,
	}
	app.ctx = kratos.NewContext(context.Background(), app)
	return app
}

func (a *App) ID() string {
	return a.id
}

func (a *App) Name() string {
	return a.name
}

func (a *App) Version() string {
	return a.version
}

func (a *App) Metadata() map[string]string {
	return a.metadata
}

func (a *App) Endpoint() []string {
	return a.endpoint
}

// ReplaceContext 仅在main.go的kratos.BeforeStart时调用
// 其它情况下，请勿调用
func (a *App) ReplaceContext(ctx context.Context) {
	a.ctx = ctx
}

// CloneContextFromBase 基于当前BaseContext生成一个新的context，并提取fromCtx的trace信息，赋值到新context中，然后返回。
// 由于大部分ctx都来源于request，当http request中生命周期束时，ctx会被取消。
// 这可能会导致一些问题，比如：在新的协程中操作redis/mysql的操作
// 所以，为了避免这种情况，需要创建一个新的context，用于新协程中的各项操作
//
//	新context包含：app、kratos、trace、auth、sign
func (a *App) CloneContextFromBase(fromCtx context.Context) context.Context {
	var ctx context.Context = a.ctx
	// ReplaceContext方法已经将BaseContext替换为kratos的context，所以这里不需要再赋值
	//// 尝试从fromCtx中获取kratos
	//if k, ok := kratos.FromContext(fromCtx); ok {
	//	ctx = kratos.NewContext(ctx, k)
	//}
	// 尝试从fromCtx中获取trace
	ctx = trace.NewContext(ctx, trace.FromContext(fromCtx))
	// 尝试从fromCtx中获取auth
	if _auth, ok := auth.FromContext(fromCtx); ok {
		ctx = auth.NewContext(ctx, _auth)
	}
	// 尝试从fromCtx中获取sign
	if _sign, ok := sign.FromContext(fromCtx); ok {
		ctx = sign.NewContext(ctx, _sign)
	}

	return ctx
}

// BaseContext 获取app的base context
func (a *App) BaseContext() context.Context {
	return a.ctx
}
