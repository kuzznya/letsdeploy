package validations

import "github.com/pkg/errors"

type MustBeOr[T any] interface {
	OrPanic()
	OrPanicWithMessage(msg string)
	OrElse(other T) T
	OrElseGet(getter func() T) T
}

func NotEmptyString(value string) bool {
	return len(value) != 0
}

func NotEmptySlice[T any](value []T) bool {
	return len(value) != 0
}

func NotEmptyMap[K comparable, V any](value map[K]V) bool {
	return len(value) != 0
}

func MustBe[T any](validation func(v T) bool) func(v T) MustBeOr[T] {
	return func(v T) MustBeOr[T] {
		return &mustBeOrImpl[T]{value: v, result: validation(v)}
	}
}

type mustBeOrImpl[T any] struct {
	value  T
	result bool
}

func (m *mustBeOrImpl[T]) OrPanic() {
	if !m.result {
		panic(errors.New("value check failed"))
	}
}

func (m *mustBeOrImpl[T]) OrPanicWithMessage(msg string) {
	if !m.result {
		panic(errors.New(msg))
	}
}

func (m *mustBeOrImpl[T]) OrElse(other T) T {
	if m.result {
		return m.value
	} else {
		return other
	}
}

func (m *mustBeOrImpl[T]) OrElseGet(getter func() T) T {
	if m.result {
		return m.value
	} else {
		return getter()
	}
}
