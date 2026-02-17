package common

import (
	"errors"
	"fmt"
)

type Specification[T any] func(*T) error

func And[T any](specs ...Specification[T]) Specification[T] {
	return func(target *T) error {
		var errs error
		for _, spec := range specs {
			if err := spec(target); err != nil {
				errs = errors.Join(errs, err)
			}
		}
		return errs
	}
}

func Or[T any](specs ...Specification[T]) Specification[T] {
	return func(target *T) error {
		var allErrors error
		for _, spec := range specs {
			err := spec(target)
			if err == nil {
				return nil
			}
			allErrors = errors.Join(allErrors, err)
		}
		return fmt.Errorf("at least one rule must be satisfied: %w", allErrors)
	}
}

func Not[T any](spec Specification[T]) Specification[T] {
	return func(target *T) error {
		if err := spec(target); err != nil {
			return nil
		}
		return errors.New("condition should not be met")
	}
}
