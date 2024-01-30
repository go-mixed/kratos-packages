package sign

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"time"
)

type ISignatureGetter interface {
	GetAppKey() string
	GetSign() string
	GetTimestamp() time.Time
}

type ISignatureSetter interface {
	SetAppKey(appKey string)
	SetSign(sign string)
	SetTimestamp(now time.Time)
}

type iStructSignature interface {
	ISignatureSetter
	SetTimestampUnit(unit time.Duration)
	ISignatureGetter
}

type IProtobufSignature interface {
	GetAppKey() string
	GetSign() string
	GetTimestamp() int64
	ProtoReflect() protoreflect.Message
}
