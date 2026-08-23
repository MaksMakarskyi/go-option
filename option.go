// Package option provides a generic Option type: a value that is either
// present (Some) or absent (None).
//
// It replaces nil pointers and (T, bool) pairs when a value is genuinely
// optional, and unlike a zero value it can tell "set to 0" from "not set".
package option

import "fmt"

// Option holds a value of type T that may or may not be present.
// Its zero value is a valid None, so a declared Option is ready to use.
type Option[T any] struct {
	value   T
	present bool
}

// Some wraps a value in a present Option. The value is copied, so later
// changes to the original do not affect the Option.
//
//	option.Some(45).Unwrap() // 45
//	option.Some(0).IsSome()  // true, a zero value is still present
func Some[T any](value T) Option[T] {
	return Option[T]{
		value:   value,
		present: true,
	}
}

// None returns an empty Option. It is identical to the zero value of the type.
//
//	option.None[int]().IsNone()    // true
//	option.None[int]().UnwrapOr(7) // 7
func None[T any]() Option[T] {
	return Option[T]{}
}

// FromTuple turns a (value, ok) pair into an Option.
// When ok is false the value is discarded, so the result equals None.
//
//	v, ok := m["key"]
//	option.FromTuple(v, ok) // Some(v) if the key was there, None otherwise
func FromTuple[T any](value T, ok bool) Option[T] {
	if !ok {
		return None[T]()
	}

	return Some(value)
}

// FromPtr turns a pointer into an Option: nil becomes None, anything else
// becomes Some holding a copy of the pointed-to value.
//
//	option.FromPtr(&n)       // Some(n), holding a copy
//	option.FromPtr[int](nil) // None
func FromPtr[T any](ptr *T) Option[T] {
	if ptr == nil {
		return None[T]()
	}

	return Some(*ptr)
}

// ToPtr returns a pointer to a copy of the value, or nil for a None.
// Writing through that pointer does not change the Option.
//
//	*option.Some(45).ToPtr()   // 45
//	option.None[int]().ToPtr() // nil
func (o Option[T]) ToPtr() *T {
	if !o.present {
		return nil
	}

	return &o.value
}

// IsSome reports whether a value is present.
//
//	option.Some(45).IsSome()    // true
//	option.None[int]().IsSome() // false
func (o Option[T]) IsSome() bool {
	return o.present
}

// IsNone reports whether the Option is empty.
//
//	option.None[int]().IsNone() // true
//	option.Some(45).IsNone()    // false
func (o Option[T]) IsNone() bool {
	return !o.present
}

// IsSomeAnd reports whether a value is present and satisfies pred.
// pred is not called on a None.
//
//	positive := func(n int) bool { return n > 0 }
//	option.Some(45).IsSomeAnd(positive)    // true
//	option.Some(-1).IsSomeAnd(positive)    // false
//	option.None[int]().IsSomeAnd(positive) // false, positive is not called
func (o Option[T]) IsSomeAnd(pred func(T) bool) bool {
	if o.IsNone() {
		return false
	}

	return pred(o.value)
}

// Get returns the value and whether it was present, in Go's usual comma-ok style.
//
//	if v, ok := o.Get(); ok {
//		fmt.Println(v)
//	}
//	option.None[int]().Get() // 0, false
func (o Option[T]) Get() (T, bool) {
	return o.value, o.present
}

// Expect returns the value, and panics with msg if there is none.
// The panic value wraps [ErrValueIsNone].
//
//	option.Some(45).Expect("port must be set")    // 45
//	option.None[int]().Expect("port must be set") // panics: the value is none: port must be set
func (o Option[T]) Expect(msg string) T {
	if o.IsNone() {
		panic(fmt.Errorf("%w: %s", ErrValueIsNone, msg))
	}

	return o.value
}

// Unwrap returns the value and panics with [ErrValueIsNone] if there is none.
// Reach for it only when a value is guaranteed.
//
//	option.Some(45).Unwrap()    // 45
//	option.None[int]().Unwrap() // panics with ErrValueIsNone
func (o Option[T]) Unwrap() T {
	if o.IsNone() {
		panic(ErrValueIsNone)
	}

	return o.value
}

// UnwrapOr returns the value, or the given fallback if the Option is None.
//
//	option.Some(45).UnwrapOr(7)    // 45
//	option.None[int]().UnwrapOr(7) // 7
func (o Option[T]) UnwrapOr(value T) T {
	if o.IsNone() {
		return value
	}

	return o.value
}

// UnwrapOrElse returns the value, or the result of fn if the Option is None.
// fn is only called on a None, so an expensive fallback stays cheap.
//
//	option.Some(45).UnwrapOrElse(loadDefault)    // 45, loadDefault is not called
//	option.None[int]().UnwrapOrElse(loadDefault) // whatever loadDefault returns
func (o Option[T]) UnwrapOrElse(fn func() T) T {
	if o.IsNone() {
		return fn()
	}

	return o.value
}

