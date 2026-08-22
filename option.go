package option

import "fmt"

type Option[T any] struct {
	value   T
	present bool
}

func Some[T any](value T) Option[T] {
	return Option[T]{
		value:   value,
		present: true,
	}
}

func None[T any]() Option[T] {
	return Option[T]{}
}

func FromTuple[T any](value T, ok bool) Option[T] {
	if !ok {
		return None[T]()
	}

	return Some(value)
}

func FromPtr[T any](ptr *T) Option[T] {
	if ptr == nil {
		return None[T]()
	}

	return Some(*ptr)
}

func (o Option[T]) ToPtr() *T {
	if !o.present {
		return nil
	}

	return &o.value
}

func (o Option[T]) IsSome() bool {
	return o.present
}

func (o Option[T]) IsNone() bool {
	return !o.present
}

func (o Option[T]) IsSomeAnd(pred func(T) bool) bool {
	if o.IsNone() {
		return false
	}

	return pred(o.value)
}

func (o Option[T]) Get() (T, bool) {
	return o.value, o.present
}

func (o Option[T]) Expect(msg string) {
	if o.IsNone() {
		panic(fmt.Errorf("%w: %s", ErrValueIsNone, msg))
	}
}

func (o Option[T]) Unwrap() T {
	if o.IsNone() {
		panic(ErrValueIsNone)
	}

	return o.value
}

func (o Option[T]) UnwrapOr(value T) T {
	if o.IsNone() {
		return value
	}

	return o.value
}

func (o Option[T]) UnwrapOrElse(fn func() T) T {
	if o.IsNone() {
		return fn()
	}

	return o.value
}

func (o Option[T]) UnwrapOrZero() T {
	if o.IsNone() {
		var zero T
		return zero
	}

	return o.value
}

func (o Option[T]) UnwrapOrErr() (T, error) {
	if o.IsNone() {
		var zero T
		return zero, ErrValueIsNone
	}

	return o.value, nil
}

func (o Option[T]) MapSome[U any](fn func(T) U) Option[U] {
	if o.IsNone() {
		return Option[U]{}
	}

	return Some(fn(o.value))
}

func (o Option[T]) MapOr[U any](value U, fn func(T) U) U {
	if o.IsNone() {
		return value
	}

	return fn(o.value)
}

func (o Option[T]) And[U any](optB Option[U]) Option[U] {
	if o.IsNone() {
		return None[U]()
	}

	return optB
}

func (o Option[T]) AndThen[U any](fn func(T) Option[U]) Option[U] {
	if o.IsNone() {
		return None[U]()
	}

	return fn(o.value)
}

func (o Option[T]) Filter(fn func(T) bool) Option[T] {
	if o.IsNone() || !fn(o.value) {
		return None[T]()
	}

	return o
}

func (o Option[T]) Or(optB Option[T]) Option[T] {
	if o.IsSome() {
		return o
	}

	return optB
}

func (o Option[T]) OrElse(fn func() Option[T]) Option[T] {
	if o.IsNone() {
		return fn()
	}

	return o
}

func (o Option[T]) Xor(optB Option[T]) Option[T] {
	if o.IsSome() && optB.IsNone() {
		return o
	}
	if o.IsNone() && optB.IsSome() {
		return optB
	}

	return None[T]()
}
