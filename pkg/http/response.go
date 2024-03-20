package http

import "github.com/go-kratos/kratos/v2/errors"

type ICommonResponse interface {
	GetCode() int
	GetMessage() string
	GetError() error
}

type BaseResponse struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

var _ ICommonResponse = (*BaseResponse)(nil)

func (r BaseResponse) GetCode() int {
	return r.Code
}

func (r BaseResponse) GetMessage() string {
	return r.Message
}

func (r BaseResponse) GetError() error {
	if r.Code == 0 {
		return nil
	}
	var code int
	if r.Code >= 200 && r.Code < 600 {
		code = r.Code
	} else {
		code = 500
	}
	return errors.New(code, "", r.Message)
}

type CommonResponseT[T any] struct {
	BaseResponse
	Data T `json:"data,omitempty"`
}

type CommonPageResponseT[T any] struct {
	BaseResponse
	Page     int `json:"page,omitempty"`
	PageSize int `json:"page_size,omitempty"`
	Total    int `json:"total,omitempty"`
	Data     []T `json:"data,omitempty"`
}