// UnwrapOrZero returns the value, or the zero value of T if there is none.
//
//	option.Some("x").UnwrapOrZero()      // "x"
//	option.None[string]().UnwrapOrZero() // ""
func (o Option[T]) UnwrapOrZero() T {
	if o.IsNone() {
		var zero T
		return zero
	}

	return o.value
}

// OkOr returns the value, or err joined with [ErrValueIsNone] if there is none,
// so errors.Is matches either one. A nil err gives [ErrValueIsNone] on its own.
// Handy for feeding an Option into ordinary error handling.
//
//	port, err := cfg.Port.OkOr(errors.New("port not set"))
//	// a None gives "the value is none: port not set", and errors.Is
//	// matches both ErrValueIsNone and the error you passed in
func (o Option[T]) OkOr(err error) (T, error) {
	if o.IsNone() {
		var zero T
		if err == nil {
			return zero, ErrValueIsNone
		}

		return zero, fmt.Errorf("%w: %w", ErrValueIsNone, err)
	}

	return o.value, nil
}

// OkOrElse returns the value, or the error from fn if there is none.
// fn is only called on a None, so the error can carry context about what was
// missing. Unlike OkOr the error is passed through untouched, so it does not
// wrap [ErrValueIsNone] unless fn puts it there.
//
//	port, err := cfg.Port.OkOrElse(func() error {
//		return fmt.Errorf("port not set in %s", path)
//	})
//	// the error arrives exactly as fn built it
func (o Option[T]) OkOrElse(fn func() error) (T, error) {
	if o.IsNone() {
		var zero T
		return zero, fn()
	}

	return o.value, nil
}

// Map applies fn to the value and wraps the result in a new Option.
// A None passes straight through and fn is not called.
//
//	option.Some(45).Map(strconv.Itoa)    // Some("45")
//	option.None[int]().Map(strconv.Itoa) // None, strconv.Itoa is not called
func (o Option[T]) Map[U any](fn func(T) U) Option[U] {
	if o.IsNone() {
		return Option[U]{}
	}

	return Some(fn(o.value))
}

// MapOr applies fn to the value, or returns the given fallback for a None.
// fn is not called on a None.
//
//	option.Some(45).MapOr("-", strconv.Itoa)    // "45"
//	option.None[int]().MapOr("-", strconv.Itoa) // "-"
func (o Option[T]) MapOr[U any](value U, fn func(T) U) U {
	if o.IsNone() {
		return value
	}

	return fn(o.value)
}

// And returns optB when this Option is Some, and None otherwise.
//
//	option.Some(45).And(option.Some("b"))    // Some("b")
//	option.None[int]().And(option.Some("b")) // None
func (o Option[T]) And[U any](optB Option[U]) Option[U] {
	if o.IsNone() {
		return None[U]()
	}

	return optB
}

// AndThen calls fn on the value and returns its result, so steps that may
// each come up empty can be chained. fn is not called on a None.
//
//	half := func(n int) option.Option[int] {
//		if n%2 == 0 {
//			return option.Some(n / 2)
//		}
//		return option.None[int]()
//	}
//	option.Some(8).AndThen(half) // Some(4)
//	option.Some(7).AndThen(half) // None
func (o Option[T]) AndThen[U any](fn func(T) Option[U]) Option[U] {
	if o.IsNone() {
		return None[U]()
	}

	return fn(o.value)
}

// Filter keeps the value only if it satisfies fn, and returns None otherwise.
// fn is not called on a None.
//
//	positive := func(n int) bool { return n > 0 }
//	option.Some(45).Filter(positive) // Some(45)
//	option.Some(-1).Filter(positive) // None
func (o Option[T]) Filter(fn func(T) bool) Option[T] {
	if o.IsNone() || !fn(o.value) {
		return None[T]()
	}

	return o
}

// Or returns this Option if it is Some, and optB otherwise.
//
//	option.Some(45).Or(option.Some(7))    // Some(45)
//	option.None[int]().Or(option.Some(7)) // Some(7)
func (o Option[T]) Or(optB Option[T]) Option[T] {
	if o.IsSome() {
		return o
	}

	return optB
}

// OrElse returns this Option if it is Some, and the result of fn otherwise.
// fn is only called on a None.
//
//	option.Some(45).OrElse(loadDefault)    // Some(45), loadDefault is not called
//	option.None[int]().OrElse(loadDefault) // whatever loadDefault returns
func (o Option[T]) OrElse(fn func() Option[T]) Option[T] {
	if o.IsNone() {
		return fn()
	}

	return o
}

// Xor returns whichever side is Some, or None if both or neither are.
//
//	option.Some(45).Xor(option.None[int]()) // Some(45)
//	option.None[int]().Xor(option.Some(7))  // Some(7)
//	option.Some(45).Xor(option.Some(7))     // None
func (o Option[T]) Xor(optB Option[T]) Option[T] {
	if o.IsSome() && optB.IsNone() {
		return o
	}
	if o.IsNone() && optB.IsSome() {
		return optB
	}

	return None[T]()
}
