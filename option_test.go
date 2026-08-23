package option

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"
)

// recoverPanic runs fn and reports the recovered value, if any.
func recoverPanic(t *testing.T, fn func()) (recovered any, panicked bool) {
	t.Helper()

	defer func() {
		if r := recover(); r != nil {
			recovered, panicked = r, true
		}
	}()

	fn()

	return nil, false
}

func TestSome(t *testing.T) {
	tests := map[string]struct {
		value int
		want  Option[int]
	}{
		"integer": {
			value: 45,
			want: Option[int]{
				value:   45,
				present: true,
			},
		},
		"zero integer": {
			value: 0,
			want: Option[int]{
				value:   0,
				present: true,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			res := Some(tc.value)

			if res != tc.want {
				t.Errorf("Want: %+v, Got: %+v", tc.want, res)
			}
		})
	}
}

func TestNone(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		res, want := None[string](), Option[string]{}

		if res != want {
			t.Errorf("Want: %+v, Got: %+v", want, res)
		}
	})

	t.Run("integer", func(t *testing.T) {
		res, want := None[int](), Option[int]{}

		if res != want {
			t.Errorf("Want: %+v, Got: %+v", want, res)
		}
	})

	t.Run("equals the zero value of the type", func(t *testing.T) {
		var zero Option[int]

		if zero != None[int]() {
			t.Errorf("Want: %+v, Got: %+v", None[int](), zero)
		}
	})
}

func TestFromTuple(t *testing.T) {
	tests := map[string]struct {
		value int
		ok    bool
		want  Option[int]
	}{
		"ok":                         {value: 45, ok: true, want: Some(45)},
		"not ok":                     {value: 0, ok: false, want: None[int]()},
		"not ok with non-zero value": {value: 45, ok: false, want: None[int]()},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			res := FromTuple(tc.value, tc.ok)

			if res != tc.want {
				t.Errorf("Want: %+v, Got: %+v", tc.want, res)
			}
		})
	}
}

func TestFromPtr(t *testing.T) {
	value := 45

	tests := map[string]struct {
		ptr  *int
		want Option[int]
	}{
		"non-nil pointer": {ptr: &value, want: Some(45)},
		"nil pointer":     {ptr: nil, want: None[int]()},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			res := FromPtr(tc.ptr)

			if res != tc.want {
				t.Errorf("Want: %+v, Got: %+v", tc.want, res)
			}
		})
	}

	t.Run("copies the pointed-to value", func(t *testing.T) {
		value := 45
		res := FromPtr(&value)
		value = 100

		if res != Some(45) {
			t.Errorf("Want: %+v, Got: %+v", Some(45), res)
		}
	})
}

func TestToPtr(t *testing.T) {
	t.Run("some", func(t *testing.T) {
		res := Some(45).ToPtr()

		if res == nil {
			t.Fatalf("Want: %+v, Got: %+v", 45, res)
		}

		if *res != 45 {
			t.Errorf("Want: %+v, Got: %+v", 45, *res)
		}
	})

	t.Run("none", func(t *testing.T) {
		res := None[int]().ToPtr()

		if res != nil {
			t.Errorf("Want: %+v, Got: %+v", nil, res)
		}
	})

	t.Run("does not alias the option", func(t *testing.T) {
		opt := Some(45)
		ptr := opt.ToPtr()
		*ptr = 100

		if opt != Some(45) {
			t.Errorf("Want: %+v, Got: %+v", Some(45), opt)
		}
	})
}

func TestIsSome(t *testing.T) {
	tests := map[string]struct {
		opt  Option[int]
		want bool
	}{
		"some":         {opt: Some(45), want: true},
		"some of zero": {opt: Some(0), want: true},
		"none":         {opt: None[int](), want: false},
		"zero value":   {opt: Option[int]{}, want: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			res := tc.opt.IsSome()

			if res != tc.want {
				t.Errorf("Want: %+v, Got: %+v", tc.want, res)
			}
		})
	}
}

