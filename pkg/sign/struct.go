package sign

import (
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/utils"
	"strings"
	"time"
)

const defaultTimestampUnit = time.Millisecond   // 默认时间戳单位，支持：time.Second, time.Millisecond, time.Microsecond, time.Nanosecond
const defaultSignedFields = "app_key,timestamp" // 默认签名字段

// BaseSignature 签名基础结构，继承此结构体，可实现签名功能
//
//	type A struct {
//		BaseSignature
//		B string `json:"b"`
//		C string `json:"c"`
//	}
//	a := &A{}
//	a.BuildSignatures(a, "AppKey", "Secret", nil, false)
//	a.CheckSignatures(a, "Secret", false)
type BaseSignature struct {
	// 以下字段为签名字段
	AppKey    string `json:"app_key"`
	Timestamp int64  `json:"timestamp"`
	Sign      string `json:"sign"`

	timestampUnit time.Duration
}

var _ iStructSignature = (*BaseSignature)(nil)

func (s *BaseSignature) SetAppKey(appKey string) {
	s.AppKey = appKey
}

func (s *BaseSignature) GetAppKey() string {
	return s.AppKey
}
func (s *BaseSignature) SetTimestamp(now time.Time) {
	if !now.IsZero() {
		now = time.Now()
	}

	var timestampUnit = s.timestampUnit

switch1:
	switch timestampUnit {
	case time.Millisecond:
		s.Timestamp = now.UnixMilli()
	case time.Microsecond:
		s.Timestamp = now.UnixMicro()
	case time.Nanosecond:
		s.Timestamp = now.UnixNano()
	case time.Second:
		s.Timestamp = now.Unix()
	default:
		timestampUnit = defaultTimestampUnit
		goto switch1
	}
}

func (s *BaseSignature) GetTimestamp() time.Time {
	if s.Timestamp <= 0 {
		return time.Time{}
	}

	var timestampUnit = s.timestampUnit

switch1:
	switch timestampUnit {
	case time.Millisecond:
		return time.UnixMilli(s.Timestamp)
	case time.Microsecond:
		return time.UnixMicro(s.Timestamp)
	case time.Nanosecond:
		return time.Unix(s.Timestamp/1e9, s.Timestamp%1e9)
	case time.Second:
		return time.Unix(s.Timestamp, 0)
	default:
		timestampUnit = defaultTimestampUnit
		goto switch1
	}
}

func (s *BaseSignature) SetTimestampUnit(unit time.Duration) {
	s.timestampUnit = unit
}

func (s *BaseSignature) SetSign(sign string) {
	s.Sign = sign
}

func (s *BaseSignature) GetSign() string {
	return s.Sign
}

// BuildSignature 传入一个 BaseSignature 的子struct，生成签名并返回
//
//	obj: BaseSignature的子struct
//	appKey: 应用的appKey
//	appSecret: 应用的appSecret
//	signedFields: 参与签名的字段，如果为空，所有字段都参与签名。注意：字段名需要是Struct中字段tag名，即json:"xxx"中的xxx
//	withBlank: 是否包含空白字段，如果为false，值为空白字符串的不参与签名
func (s *BaseSignature) BuildSignature(obj iStructSignature, appKey, appSecret string, options Options) {
	obj.SetAppKey(appKey)
	obj.SetTimestamp(time.Now())

	values := utils.AnyToUrlValues(obj, "json")
	delete(values, "sign")
	if len(options.signedFields) > 0 {
		options.signedFields = append(options.signedFields, strings.Split(defaultSignedFields, ",")...)
	}

	obj.SetSign(CalcSignature(appSecret, values, options))
}

// CheckSignature 传入一个 BaseSignature 的子struct，检查签名是否正确
//
//	obj: BaseSignature的子struct
//	appSecret: 应用的appSecret
//	signedFields: 参与签名的字段，如果为空，所有字段都参与签名。注意：字段名需要是Struct中字段tag名，即json:"xxx"中的xxx
//	withBlank: 是否包含空白字段，如果为false，值为空白字符串的不参与签名
func (s *BaseSignature) CheckSignature(obj iStructSignature, appSecret string, options Options) (bool, error) {
	return CheckSignature(obj, appSecret, options)
}
