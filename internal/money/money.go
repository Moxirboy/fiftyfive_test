package money

import "fmt"

type Money int64

func FromCents(cents int64) Money {
	return Money(cents)
}

func (m Money) Add(other Money) Money {
	return m + other
}

func (m Money) Mul(multiplier int64) Money {
	return Money(int64(m) * multiplier)
}

func (m Money) Cents() int64 {
	return int64(m)
}

func (m Money) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%.2f", float64(m)/100)), nil
}