func TestIsNone(t *testing.T) {
	tests := map[string]struct {
		opt  Option[int]
		want bool
	}{
		"some":         {opt: Some(45), want: false},
		"some of zero": {opt: Some(0), want: false},
		"none":         {opt: None[int](), want: true},
		"zero value":   {opt: Option[int]{}, want: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			res := tc.opt.IsNone()

			if res != tc.want {
				t.Errorf("Want: %+v, Got: %+v", tc.want, res)
			}
		})
	}
}

func TestIsSomeAnd(t *testing.T) {
	isPositive := func(i int) bool { return i > 0 }

	tests := map[string]struct {
		opt       Option[int]
		want      bool
		wantCalls int
	}{
		"some matching the predicate":     {opt: Some(45), want: true, wantCalls: 1},
		"some not matching the predicate": {opt: Some(-45), want: false, wantCalls: 1},
		"none":                            {opt: None[int](), want: false, wantCalls: 0},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			calls := 0
			res := tc.opt.IsSomeAnd(func(i int) bool {
				calls++
				return isPositive(i)
			})

			if res != tc.want {
				t.Errorf("Want: %+v, Got: %+v", tc.want, res)
			}

			if calls != tc.wantCalls {
				t.Errorf("Want: %+v calls, Got: %+v", tc.wantCalls, calls)
			}
		})
	}
}

func TestGet(t *testing.T) {
	tests := map[string]struct {
		opt       Option[int]
		wantValue int
		wantOk    bool
	}{
		"some":         {opt: Some(45), wantValue: 45, wantOk: true},
		"some of zero": {opt: Some(0), wantValue: 0, wantOk: true},
		"none":         {opt: None[int](), wantValue: 0, wantOk: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			value, ok := tc.opt.Get()

			if value != tc.wantValue || ok != tc.wantOk {
				t.Errorf("Want: %+v, %+v, Got: %+v, %+v", tc.wantValue, tc.wantOk, value, ok)
			}
		})
	}
}

func TestExpect(t *testing.T) {
	t.Run("some returns the value without panicking", func(t *testing.T) {
		var res int
		_, panicked := recoverPanic(t, func() { res = Some(45).Expect("must hold a value") })

		if panicked {
			t.Fatalf("Want: %+v, Got: %+v", false, panicked)
		}

		if res != 45 {
			t.Errorf("Want: %+v, Got: %+v", 45, res)
		}
	})

	t.Run("some of the zero value returns it", func(t *testing.T) {
		var res int
		_, panicked := recoverPanic(t, func() { res = Some(0).Expect("must hold a value") })

		if panicked {
			t.Fatalf("Want: %+v, Got: %+v", false, panicked)
		}

		if res != 0 {
			t.Errorf("Want: %+v, Got: %+v", 0, res)
		}
	})

	t.Run("none panics with the message", func(t *testing.T) {
		recovered, panicked := recoverPanic(t, func() { None[int]().Expect("must hold a value") })

		if !panicked {
			t.Fatalf("Want: %+v, Got: %+v", true, panicked)
		}

		err, ok := recovered.(error)
		if !ok {
			t.Fatalf("Want: %+v, Got: %+v", "error", recovered)
		}

		if !errors.Is(err, ErrValueIsNone) {
			t.Errorf("Want: %+v, Got: %+v", ErrValueIsNone, err)
		}

		if msg := "must hold a value"; !strings.Contains(err.Error(), msg) {
			t.Errorf("Want: %+v to contain %+v, Got: %+v", err.Error(), msg, err.Error())
		}
	})
}

func TestUnwrap(t *testing.T) {
	t.Run("some returns the value", func(t *testing.T) {
		res := Some(45).Unwrap()

		if res != 45 {
			t.Errorf("Want: %+v, Got: %+v", 45, res)
		}
	})

	t.Run("none panics", func(t *testing.T) {
		recovered, panicked := recoverPanic(t, func() { None[int]().Unwrap() })

		if !panicked {
			t.Fatalf("Want: %+v, Got: %+v", true, panicked)
		}

		err, ok := recovered.(error)
		if !ok {
			t.Fatalf("Want: %+v, Got: %+v", "error", recovered)
		}

		if !errors.Is(err, ErrValueIsNone) {
			t.Errorf("Want: %+v, Got: %+v", ErrValueIsNone, err)
		}
	})
}

