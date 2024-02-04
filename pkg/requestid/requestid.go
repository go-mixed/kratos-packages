package requestid

import (
	"context"
	stdLog "github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/google/uuid"
	"net/http"
	"strings"
)

const (
	HeaderXRequestID = "X-Request-ID"
)

// 当前节点的requestId
type requestId struct{}

// LogValuer 提供日志valuer
func LogValuer() stdLog.Valuer {
	return func(ctx context.Context) interface{} {
		return FromContext(ctx)
	}
}

// GenerateRequestId 获取RequestID
func GenerateRequestId() string {
	return uuid.New().String()
}

// NewContext 构造requestId上下文
//
//  1. 新的requestId会被追加到原有的requestId后面，以逗号分隔
//  2. 如果新的requestId存在于旧requestId最后一项，不会重复追加，直接返回ctx
//  3. 由于链路可能会跨越多个节点，所以在requestId中出现重复也不奇怪
//     - 比如第一项是A，第二项是B，第三项是A，表示A节点调用B节点，B节点调用A节点
//     - 但是因为2有过滤末位的原因，所以不会出现A,A,B这种情况
//  4. 因为context的取值是从当前往父级递推，所以肯定会取到最新的requestId
func NewContext(ctx context.Context, newRequestId string) context.Context {
	if newRequestId == "" {
		return ctx
	}
	originalId := FromContext(ctx)
	segments := strings.Split(originalId, ",")
	if originalId == "" {
		segments = nil
	} else if len(segments) >= 1 && segments[len(segments)-1] == newRequestId {
		return ctx
	}

	segments = append(segments, newRequestId)
	return context.WithValue(ctx, requestId{}, strings.Join(segments, ","))
}

// FromContext 从上下文获取requestId
func FromContext(ctx context.Context) string {
	value := ctx.Value(requestId{})
	if value == nil {
		return ""
	}
	return value.(string)
}

// GetFromReplyHeader 从reply header中获取requestId
func GetFromReplyHeader(ctx context.Context) string {
	if t, ok := transport.FromServerContext(ctx); ok {
		return t.ReplyHeader().Get(HeaderXRequestID)
	}
	return ""
}

// GetOrGenerateRequestId 从header中获取requestId，如果没有则生成一个
func GetOrGenerateRequestId(r *http.Request) string {
	id := uuid.New().String()
	if r != nil {
		if _id := r.Header.Get(HeaderXRequestID); _id != "" {
			id = _id
		}
	}

	return id
}
