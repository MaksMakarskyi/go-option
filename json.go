package option

import (
	"bytes"
	"encoding/json"
)

var _ json.Marshaler = Option[struct{}]{}

func (o Option[T]) MarshalJSON() ([]byte, error) {
	if o.IsNone() {
		return json.Marshal(nil)
	}

	return json.Marshal(o.value)
}

var _ json.Unmarshaler = (*Option[struct{}])(nil)

func (o *Option[T]) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		*o = None[T]()
		return nil
	}

	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	*o = Some(value)
	return nil
}
