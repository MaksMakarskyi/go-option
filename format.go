package option

import (
	"fmt"
	"strconv"
)

// noneText is how an empty Option prints.
const noneText = "None"

var _ fmt.Formatter = Option[struct{}]{}

// Format forwards the verb, flags, width and precision to the contained value,
// so an Option prints exactly as that value would.
// A None prints as None under every verb.
func (o Option[T]) Format(f fmt.State, verb rune) {
	if o.IsNone() {
		writeNone(f)
		return
	}

	fmt.Fprintf(f, fmt.FormatString(f, verb), o.value)
}

// writeNone prints noneText, honouring only width and left-alignment so that a
// None still lines up in a padded column whatever the verb was.
func writeNone(f fmt.State) {
	format := "%"
	if f.Flag('-') {
		format += "-"
	}

	if width, ok := f.Width(); ok {
		format += strconv.Itoa(width)
	}

	fmt.Fprintf(f, format+"s", noneText)
}
