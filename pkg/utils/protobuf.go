package utils

import "google.golang.org/protobuf/reflect/protoreflect"

type IProtobuf interface {
	ProtoReflect() protoreflect.Message
}
