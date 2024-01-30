package signature

import (
	"context"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/auth"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/log"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/sign"
)

type signatureMiddlewareFunc func(ctx context.Context, transporter transport.Transporter, request sign.IProtobufSignature) (auth.IThirdParty, error)

func NewSignatureMiddleware(signatureFunc signatureMiddlewareFunc, logger log.Logger) middleware.Middleware {
	return func(nextHandler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			l := log.NewModuleHelper(logger, "middleware/signature").WithContext(ctx)
			transporter, ok := transport.FromServerContext(ctx)
			if !ok {
				l.Error("wrong transport context for signature middleware")
				return nil, auth.ErrWrongContext
			}
			if request, ok := req.(sign.IProtobufSignature); ok {
				thirdPartyApp, err := signatureFunc(ctx, transporter, request)
				if err != nil {
					return nil, err
				}

				// 传递第三方app信息
				ctx = sign.NewContext(ctx, thirdPartyApp)

				if ok, err = sign.CheckProtobufSignature(
					request,
					thirdPartyApp.GetAppSecret(),
					sign.DefaultOptions().WithLogger(l).WithValidateTimestamp(validateTimestamp(thirdPartyApp)),
				); ok {
					return nextHandler(ctx, req)
				} else if err != nil {
					return nil, err
				}
			}
			return nil, errors.BadRequest("signature", "请传递正确的签名参数")
		}
	}
}

func validateTimestamp(thirdPartyApp auth.IThirdParty) bool {
	switch thirdPartyApp.GetAppKey() {

	}

	return true
}