package utils

import (
	"context"
	"encoding/json"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/requestid"
)

type transportContext struct {
	RequestID string `json:"request_id"`
}

// SerializeContext 将本项目的context序列化为json字符串（仅仅用于websocket、http传输）
func SerializeContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	wsContext := transportContext{
		RequestID: requestid.FromContext(ctx),
	}

	j, _ := json.Marshal(wsContext)
	return string(j)
}

// DeserializeContext 将json字符串反序列化为本项目的context（仅仅用于websocket、http传输）
func DeserializeContext(parentCtx context.Context, ctxStr string) context.Context {
	var wsContext transportContext
	_ = json.Unmarshal([]byte(ctxStr), &wsContext)

	// 包裹父级parentCtx，保证日志的链路跟踪
	ctx := requestid.NewContext(parentCtx, wsContext.RequestID)
	return ctx
}
