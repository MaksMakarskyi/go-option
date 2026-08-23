package option

import (
	"encoding"
	"fmt"
	"strconv"
)

var _ encoding.TextMarshaler = Option[struct{}]{}

// MarshalText encodes the value as text, so an Option can be used as a JSON map
// key or anywhere else encoding.TextMarshaler is accepted.
// A None encodes as empty text, which Some of an empty string does too.
func (o Option[T]) MarshalText() ([]byte, error) {
	if o.IsNone() {
		return []byte{}, nil
	}

	if tm, ok := any(o.value).(encoding.TextMarshaler); ok {
		return tm.MarshalText()
	}

	if tm, ok := any(&o.value).(encoding.TextMarshaler); ok {
		return tm.MarshalText()
	}

	if text, ok := any(o.value).([]byte); ok {
		return append([]byte(nil), text...), nil
	}

	return fmt.Append(nil, o.value), nil
}

var _ encoding.TextUnmarshaler = (*Option[struct{}])(nil)

// UnmarshalText decodes text written by MarshalText, reading empty text as None.
// It supports any T whose pointer implements encoding.TextUnmarshaler as well as
// the basic kinds, and reports an error for anything else.
func (o *Option[T]) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*o = None[T]()
		return nil
	}

	var value T
	if err := unmarshalText(&value, text); err != nil {
		return err
	}

	*o = Some(value)
	return nil
}

// unmarshalText decodes text into value, preferring T's own TextUnmarshaler.
func unmarshalText[T any](value *T, text []byte) error {
	if tu, ok := any(value).(encoding.TextUnmarshaler); ok {
		return tu.UnmarshalText(text)
	}

	s := string(text)

	switch p := any(value).(type) {
	case *string:
		*p = s
	case *[]byte:
		*p = append([]byte(nil), text...)
	case *bool:
		v, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}

		*p = v
	case *int:
		return setSigned(p, s, strconv.IntSize)
	case *int8:
		return setSigned(p, s, 8)
	case *int16:
		return setSigned(p, s, 16)
	case *int32:
		return setSigned(p, s, 32)
	case *int64:
		return setSigned(p, s, 64)
	case *uint:
		return setUnsigned(p, s, strconv.IntSize)
	case *uint8:
		return setUnsigned(p, s, 8)
	case *uint16:
		return setUnsigned(p, s, 16)
	case *uint32:
		return setUnsigned(p, s, 32)
	case *uint64:
		return setUnsigned(p, s, 64)
	case *float32:
		return setFloat(p, s, 32)
	case *float64:
		return setFloat(p, s, 64)
	default:
		return fmt.Errorf("option: cannot unmarshal text into %T", *value)
	}

	return nil
}

// setSigned, setUnsigned and setFloat parse s at the given bit size and store
// the result, so each numeric case above stays a single line.
func setSigned[T ~int | ~int8 | ~int16 | ~int32 | ~int64](p *T, s string, bits int) error {
	v, err := strconv.ParseInt(s, 10, bits)
	if err != nil {
		return err
	}

	*p = T(v)
	return nil
}

func setUnsigned[T ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64](p *T, s string, bits int) error {
	v, err := strconv.ParseUint(s, 10, bits)
	if err != nil {
		return err
	}

	*p = T(v)
	return nil
}

func setFloat[T ~float32 | ~float64](p *T, s string, bits int) error {
	v, err := strconv.ParseFloat(s, bits)
	if err != nil {
		return err
	}

	*p = T(v)
	return nil
}
