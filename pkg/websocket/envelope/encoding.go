package envelope

import (
	"encoding/json"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/websocket/base"
)

func marshal[T base.IEnvelope](envelope base.IEnvelope) ([]byte, error) {
	_e, ok := envelope.(T)
	if ok {
		return json.Marshal(_e)
	}
	return nil, base.ErrInvalidEnvelope
}

func unmarshal[T base.IEnvelope](data []byte) (base.IEnvelope, error) {
	var _e T
	err := json.Unmarshal(data, &_e)
	if err != nil {
		return nil, err
	}
	return _e, nil
}