func TestUnwrapOr(t *testing.T) {
	tests := map[string]struct {
		opt      Option[int]
		fallback int
		want     int
	}{
		"some":         {opt: Some(45), fallback: 100, want: 45},
		"some of zero": {opt: Some(0), fallback: 100, want: 0},
		"none":         {opt: None[int](), fallback: 100, want: 100},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			res := tc.opt.UnwrapOr(tc.fallback)

			if res != tc.want {
				t.Errorf("Want: %+v, Got: %+v", tc.want, res)
			}
		})
	}
}

func TestUnwrapOrElse(t *testing.T) {
	tests := map[string]struct {
		opt       Option[int]
		want      int
		wantCalls int
	}{
		"some":         {opt: Some(45), want: 45, wantCalls: 0},
		"some of zero": {opt: Some(0), want: 0, wantCalls: 0},
		"none":         {opt: None[int](), want: 100, wantCalls: 1},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			calls := 0
			res := tc.opt.UnwrapOrElse(func() int {
				calls++
				return 100
			})

			if res != tc.want {
				t.Errorf("Want: %+v, Got: %+v", tc.want, res)
			}

			if calls != tc.wantCalls {
				t.Errorf("Want: %+v calls, Got: %+v", tc.wantCalls, calls)
			}
		})
	}
}

func TestUnwrapOrZero(t *testing.T) {
	tests := map[string]struct {
		opt  Option[string]
		want string
	}{
		"some":             {opt: Some("test"), want: "test"},
		"some of the zero": {opt: Some(""), want: ""},
		"none":             {opt: None[string](), want: ""},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			res := tc.opt.UnwrapOrZero()

			if res != tc.want {
				t.Errorf("Want: %+v, Got: %+v", tc.want, res)
			}
		})
	}
}

func TestOkOr(t *testing.T) {
	errMissing := errors.New("the port is not configured")

	tests := map[string]struct {
		opt       Option[int]
		err       error
		wantValue int
		wantErr   error
	}{
		"some":              {opt: Some(45), err: errMissing, wantValue: 45, wantErr: nil},
		"some of zero":      {opt: Some(0), err: errMissing, wantValue: 0, wantErr: nil},
		"none":              {opt: None[int](), err: errMissing, wantValue: 0, wantErr: errMissing},
		"none with nil err": {opt: None[int](), err: nil, wantValue: 0, wantErr: ErrValueIsNone},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			res, err := tc.opt.OkOr(tc.err)

			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("Want: %+v, Got: %+v", tc.wantErr, err)
			}

			if res != tc.wantValue {
				t.Errorf("Want: %+v, Got: %+v", tc.wantValue, res)
			}
		})
	}

	t.Run("a none matches both the sentinel and the given error", func(t *testing.T) {
		_, err := None[int]().OkOr(errMissing)

		if !errors.Is(err, ErrValueIsNone) {
			t.Errorf("Want: %+v, Got: %+v", ErrValueIsNone, err)
		}

		if !errors.Is(err, errMissing) {
			t.Errorf("Want: %+v, Got: %+v", errMissing, err)
		}
	})

	t.Run("reads as the sentinel followed by the given error", func(t *testing.T) {
		_, err := None[int]().OkOr(errMissing)

		want := ErrValueIsNone.Error() + ": " + errMissing.Error()
		if err.Error() != want {
			t.Errorf("Want: %+v, Got: %+v", want, err.Error())
		}
	})

	t.Run("a nil error leaves no formatting artefact", func(t *testing.T) {
		// fmt would otherwise render a nil %w as "%!w(<nil>)".
		_, err := None[int]().OkOr(nil)

		if want := ErrValueIsNone.Error(); err.Error() != want {
			t.Errorf("Want: %+v, Got: %+v", want, err.Error())
		}
	})

	t.Run("keeps an error the caller already wrapped", func(t *testing.T) {
		_, err := None[int]().OkOr(fmt.Errorf("loading config: %w", errMissing))

		if !errors.Is(err, errMissing) {
			t.Errorf("Want: %+v, Got: %+v", errMissing, err)
		}
	})
}

