package option

import (
	"encoding/json"
	"testing"
	"time"
)

func TestMarshalJSON(t *testing.T) {
	tests := map[string]struct {
		opt  Option[any]
		want string
	}{
		"some string":       {opt: Some[any]("test"), want: `"test"`},
		"some integer":      {opt: Some[any](45), want: `45`},
		"some zero integer": {opt: Some[any](0), want: `0`},
		"some empty string": {opt: Some[any](""), want: `""`},
		"some false":        {opt: Some[any](false), want: `false`},
		"some slice":        {opt: Some[any]([]int{1, 2}), want: `[1,2]`},
		"some empty slice":  {opt: Some[any]([]int{}), want: `[]`},
		"none":              {opt: None[any](), want: `null`},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			res, err := json.Marshal(tc.opt)
			if err != nil {
				t.Fatalf("Want: %+v, Got: %+v", nil, err)
			}

			if string(res) != tc.want {
				t.Errorf("Want: %+v, Got: %+v", tc.want, string(res))
			}
		})
	}

	t.Run("delegates to the marshaler of T", func(t *testing.T) {
		opt := Some(time.Date(2026, 8, 22, 0, 0, 0, 0, time.UTC))

		res, err := json.Marshal(opt)
		if err != nil {
			t.Fatalf("Want: %+v, Got: %+v", nil, err)
		}

		if want := `"2026-08-22T00:00:00Z"`; string(res) != want {
			t.Errorf("Want: %+v, Got: %+v", want, string(res))
		}
	})

	t.Run("marshals when not addressable", func(t *testing.T) {
		res, err := json.Marshal(map[string]Option[int]{"k": Some(45)})
		if err != nil {
			t.Fatalf("Want: %+v, Got: %+v", nil, err)
		}

		if want := `{"k":45}`; string(res) != want {
			t.Errorf("Want: %+v, Got: %+v", want, string(res))
		}
	})
}

func TestUnmarshalJSON(t *testing.T) {
	tests := map[string]struct {
		data string
		want Option[int]
	}{
		"integer":      {data: `45`, want: Some(45)},
		"zero integer": {data: `0`, want: Some(0)},
		"null":         {data: `null`, want: None[int]()},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var res Option[int]
			if err := json.Unmarshal([]byte(tc.data), &res); err != nil {
				t.Fatalf("Want: %+v, Got: %+v", nil, err)
			}

			if res != tc.want {
				t.Errorf("Want: %+v, Got: %+v", tc.want, res)
			}
		})
	}

	t.Run("null produces a value indistinguishable from None", func(t *testing.T) {
		var res Option[int]
		if err := json.Unmarshal([]byte(`null`), &res); err != nil {
			t.Fatalf("Want: %+v, Got: %+v", nil, err)
		}

		if res != (Option[int]{}) {
			t.Errorf("Want: %+v, Got: %+v", Option[int]{}, res)
		}
	})

	t.Run("null clears a reused option", func(t *testing.T) {
		res := Some(45)
		if err := json.Unmarshal([]byte(`null`), &res); err != nil {
			t.Fatalf("Want: %+v, Got: %+v", nil, err)
		}

		if res != None[int]() {
			t.Errorf("Want: %+v, Got: %+v", None[int](), res)
		}
	})

	t.Run("reports the error of T", func(t *testing.T) {
		var res Option[int]
		err := json.Unmarshal([]byte(`"nope"`), &res)

		if err == nil {
			t.Fatalf("Want: %+v, Got: %+v", "an error", nil)
		}
	})

	t.Run("leaves the option untouched on error", func(t *testing.T) {
		res := Some(45)
		if err := json.Unmarshal([]byte(`"nope"`), &res); err == nil {
			t.Fatalf("Want: %+v, Got: %+v", "an error", nil)
		}

		if res != Some(45) {
			t.Errorf("Want: %+v, Got: %+v", Some(45), res)
		}
	})

	t.Run("non-comparable value type", func(t *testing.T) {
		var res Option[[]int]
		if err := json.Unmarshal([]byte(`[1,2]`), &res); err != nil {
			t.Fatalf("Want: %+v, Got: %+v", nil, err)
		}

		value, ok := res.Get()
		if !ok || len(value) != 2 || value[0] != 1 || value[1] != 2 {
			t.Errorf("Want: %+v, Got: %+v", []int{1, 2}, value)
		}
	})
}

func TestJSONRoundTrip(t *testing.T) {
	type payload struct {
		Name     Option[string] `json:"name"`
		Age      Option[int]    `json:"age"`
		Nickname Option[string] `json:"nickname,omitzero"`
	}

	t.Run("some values survive the round trip", func(t *testing.T) {
		want := payload{Name: Some("bob"), Age: Some(0), Nickname: Some("bo")}

		data, err := json.Marshal(want)
		if err != nil {
			t.Fatalf("Want: %+v, Got: %+v", nil, err)
		}

		var res payload
		if err := json.Unmarshal(data, &res); err != nil {
			t.Fatalf("Want: %+v, Got: %+v", nil, err)
		}

		if res != want {
			t.Errorf("Want: %+v, Got: %+v", want, res)
		}
	})

	t.Run("none marshals as null, and omitzero drops the field", func(t *testing.T) {
		res, err := json.Marshal(payload{})
		if err != nil {
			t.Fatalf("Want: %+v, Got: %+v", nil, err)
		}

		if want := `{"name":null,"age":null}`; string(res) != want {
			t.Errorf("Want: %+v, Got: %+v", want, string(res))
		}
	})

	t.Run("a missing field unmarshals as none", func(t *testing.T) {
		var res payload
		if err := json.Unmarshal([]byte(`{"name":"bob"}`), &res); err != nil {
			t.Fatalf("Want: %+v, Got: %+v", nil, err)
		}

		want := payload{Name: Some("bob")}
		if res != want {
			t.Errorf("Want: %+v, Got: %+v", want, res)
		}
	})
}
