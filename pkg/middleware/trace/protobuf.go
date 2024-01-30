package trace

import (
	"context"
	"github.com/go-kratos/kratos/v2"
	"google.golang.org/protobuf/reflect/protoreflect"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/utils"
)

func fillReqIdAppId(ctx context.Context, reply any, reqId string) {
	// 调用protobuf原生方法，将traceId和appId添加到返回值中
	if reply != nil {
		if response, ok := reply.(utils.IProtobuf); ok {
			messageDesc := response.ProtoReflect().Descriptor()
			// fill traceId
			traceField := messageDesc.Fields().ByName("trace")
			if traceField != nil && traceField.Kind() == protoreflect.StringKind && response.ProtoReflect().Get(traceField).String() == "" {
				response.ProtoReflect().Set(traceField, protoreflect.ValueOfString(reqId))
			}

			// fill appId
			if appImpl, ok := kratos.FromContext(ctx); ok {
				appField := messageDesc.Fields().ByName("app_id")
				if appField != nil && appField.Kind() == protoreflect.StringKind && response.ProtoReflect().Get(appField).String() == "" {
					response.ProtoReflect().Set(appField, protoreflect.ValueOfString(appImpl.ID()))
				}
			}
		}
	}
}