func TestOkOrElse(t *testing.T) {
	errMissing := errors.New("the port is not configured")

	tests := map[string]struct {
		opt       Option[int]
		wantValue int
		wantErr   error
		wantCalls int
	}{
		"some":         {opt: Some(45), wantValue: 45, wantErr: nil, wantCalls: 0},
		"some of zero": {opt: Some(0), wantValue: 0, wantErr: nil, wantCalls: 0},
		"none":         {opt: None[int](), wantValue: 0, wantErr: errMissing, wantCalls: 1},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			calls := 0
			res, err := tc.opt.OkOrElse(func() error {
				calls++
				return errMissing
			})

			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("Want: %+v, Got: %+v", tc.wantErr, err)
			}

			if res != tc.wantValue {
				t.Errorf("Want: %+v, Got: %+v", tc.wantValue, res)
			}

			if calls != tc.wantCalls {
				t.Errorf("Want: %+v calls, Got: %+v", tc.wantCalls, calls)
			}
		})
	}

	t.Run("does not wrap the sentinel", func(t *testing.T) {
		// Unlike OkOr, fn's error is passed through untouched.
		_, err := None[int]().OkOrElse(func() error { return errMissing })

		if errors.Is(err, ErrValueIsNone) {
			t.Errorf("Want: %+v, Got: %+v", "no ErrValueIsNone", err)
		}
	})

	t.Run("keeps the error of fn wrappable", func(t *testing.T) {
		_, err := None[int]().OkOrElse(func() error {
			return fmt.Errorf("loading config: %w", errMissing)
		})

		if !errors.Is(err, errMissing) {
			t.Errorf("Want: %+v, Got: %+v", errMissing, err)
		}
	})
}

func TestMap(t *testing.T) {
	itoa := func(i int) string { return strconv.Itoa(i) }

	tests := map[string]struct {
		opt       Option[int]
		want      Option[string]
		wantCalls int
	}{
		"some":         {opt: Some(45), want: Some("45"), wantCalls: 1},
		"some of zero": {opt: Some(0), want: Some("0"), wantCalls: 1},
		"none":         {opt: None[int](), want: None[string](), wantCalls: 0},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			calls := 0
			res := tc.opt.Map(func(i int) string {
				calls++
				return itoa(i)
			})

			if res != tc.want {
				t.Errorf("Want: %+v, Got: %+v", tc.want, res)
			}

			if calls != tc.wantCalls {
				t.Errorf("Want: %+v calls, Got: %+v", tc.wantCalls, calls)
			}
		})
	}

	t.Run("maps to a different type", func(t *testing.T) {
		res := Some("test").Map(func(s string) int { return len(s) })

		if res != Some(4) {
			t.Errorf("Want: %+v, Got: %+v", Some(4), res)
		}
	})
}

func TestMapOr(t *testing.T) {
	itoa := func(i int) string { return strconv.Itoa(i) }

	tests := map[string]struct {
		opt       Option[int]
		fallback  string
		want      string
		wantCalls int
	}{
		"some":         {opt: Some(45), fallback: "fallback", want: "45", wantCalls: 1},
		"some of zero": {opt: Some(0), fallback: "fallback", want: "0", wantCalls: 1},
		"none":         {opt: None[int](), fallback: "fallback", want: "fallback", wantCalls: 0},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			calls := 0
			res := tc.opt.MapOr(tc.fallback, func(i int) string {
				calls++
				return itoa(i)
			})

			if res != tc.want {
				t.Errorf("Want: %+v, Got: %+v", tc.want, res)
			}

			if calls != tc.wantCalls {
				t.Errorf("Want: %+v calls, Got: %+v", tc.wantCalls, calls)
			}
		})
	}
}

func TestAnd(t *testing.T) {
	tests := map[string]struct {
		opt  Option[int]
		optB Option[string]
		want Option[string]
	}{
		"some and some": {opt: Some(45), optB: Some("test"), want: Some("test")},
		"some and none": {opt: Some(45), optB: None[string](), want: None[string]()},
		"none and some": {opt: None[int](), optB: Some("test"), want: None[string]()},
		"none and none": {opt: None[int](), optB: None[string](), want: None[string]()},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			res := tc.opt.And(tc.optB)

			if res != tc.want {
				t.Errorf("Want: %+v, Got: %+v", tc.want, res)
			}
		})
	}
}

