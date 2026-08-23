package option

import (
	"fmt"
	"testing"
)

func TestFormat(t *testing.T) {
	tests := map[string]struct {
		format string
		opt    Option[any]
		want   string
	}{
		"v integer":      {format: "%v", opt: Some[any](45), want: "45"},
		"v zero integer": {format: "%v", opt: Some[any](0), want: "0"},
		"v string":       {format: "%v", opt: Some[any]("test"), want: "test"},
		"v empty string": {format: "%v", opt: Some[any](""), want: ""},
		"v false":        {format: "%v", opt: Some[any](false), want: "false"},
		"v none":         {format: "%v", opt: None[any](), want: "None"},

		"d integer":  {format: "%d", opt: Some[any](45), want: "45"},
		"x integer":  {format: "%x", opt: Some[any](255), want: "ff"},
		"q string":   {format: "%q", opt: Some[any]("test"), want: `"test"`},
		"s string":   {format: "%s", opt: Some[any]("test"), want: "test"},
		"6.2f float": {format: "%6.2f", opt: Some[any](3.14159), want: "  3.14"},
		"plus v struct": {
			format: "%+v",
			opt:    Some[any](struct{ X, Y int }{1, 2}),
			want:   "{X:1 Y:2}",
		},

		"width some":      {format: "%6v", opt: Some[any](45), want: "    45"},
		"left width some": {format: "%-6v", opt: Some[any](45), want: "45    "},
		"width none":      {format: "%6v", opt: None[any](), want: "  None"},
		"left width none": {format: "%-6v", opt: None[any](), want: "None  "},

		// A None has no value to apply a verb to, so every verb prints the same.
		"d none":  {format: "%d", opt: None[any](), want: "None"},
		"q none":  {format: "%q", opt: None[any](), want: "None"},
		"s none":  {format: "%s", opt: None[any](), want: "None"},
		"#v none": {format: "%#v", opt: None[any](), want: "None"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			res := fmt.Sprintf(tc.format, tc.opt)

			if res != tc.want {
				t.Errorf("Want: %+v, Got: %+v", tc.want, res)
			}
		})
	}

	t.Run("prints exactly as the bare value does", func(t *testing.T) {
		cases := []struct {
			format string
			value  any
		}{
			{"%v", 45}, {"%v", 0}, {"%v", ""}, {"%d", 45}, {"%x", 255},
			{"%q", "test"}, {"%6.2f", 3.14159}, {"%6v", 45}, {"%-6v", 45},
			{"%+v", struct{ X int }{1}},
		}

		for _, c := range cases {
			want := fmt.Sprintf(c.format, c.value)
			res := fmt.Sprintf(c.format, Some(c.value))

			if res != want {
				t.Errorf("%s: Want: %+v, Got: %+v", c.format, want, res)
			}
		}
	})

	t.Run("reports the option type for T", func(t *testing.T) {
		// fmt resolves %T before consulting Formatter, so the wrapper stays visible.
		res := fmt.Sprintf("%T", Some(45))

		if want := "option.Option[int]"; res != want {
			t.Errorf("Want: %+v, Got: %+v", want, res)
		}
	})

	t.Run("nested options collapse", func(t *testing.T) {
		if res := fmt.Sprintf("%v", Some(Some(45))); res != "45" {
			t.Errorf("Want: %+v, Got: %+v", "45", res)
		}

		if res := fmt.Sprintf("%v", Some(None[int]())); res != "None" {
			t.Errorf("Want: %+v, Got: %+v", "None", res)
		}
	})

	t.Run("delegates to the stringer of T", func(t *testing.T) {
		res := fmt.Sprintf("%v", Some(fmt.Errorf("boom")))

		if want := "boom"; res != want {
			t.Errorf("Want: %+v, Got: %+v", want, res)
		}
	})
}
