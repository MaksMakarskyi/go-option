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
func Some[T any](value T) Option[T] {
	return Option[T]{
		value:   value,
		present: true,
	}
}

// None returns an empty Option. It is identical to the zero value of the type.
func None[T any]() Option[T] {
	return Option[T]{}
}

// FromTuple turns a (value, ok) pair into an Option.
// When ok is false the value is discarded, so the result equals None.
func FromTuple[T any](value T, ok bool) Option[T] {
	if !ok {
		return None[T]()
	}

	return Some(value)
}

// FromPtr turns a pointer into an Option: nil becomes None, anything else
// becomes Some holding a copy of the pointed-to value.
func FromPtr[T any](ptr *T) Option[T] {
	if ptr == nil {
		return None[T]()
	}

	return Some(*ptr)
}

// ToPtr returns a pointer to a copy of the value, or nil for a None.
// Writing through that pointer does not change the Option.
func (o Option[T]) ToPtr() *T {
	if !o.present {
		return nil
	}

	return &o.value
}

// IsSome reports whether a value is present.
func (o Option[T]) IsSome() bool {
	return o.present
}

// IsNone reports whether the Option is empty.
func (o Option[T]) IsNone() bool {
	return !o.present
}

// IsSomeAnd reports whether a value is present and satisfies pred.
// pred is not called on a None.
func (o Option[T]) IsSomeAnd(pred func(T) bool) bool {
	if o.IsNone() {
		return false
	}

	return pred(o.value)
}

// Get returns the value and whether it was present, in Go's usual comma-ok style.
func (o Option[T]) Get() (T, bool) {
	return o.value, o.present
}

// Expect returns the value, and panics with msg if there is none.
// The panic value wraps ErrValueIsNone.
func (o Option[T]) Expect(msg string) T {
	if o.IsNone() {
		panic(fmt.Errorf("%w: %s", ErrValueIsNone, msg))
	}

	return o.value
}

// Unwrap returns the value and panics with ErrValueIsNone if there is none.
// Reach for it only when a value is guaranteed.
func (o Option[T]) Unwrap() T {
	if o.IsNone() {
		panic(ErrValueIsNone)
	}

	return o.value
}

// UnwrapOr returns the value, or the given fallback if the Option is None.
func (o Option[T]) UnwrapOr(value T) T {
	if o.IsNone() {
		return value
	}

	return o.value
}

// UnwrapOrElse returns the value, or the result of fn if the Option is None.
// fn is only called on a None, so an expensive fallback stays cheap.
func (o Option[T]) UnwrapOrElse(fn func() T) T {
	if o.IsNone() {
		return fn()
	}

	return o.value
}

// UnwrapOrZero returns the value, or the zero value of T if there is none.
func (o Option[T]) UnwrapOrZero() T {
	if o.IsNone() {
		var zero T
		return zero
	}

	return o.value
}

// OkOr returns the value, or err joined with ErrValueIsNone if there is none,
// so errors.Is matches either one. A nil err gives ErrValueIsNone on its own.
// Handy for feeding an Option into ordinary error handling.
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
// wrap ErrValueIsNone unless fn puts it there.
func (o Option[T]) OkOrElse(fn func() error) (T, error) {
	if o.IsNone() {
		var zero T
		return zero, fn()
	}

	return o.value, nil
}

// Map applies fn to the value and wraps the result in a new Option.
// A None passes straight through and fn is not called.
func (o Option[T]) Map[U any](fn func(T) U) Option[U] {
	if o.IsNone() {
		return Option[U]{}
	}

	return Some(fn(o.value))
}

// MapOr applies fn to the value, or returns the given fallback for a None.
// fn is not called on a None.
func (o Option[T]) MapOr[U any](value U, fn func(T) U) U {
	if o.IsNone() {
		return value
	}

	return fn(o.value)
}

// And returns optB when this Option is Some, and None otherwise.
func (o Option[T]) And[U any](optB Option[U]) Option[U] {
	if o.IsNone() {
		return None[U]()
	}

	return optB
}

// AndThen calls fn on the value and returns its result, so steps that may
// each come up empty can be chained. fn is not called on a None.
func (o Option[T]) AndThen[U any](fn func(T) Option[U]) Option[U] {
	if o.IsNone() {
		return None[U]()
	}

	return fn(o.value)
}

// Filter keeps the value only if it satisfies fn, and returns None otherwise.
// fn is not called on a None.
func (o Option[T]) Filter(fn func(T) bool) Option[T] {
	if o.IsNone() || !fn(o.value) {
		return None[T]()
	}

	return o
}

// Or returns this Option if it is Some, and optB otherwise.
func (o Option[T]) Or(optB Option[T]) Option[T] {
	if o.IsSome() {
		return o
	}

	return optB
}

// OrElse returns this Option if it is Some, and the result of fn otherwise.
// fn is only called on a None.
func (o Option[T]) OrElse(fn func() Option[T]) Option[T] {
	if o.IsNone() {
		return fn()
	}

	return o
}

// Xor returns whichever side is Some, or None if both or neither are.
func (o Option[T]) Xor(optB Option[T]) Option[T] {
	if o.IsSome() && optB.IsNone() {
		return o
	}
	if o.IsNone() && optB.IsSome() {
		return optB
	}

	return None[T]()
}
