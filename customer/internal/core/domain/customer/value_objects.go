package customer

import (
	"errors"
	"regexp"
)

var (
	ErrInvalidEmail = errors.New("invalid email format")
	ErrInvalidPhone = errors.New("invalid phone format")
	ErrEmptyName    = errors.New("name cannot be empty")
)

type Name struct {
	value string
}

func NewName(value string) (Name, error) {
	if value == "" {
		return Name{}, ErrEmptyName
	}
	return Name{value: value}, nil
}

func (n Name) String() string {
	return n.value
}

type Email struct {
	value string
}

var emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)

func NewEmail(value string) (Email, error) {
	if !emailRegex.MatchString(value) {
		return Email{}, ErrInvalidEmail
	}
	return Email{value: value}, nil
}

func (e Email) String() string {
	return e.value
}

type Phone struct {
	value string
}

var phoneRegex = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)

func NewPhone(value string) (Phone, error) {
	if !phoneRegex.MatchString(value) {
		return Phone{}, ErrInvalidPhone
	}
	return Phone{value: value}, nil
}

func (p Phone) String() string {
	return p.value
}