func TestAndThen(t *testing.T) {
	tests := map[string]struct {
		opt       Option[int]
		fn        func(int) Option[string]
		want      Option[string]
		wantCalls int
	}{
		"some returning some": {
			opt:       Some(45),
			fn:        func(i int) Option[string] { return Some(strconv.Itoa(i)) },
			want:      Some("45"),
			wantCalls: 1,
		},
		"some returning none": {
			opt:       Some(45),
			fn:        func(int) Option[string] { return None[string]() },
			want:      None[string](),
			wantCalls: 1,
		},
		"none": {
			opt:       None[int](),
			fn:        func(i int) Option[string] { return Some(strconv.Itoa(i)) },
			want:      None[string](),
			wantCalls: 0,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			calls := 0
			res := tc.opt.AndThen(func(i int) Option[string] {
				calls++
				return tc.fn(i)
			})

			if res != tc.want {
				t.Errorf("Want: %+v, Got: %+v", tc.want, res)
			}

			if calls != tc.wantCalls {
				t.Errorf("Want: %+v calls, Got: %+v", tc.wantCalls, calls)
			}
		})
	}
}

func TestFilter(t *testing.T) {
	isPositive := func(i int) bool { return i > 0 }

	tests := map[string]struct {
		opt       Option[int]
		want      Option[int]
		wantCalls int
	}{
		"some matching the predicate":     {opt: Some(45), want: Some(45), wantCalls: 1},
		"some not matching the predicate": {opt: Some(-45), want: None[int](), wantCalls: 1},
		"none":                            {opt: None[int](), want: None[int](), wantCalls: 0},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			calls := 0
			res := tc.opt.Filter(func(i int) bool {
				calls++
				return isPositive(i)
			})

			if res != tc.want {
				t.Errorf("Want: %+v, Got: %+v", tc.want, res)
			}

			if calls != tc.wantCalls {
				t.Errorf("Want: %+v calls, Got: %+v", tc.wantCalls, calls)
			}
		})
	}
}

func TestOr(t *testing.T) {
	tests := map[string]struct {
		opt  Option[int]
		optB Option[int]
		want Option[int]
	}{
		"some or some": {opt: Some(45), optB: Some(100), want: Some(45)},
		"some or none": {opt: Some(45), optB: None[int](), want: Some(45)},
		"none or some": {opt: None[int](), optB: Some(100), want: Some(100)},
		"none or none": {opt: None[int](), optB: None[int](), want: None[int]()},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			res := tc.opt.Or(tc.optB)

			if res != tc.want {
				t.Errorf("Want: %+v, Got: %+v", tc.want, res)
			}
		})
	}
}

func TestOrElse(t *testing.T) {
	tests := map[string]struct {
		opt       Option[int]
		want      Option[int]
		wantCalls int
	}{
		"some": {opt: Some(45), want: Some(45), wantCalls: 0},
		"none": {opt: None[int](), want: Some(100), wantCalls: 1},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			calls := 0
			res := tc.opt.OrElse(func() Option[int] {
				calls++
				return Some(100)
			})

			if res != tc.want {
				t.Errorf("Want: %+v, Got: %+v", tc.want, res)
			}

			if calls != tc.wantCalls {
				t.Errorf("Want: %+v calls, Got: %+v", tc.wantCalls, calls)
			}
		})
	}
}

func TestXor(t *testing.T) {
	tests := map[string]struct {
		opt  Option[int]
		optB Option[int]
		want Option[int]
	}{
		"some xor none": {opt: Some(45), optB: None[int](), want: Some(45)},
		"none xor some": {opt: None[int](), optB: Some(100), want: Some(100)},
		"some xor some": {opt: Some(45), optB: Some(100), want: None[int]()},
		"none xor none": {opt: None[int](), optB: None[int](), want: None[int]()},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			res := tc.opt.Xor(tc.optB)

			if res != tc.want {
				t.Errorf("Want: %+v, Got: %+v", tc.want, res)
			}
		})
	}
}
