package option

import (
	"bytes"
	"encoding/json"
	"math"
	"testing"
	"time"
)

// ptrMarshaler implements TextMarshaler on its pointer, not its value.
type ptrMarshaler struct{ value string }

func (p *ptrMarshaler) MarshalText() ([]byte, error) { return []byte(p.value), nil }

func TestMarshalText(t *testing.T) {
	tests := map[string]struct {
		opt  Option[any]
		want string
	}{
		"string":       {opt: Some[any]("test"), want: "test"},
		"integer":      {opt: Some[any](45), want: "45"},
		"zero integer": {opt: Some[any](0), want: "0"},
		"float":        {opt: Some[any](3.5), want: "3.5"},
		"bool":         {opt: Some[any](true), want: "true"},
		"empty string": {opt: Some[any](""), want: ""},
		"none":         {opt: None[any](), want: ""},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			res, err := tc.opt.MarshalText()
			if err != nil {
				t.Fatalf("Want: %+v, Got: %+v", nil, err)
			}

			if string(res) != tc.want {
				t.Errorf("Want: %+v, Got: %+v", tc.want, string(res))
			}
		})
	}

	t.Run("delegates to the text marshaler of T", func(t *testing.T) {
		opt := Some(time.Date(2026, 8, 23, 0, 0, 0, 0, time.UTC))

		res, err := opt.MarshalText()
		if err != nil {
			t.Fatalf("Want: %+v, Got: %+v", nil, err)
		}

		if want := "2026-08-23T00:00:00Z"; string(res) != want {
			t.Errorf("Want: %+v, Got: %+v", want, string(res))
		}
	})

	t.Run("delegates to a pointer-receiver text marshaler", func(t *testing.T) {
		res, err := Some(ptrMarshaler{value: "test"}).MarshalText()
		if err != nil {
			t.Fatalf("Want: %+v, Got: %+v", nil, err)
		}

		if want := "test"; string(res) != want {
			t.Errorf("Want: %+v, Got: %+v", want, string(res))
		}
	})

	t.Run("none and some of empty text are indistinguishable", func(t *testing.T) {
		// Text has no null, so this collision is unavoidable and documented.
		none, _ := None[string]().MarshalText()
		empty, _ := Some("").MarshalText()

		if string(none) != string(empty) {
			t.Errorf("Want: %+v, Got: %+v", string(empty), string(none))
		}
	})
}

func TestUnmarshalText(t *testing.T) {
	tests := map[string]struct {
		text string
		want Option[int]
	}{
		"integer":      {text: "45", want: Some(45)},
		"zero integer": {text: "0", want: Some(0)},
		"negative":     {text: "-45", want: Some(-45)},
		"empty":        {text: "", want: None[int]()},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var res Option[int]
			if err := res.UnmarshalText([]byte(tc.text)); err != nil {
				t.Fatalf("Want: %+v, Got: %+v", nil, err)
			}

			if res != tc.want {
				t.Errorf("Want: %+v, Got: %+v", tc.want, res)
			}
		})
	}

	t.Run("empty text clears a reused option", func(t *testing.T) {
		res := Some(45)
		if err := res.UnmarshalText(nil); err != nil {
			t.Fatalf("Want: %+v, Got: %+v", nil, err)
		}

		if res != None[int]() {
			t.Errorf("Want: %+v, Got: %+v", None[int](), res)
		}
	})

	t.Run("every supported kind round trips", func(t *testing.T) {
		assertRoundTrip(t, "test")
		assertRoundTrip(t, true)
		assertRoundTrip(t, false)
		assertRoundTrip(t, 45)
		assertRoundTrip(t, int8(-8))
		assertRoundTrip(t, int16(-16))
		assertRoundTrip(t, int32(-32))
		assertRoundTrip(t, int64(-64))
		assertRoundTrip(t, uint(45))
		assertRoundTrip(t, uint8(255))
		assertRoundTrip(t, uint16(65535))
		assertRoundTrip(t, uint32(4294967295))
		assertRoundTrip(t, uint64(18446744073709551615))
		assertRoundTrip(t, float32(2.5))
		assertRoundTrip(t, math.Pi)
	})

	t.Run("some of an empty string decodes as none", func(t *testing.T) {
		// The collision documented on MarshalText, pinned so it stays deliberate.
		if res := roundTrip(t, Some("")); res != None[string]() {
			t.Errorf("Want: %+v, Got: %+v", None[string](), res)
		}
	})

	t.Run("byte slices round trip", func(t *testing.T) {
		res := roundTrip(t, Some([]byte("test")))

		value, ok := res.Get()
		if !ok || !bytes.Equal(value, []byte("test")) {
			t.Errorf("Want: %+v, Got: %+v", "test", string(value))
		}
	})

	t.Run("reports errors for every numeric family", func(t *testing.T) {
		var i Option[int]
		if err := i.UnmarshalText([]byte("nope")); err == nil {
			t.Errorf("int: Want: %+v, Got: %+v", "an error", nil)
		}

		var u Option[uint]
		if err := u.UnmarshalText([]byte("-1")); err == nil {
			t.Errorf("uint: Want: %+v, Got: %+v", "an error", nil)
		}

		var f Option[float32]
		if err := f.UnmarshalText([]byte("nope")); err == nil {
			t.Errorf("float32: Want: %+v, Got: %+v", "an error", nil)
		}

		var b Option[bool]
		if err := b.UnmarshalText([]byte("nope")); err == nil {
			t.Errorf("bool: Want: %+v, Got: %+v", "an error", nil)
		}
	})

	t.Run("delegates to the text unmarshaler of T", func(t *testing.T) {
		want := Some(time.Date(2026, 8, 23, 0, 0, 0, 0, time.UTC))

		if res := roundTrip(t, want); !res.Unwrap().Equal(want.Unwrap()) {
			t.Errorf("Want: %+v, Got: %+v", want, res)
		}
	})

	t.Run("reports a parse error", func(t *testing.T) {
		var res Option[int]

		if err := res.UnmarshalText([]byte("nope")); err == nil {
			t.Fatalf("Want: %+v, Got: %+v", "an error", nil)
		}
	})

	t.Run("reports an overflow", func(t *testing.T) {
		var res Option[int8]

		if err := res.UnmarshalText([]byte("999")); err == nil {
			t.Fatalf("Want: %+v, Got: %+v", "an error", nil)
		}
	})

	t.Run("reports an unsupported type", func(t *testing.T) {
		var res Option[struct{ A int }]

		if err := res.UnmarshalText([]byte("test")); err == nil {
			t.Fatalf("Want: %+v, Got: %+v", "an error", nil)
		}
	})

	t.Run("leaves the option untouched on error", func(t *testing.T) {
		res := Some(45)
		if err := res.UnmarshalText([]byte("nope")); err == nil {
			t.Fatalf("Want: %+v, Got: %+v", "an error", nil)
		}

		if res != Some(45) {
			t.Errorf("Want: %+v, Got: %+v", Some(45), res)
		}
	})
}

