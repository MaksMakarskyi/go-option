package option

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"
	"time"
)

func TestSQLInterfaces(t *testing.T) {
	// Compile-time proof that database/sql will actually pick these up.
	var _ driver.Valuer = Option[int]{}
	var _ sql.Scanner = (*Option[int])(nil)
}

func TestValue(t *testing.T) {
	tests := map[string]struct {
		opt  driver.Valuer
		want driver.Value
	}{
		"int becomes int64":       {opt: Some(45), want: int64(45)},
		"zero int becomes int64":  {opt: Some(0), want: int64(0)},
		"int64 stays int64":       {opt: Some(int64(45)), want: int64(45)},
		"float32 becomes float64": {opt: Some(float32(2.5)), want: float64(2.5)},
		"float64 stays float64":   {opt: Some(3.5), want: 3.5},
		"string":                  {opt: Some("test"), want: "test"},
		"empty string":            {opt: Some(""), want: ""},
		"bool":                    {opt: Some(true), want: true},
		"false":                   {opt: Some(false), want: false},
		"none is null":            {opt: None[int](), want: nil},
		"none string is null":     {opt: None[string](), want: nil},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			res, err := tc.opt.Value()
			if err != nil {
				t.Fatalf("Want: %+v, Got: %+v", nil, err)
			}

			if res != tc.want {
				t.Errorf("Want: %+v (%T), Got: %+v (%T)", tc.want, tc.want, res, res)
			}
		})
	}

	t.Run("byte slice", func(t *testing.T) {
		res, err := Some([]byte("test")).Value()
		if err != nil {
			t.Fatalf("Want: %+v, Got: %+v", nil, err)
		}

		value, ok := res.([]byte)
		if !ok || !bytes.Equal(value, []byte("test")) {
			t.Errorf("Want: %+v, Got: %+v", "test", res)
		}
	})

	t.Run("time", func(t *testing.T) {
		want := time.Date(2026, 8, 23, 0, 0, 0, 0, time.UTC)

		res, err := Some(want).Value()
		if err != nil {
			t.Fatalf("Want: %+v, Got: %+v", nil, err)
		}

		if res != driver.Value(want) {
			t.Errorf("Want: %+v, Got: %+v", want, res)
		}
	})

	t.Run("reports a type the driver cannot store", func(t *testing.T) {
		if _, err := Some([]int{1, 2}).Value(); err == nil {
			t.Errorf("Want: %+v, Got: %+v", "an error", nil)
		}

		if _, err := Some(struct{ A int }{1}).Value(); err == nil {
			t.Errorf("Want: %+v, Got: %+v", "an error", nil)
		}
	})

	t.Run("a none of an unstorable type is still null", func(t *testing.T) {
		// Nothing is written, so the element type never has to be convertible.
		res, err := None[[]int]().Value()
		if err != nil {
			t.Fatalf("Want: %+v, Got: %+v", nil, err)
		}

		if res != nil {
			t.Errorf("Want: %+v, Got: %+v", nil, res)
		}
	})

	t.Run("output is accepted by database/sql", func(t *testing.T) {
		// The conversion database/sql runs on every argument.
		for _, opt := range []driver.Valuer{
			Some(45), Some("test"), Some(true), Some(3.5),
			Some([]byte("b")), Some(time.Now()), None[int](),
		} {
			if _, err := driver.DefaultParameterConverter.ConvertValue(opt); err != nil {
				t.Errorf("%T: Want: %+v, Got: %+v", opt, nil, err)
			}
		}
	})
}

