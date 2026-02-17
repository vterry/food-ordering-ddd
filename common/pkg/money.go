package common

import (
	"math"
	"strconv"
)

type Money struct {
	amount int64
}

var Zero = Money{amount: 0}

func NewMoneyFromFloat(amount float64) Money {
	return Money{
		amount: int64(math.Round(amount * 100)),
	}
}

func NewMoneyFromCents(cents int64) Money {
	return Money{
		amount: cents,
	}
}

func (m Money) IsGreaterThanZero() bool {
	return m.amount > 0
}

func (m Money) IsGreaterThan(other Money) bool {
	return m.amount > other.amount
}

func (m Money) Add(other Money) Money {
	return Money{amount: m.amount + other.amount}
}

func (m Money) Subtract(other Money) Money {
	return Money{amount: m.amount - other.amount}
}

func (m Money) Multiply(multiplier int) Money {
	return Money{amount: m.amount * int64(multiplier)}
}

func (m Money) Amount() int64 {
	return m.amount
}

func (m Money) ToFloat() float64 {
	return float64(m.amount) / 100.0
}

func (m Money) String() string {
	return strconv.FormatInt(m.amount, 10)
}