// assertRoundTrip checks that Some(value) survives a text round trip.
func assertRoundTrip[T comparable](t *testing.T, value T) {
	t.Helper()

	want := Some(value)
	if res := roundTrip(t, want); res != want {
		t.Errorf("%T: Want: %+v, Got: %+v", value, want, res)
	}
}

// roundTrip encodes opt to text and decodes it back.
func roundTrip[T any](t *testing.T, opt Option[T]) Option[T] {
	t.Helper()

	text, err := opt.MarshalText()
	if err != nil {
		t.Fatalf("Want: %+v, Got: %+v", nil, err)
	}

	var res Option[T]
	if err := res.UnmarshalText(text); err != nil {
		t.Fatalf("Want: %+v, Got: %+v", nil, err)
	}

	return res
}

func TestTextAsJSONMapKey(t *testing.T) {
	t.Run("string keys", func(t *testing.T) {
		res, err := json.Marshal(map[Option[string]]int{Some("a"): 1})
		if err != nil {
			t.Fatalf("Want: %+v, Got: %+v", nil, err)
		}

		if want := `{"a":1}`; string(res) != want {
			t.Errorf("Want: %+v, Got: %+v", want, string(res))
		}
	})

	t.Run("integer keys", func(t *testing.T) {
		res, err := json.Marshal(map[Option[int]]int{Some(7): 1})
		if err != nil {
			t.Fatalf("Want: %+v, Got: %+v", nil, err)
		}

		if want := `{"7":1}`; string(res) != want {
			t.Errorf("Want: %+v, Got: %+v", want, string(res))
		}
	})

	t.Run("keys decode again", func(t *testing.T) {
		var res map[Option[string]]int
		if err := json.Unmarshal([]byte(`{"a":1}`), &res); err != nil {
			t.Fatalf("Want: %+v, Got: %+v", nil, err)
		}

		if res[Some("a")] != 1 {
			t.Errorf("Want: %+v, Got: %+v", 1, res[Some("a")])
		}
	})

	t.Run("marshal json still wins over marshal text", func(t *testing.T) {
		// Adding TextMarshaler must not turn values into JSON strings.
		res, err := json.Marshal(Some(45))
		if err != nil {
			t.Fatalf("Want: %+v, Got: %+v", nil, err)
		}

		if want := "45"; string(res) != want {
			t.Errorf("Want: %+v, Got: %+v", want, string(res))
		}
	})
}
