package vo

import (
	"errors"
	"fmt"
)

var ErrNegativeAmount = errors.New("amount cannot be negative")

type Money struct {
	amount   float64
	currency string
}

func NewMoney(amount float64, currency string) (Money, error) {
	if amount < 0 {
		return Money{}, ErrNegativeAmount
	}
	return Money{
		amount:   amount,
		currency: currency,
	}, nil
}

func (m Money) Amount() float64 {
	return m.amount
}

func (m Money) Currency() string {
	return m.currency
}

func (m Money) Add(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, fmt.Errorf("cannot add different currencies: %s and %s", m.currency, other.currency)
	}
	return NewMoney(m.amount+other.amount, m.currency)
}

func (m Money) Subtract(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, fmt.Errorf("cannot subtract different currencies: %s and %s", m.currency, other.currency)
	}
	return NewMoney(m.amount-other.amount, m.currency)
}

func (m Money) Multiply(factor float64) (Money, error) {
	return NewMoney(m.amount*factor, m.currency)
}

func (m Money) IsZero() bool {
	return m.amount == 0
}

func (m Money) String() string {
	return fmt.Sprintf("%.2f %s", m.amount, m.currency)
}
