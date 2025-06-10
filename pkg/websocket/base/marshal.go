package base

import (
	"github.com/go-kratos/kratos/v2/encoding/json"
	"google.golang.org/protobuf/proto"
)

func ProtoUnmarshal(messageType int, data []byte, in proto.Message) error {
	if messageType == BinaryMessage {
		return proto.Unmarshal(data, in)
	} else if messageType == TextMessage {
		return json.UnmarshalOptions.Unmarshal(data, in)
	}
	return nil
}

func ProtoMarshal(messageType int, message proto.Message) []byte {
	var bytes []byte
	if messageType == BinaryMessage {
		bytes, _ = proto.Marshal(message)
	} else if messageType == TextMessage {
		bytes, _ = json.MarshalOptions.Marshal(message)
	}
	return bytes
}
