package sign

import (
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/utils"
	"math"
	"net/url"
	"sort"
	"strings"
	"time"
)

// BuildSignature 传入一个 BaseSignature 的子struct，生成签名并返回
//
//	obj: BaseSignature的子struct
//	appKey: 应用的appKey
//	appSecret: 应用的appSecret
//	opts:
//	- WitSignedFields: 参与签名的字段，如果为空，所有字段都参与签名。注意：字段名需要是Struct中字段tag名，即json:"xxx"中的xxx
//	- WithBlank: 是否包含空白字段，如果为false，值为空白字符串的不参与签名
func BuildSignature(obj iStructSignature, appKey, appSecret string, opts ...Option) url.Values {
	obj.SetAppKey(appKey)
	obj.SetTimestamp(time.Now())

	options := defaultOptions().apply(opts...)

	values := utils.AnyToUrlValues(obj, "json")
	delete(values, "sign")
	if len(options.signedFields) > 0 {
		options.signedFields = append(options.signedFields, strings.Split(defaultSignedFields, ",")...)
	}

	obj.SetSign(CalcSignature(appSecret, values, opts...))
	values.Set("sign", obj.GetSign())
	return values
}

// CalcSignature 输入values（url.Values）的数据，计算为签名，如果withBlank为false，值为空白字符串的不参与签名
//
//	appSecret: 应用的appSecret
//	values: 签名的数据
//	opts:
//	- WitSignedFields: 参与签名的字段，如果为空，所有字段都参与签名。注意：字段名需要是Struct中字段tag名，即json:"xxx"中的xxx
//	- WithBlank: 是否包含空白字段，如果为false，值为空白字符串的不参与签名
func CalcSignature(appSecret string, values url.Values, opts ...Option) string {
	options := defaultOptions().apply(opts...)

	// 从values中获取keys
	keys := lo.Keys(values)

	// 筛选signedFields中的字段；signedFields为空，所有字段都参与签名
	if len(options.signedFields) > 0 {
		keys = lo.Intersect(keys, options.signedFields)
	}

	// 如果withBlank为false，值为空白字符串的不参与签名
	if !options.includeBlank {
		keys = lo.Filter(keys, func(key string, _ int) bool {
			return values.Get(key) != ""
		})
	}

	sort.Strings(keys) // key的正序排序

	//"k1=v1&k2=v2" + secret
	content := strings.Join(lo.Map(keys, func(key string, _ int) string {
		return key + "=" + values.Get(key)
	}), "&") + appSecret

	res := utils.MD5String(content)

	if options.logger != nil {
		options.logger.Debugf("signature: %s, query: %+v, signature string %s", res, values, content)
	}

	return res
}

// checkSignature 传入appKey、appSecret、sign、timestamp、签名字段，验证签名是否正确
//
//	obj: 支持struct、map、url.Values、querystring
//	appKey: 应用的appKey
//	appSecret: 应用的appSecret
//	sign: 签名
//	opts:
//	 - WithTimestamp: 时间戳
//	 - WithSignedFields: 参与签名的字段，如果为空，所有字段都参与签名。注意：字段名需要是Struct中字段tag名，即json:"xxx"中的xxx
//	 - WithBlank: 是否包含空白字段，如果为false，值为空白字符串的不参与签名
//	 - WithLogger: log.Helper，如果不为空，会打印签名的详细信息
func checkSignature(obj any, appSecret, sign string, timestamp time.Time, opts ...Option) error {
	options := defaultOptions().apply(opts...)

	if sign == "" {
		return errors.New("signature is empty")
	}

	if options.validateTimestamp && math.Abs(time.Since(timestamp).Seconds()) >= 5.*60. {
		return errors.Errorf("signature is expired(must <=5min): %s. query: %+v", timestamp, obj)
	}

	values := utils.AnyToUrlValues(obj, "json")
	delete(values, "sign")

	if len(options.signedFields) > 0 {
		options.signedFields = append(options.signedFields, strings.Split(defaultSignedFields, ",")...)
	}

	actualSign := CalcSignature(appSecret, values, opts...)
	if strings.EqualFold(actualSign, sign) {
		return nil
	}
	if options.logger != nil {
		options.logger.Infof("sign failed, actual sign = %s, obj = %+v", sign, obj)
	}
	return errors.Errorf("signature is invalid: %s", sign)
}

// CheckSignature 传入一个 BaseSignature 的子struct，验证签名是否正确
//
//	obj: ISignatureGetter的struct
//	appSecret: 应用的appSecret
//	opts:
//	- WithSignedFields: 参与签名的字段，如果为空，所有字段都参与签名。注意：字段名需要是Struct中字段tag名，即json:"xxx"中的xxx
//	- WithBlank: 是否包含空白字段，如果为false，值为空白字符串的不参与签名
//	- WithLogger: log.Helper，如果不为空，会打印签名的详细信息
func CheckSignature(obj ISignatureGetter, appSecret string, opts ...Option) error {
	sign := obj.GetSign()
	timestamp := obj.GetTimestamp()

	return checkSignature(obj, appSecret, sign, timestamp, opts...)
}

// CheckProtobufSignature 传入一个 BaseSignature 的子struct，验证签名是否正确
//
//	obj: IProtobufSignature的struct
//	appSecret: 应用的appSecret
//	opts:
//	- WithSignedFields: 参与签名的字段，如果为空，所有字段都参与签名。注意：字段名需要是Struct中字段tag名，即json:"xxx"中的xxx
//	- WithBlank: 是否包含空白字段，如果为false，值为空白字符串的不参与签名
//	- WithLogger: log.Helper，如果不为空，会打印签名的详细信息
func CheckProtobufSignature(obj IProtobufSignature, appSecret string, opts ...Option) error {
	sign := obj.GetSign()
	timestamp := time.UnixMilli(obj.GetTimestamp())

	return checkSignature(obj, appSecret, sign, timestamp, opts...)
}
