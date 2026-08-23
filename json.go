package option

import (
	"bytes"
	"encoding/json"
)

var _ json.Marshaler = Option[struct{}]{}

// MarshalJSON encodes the value, or null for a None.
// Tag a field with omitzero to drop a None entirely; omitempty does not
// work on structs and will still emit null.
func (o Option[T]) MarshalJSON() ([]byte, error) {
	if o.IsNone() {
		return json.Marshal(nil)
	}

	return json.Marshal(o.value)
}

var _ json.Unmarshaler = (*Option[struct{}])(nil)

// UnmarshalJSON decodes into the Option, reading JSON null as None.
// A missing field leaves the Option untouched, which is also None.
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
