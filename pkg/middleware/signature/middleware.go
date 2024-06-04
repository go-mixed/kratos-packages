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

type signatureMiddlewareFunc func(
	ctx context.Context,
	transporter transport.Transporter,
	request sign.IProtobufSignature) (thirdPartyApp auth.IThirdParty, options []sign.Option, err error)

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
				thirdPartyApp, opts, err := signatureFunc(ctx, transporter, request)
				if err != nil {
					return nil, err
				}

				// 传递第三方app信息
				ctx = sign.NewContext(ctx, thirdPartyApp)
				// 将日志传递给签名中间件
				opts = append(opts, sign.WithLogger(l))

				if err = sign.CheckProtobufSignature(
					request,
					thirdPartyApp.GetAppSecret(),
					opts...,
				); err != nil {
					return nil, err
				}

				return nextHandler(ctx, req)
			}
			return nil, errors.BadRequest("signature", "Please input sign(string)、timestamp(int)、app_key(string)")
		}
	}
}
