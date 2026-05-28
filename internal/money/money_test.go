package money_test

import (
	"encoding/json"
	"testing"

	"flysoft-flight-service/internal/money"
)

func TestMoneyHelpers(t *testing.T) {
	got := money.FromCents(1250).
		Add(money.FromCents(375)).
		Mul(2)

	if got.Cents() != 3250 {
		t.Fatalf("Cents() = %d, want 3250", got.Cents())
	}
}

func TestMoneyMarshalJSON(t *testing.T) {
	got, err := json.Marshal(money.FromCents(46600))
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	if string(got) != "466.00" {
		t.Fatalf("MarshalJSON() = %s, want 466.00", got)
	}
}