func TestScan(t *testing.T) {
	tests := map[string]struct {
		src  any
		want Option[int]
	}{
		"int64":      {src: int64(45), want: Some(45)},
		"zero":       {src: int64(0), want: Some(0)},
		"negative":   {src: int64(-45), want: Some(-45)},
		"null":       {src: nil, want: None[int]()},
		"string":     {src: "45", want: Some(45)},
		"byte slice": {src: []byte("45"), want: Some(45)},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var res Option[int]
			if err := res.Scan(tc.src); err != nil {
				t.Fatalf("Want: %+v, Got: %+v", nil, err)
			}

			if res != tc.want {
				t.Errorf("Want: %+v, Got: %+v", tc.want, res)
			}
		})
	}

	t.Run("driver types reach the matching option", func(t *testing.T) {
		var s Option[string]
		if err := s.Scan([]byte("test")); err != nil || s != Some("test") {
			t.Errorf("Want: %+v, Got: %+v (err %v)", Some("test"), s, err)
		}

		var f Option[float64]
		if err := f.Scan(3.5); err != nil || f != Some(3.5) {
			t.Errorf("Want: %+v, Got: %+v (err %v)", Some(3.5), f, err)
		}

		var b Option[bool]
		if err := b.Scan(true); err != nil || b != Some(true) {
			t.Errorf("Want: %+v, Got: %+v (err %v)", Some(true), b, err)
		}

		want := time.Date(2026, 8, 23, 0, 0, 0, 0, time.UTC)
		var tm Option[time.Time]
		if err := tm.Scan(want); err != nil || tm != Some(want) {
			t.Errorf("Want: %+v, Got: %+v (err %v)", Some(want), tm, err)
		}
	})

	t.Run("null clears a reused option", func(t *testing.T) {
		res := Some(45)
		if err := res.Scan(nil); err != nil {
			t.Fatalf("Want: %+v, Got: %+v", nil, err)
		}

		if res != None[int]() {
			t.Errorf("Want: %+v, Got: %+v", None[int](), res)
		}
	})

	t.Run("null produces a value indistinguishable from None", func(t *testing.T) {
		var res Option[int]
		if err := res.Scan(nil); err != nil {
			t.Fatalf("Want: %+v, Got: %+v", nil, err)
		}

		if res != (Option[int]{}) {
			t.Errorf("Want: %+v, Got: %+v", Option[int]{}, res)
		}
	})

	t.Run("reports a conversion error", func(t *testing.T) {
		var res Option[int]

		if err := res.Scan("not a number"); err == nil {
			t.Errorf("Want: %+v, Got: %+v", "an error", nil)
		}
	})

	t.Run("reports an unsupported destination", func(t *testing.T) {
		var res Option[[]int]

		if err := res.Scan(int64(1)); err == nil {
			t.Errorf("Want: %+v, Got: %+v", "an error", nil)
		}
	})

	t.Run("leaves the option untouched on error", func(t *testing.T) {
		res := Some(45)
		if err := res.Scan("not a number"); err == nil {
			t.Fatalf("Want: %+v, Got: %+v", "an error", nil)
		}

		if res != Some(45) {
			t.Errorf("Want: %+v, Got: %+v", Some(45), res)
		}
	})
}

// scannerValuer is a T that brings its own database conversion.
type scannerValuer struct{ value string }

func (s scannerValuer) Value() (driver.Value, error) { return "sv:" + s.value, nil }

func (s *scannerValuer) Scan(src any) error {
	text, ok := src.(string)
	if !ok {
		return errors.New("scannerValuer: want string")
	}

	s.value = text
	return nil
}

func TestSQLDelegatesToT(t *testing.T) {
	t.Run("value", func(t *testing.T) {
		res, err := Some(scannerValuer{value: "test"}).Value()
		if err != nil {
			t.Fatalf("Want: %+v, Got: %+v", nil, err)
		}

		if want := "sv:test"; res != driver.Value(want) {
			t.Errorf("Want: %+v, Got: %+v", want, res)
		}
	})

	t.Run("scan", func(t *testing.T) {
		var res Option[scannerValuer]
		if err := res.Scan("test"); err != nil {
			t.Fatalf("Want: %+v, Got: %+v", nil, err)
		}

		if want := (scannerValuer{value: "test"}); res != Some(want) {
			t.Errorf("Want: %+v, Got: %+v", Some(want), res)
		}
	})
}

func TestSQLRoundTrip(t *testing.T) {
	t.Run("some", func(t *testing.T) {
		want := Some(45)

		encoded, err := want.Value()
		if err != nil {
			t.Fatalf("Want: %+v, Got: %+v", nil, err)
		}

		var res Option[int]
		if err := res.Scan(encoded); err != nil {
			t.Fatalf("Want: %+v, Got: %+v", nil, err)
		}

		if res != want {
			t.Errorf("Want: %+v, Got: %+v", want, res)
		}
	})

	t.Run("none", func(t *testing.T) {
		encoded, err := None[string]().Value()
		if err != nil {
			t.Fatalf("Want: %+v, Got: %+v", nil, err)
		}

		var res Option[string]
		if err := res.Scan(encoded); err != nil {
			t.Fatalf("Want: %+v, Got: %+v", nil, err)
		}

		if res != None[string]() {
			t.Errorf("Want: %+v, Got: %+v", None[string](), res)
		}
	})
}
