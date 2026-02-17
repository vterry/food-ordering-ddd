package common

import (
	"fmt"
)

type BaseID[T comparable] struct {
	value T
}

func NewBaseID[T comparable](value T) BaseID[T] {
	return BaseID[T]{value: value}
}

func (b BaseID[T]) ID() T {
	return b.value
}

func (b BaseID[T]) Equals(object any) bool {
	if o, ok := object.(BaseID[T]); ok {
		return b.value == o.value
	}
	type identifiable interface {
		ID() T
	}
	if o, ok := object.(identifiable); ok {
		return b.value == o.ID()
	}
	return false
}

func (b *BaseID[T]) String() string {
	return fmt.Sprintf("%v", b.value)
}

func (b *BaseID[T]) IsZero() bool {
	var zero T
	return b.value == zero
}
