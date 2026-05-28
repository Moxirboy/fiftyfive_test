package pricing_test

import (
	"testing"

	"flysoft-flight-service/internal/money"
	"flysoft-flight-service/internal/pricing"
)

func TestCalculatorCalculate(t *testing.T) {
	tests := []struct {
		name       string
		config     pricing.Config
		passengers pricing.PassengerCounts
		prices     pricing.PassengerPrices
		want       pricing.Breakdown
	}{
		{
			name: "spec worked example",
			config: pricing.Config{
				CommissionPercent: 5,
				ServiceFees: pricing.ServiceFees{
					Adult:  money.FromCents(1500),
					Child:  money.FromCents(1000),
					Infant: money.FromCents(0),
				},
			},
			passengers: pricing.PassengerCounts{
				Adults:   1,
				Children: 1,
			},
			prices: pricing.PassengerPrices{
				Adult: money.FromCents(30000),
				Child: money.FromCents(20000),
			},
			want: pricing.Breakdown{
				Base:       money.FromCents(50000),
				Commission: money.FromCents(2500),
				ServiceFee: money.FromCents(2500),
				Total:      money.FromCents(55000),
				Profit:     money.FromCents(5000),
			},
		},
		{
			name: "infants use configured price and fee",
			config: pricing.Config{
				CommissionPercent: 10,
				ServiceFees: pricing.ServiceFees{
					Adult:  money.FromCents(100),
					Child:  money.FromCents(0),
					Infant: money.FromCents(25),
				},
			},
			passengers: pricing.PassengerCounts{
				Adults:  1,
				Infants: 1,
			},
			prices: pricing.PassengerPrices{
				Adult:  money.FromCents(10000),
				Infant: money.FromCents(2500),
			},
			want: pricing.Breakdown{
				Base:       money.FromCents(12500),
				Commission: money.FromCents(1250),
				ServiceFee: money.FromCents(125),
				Total:      money.FromCents(13875),
				Profit:     money.FromCents(1375),
			},
		},
		{
			name: "commission rounds half up",
			config: pricing.Config{
				CommissionPercent: 5,
				ServiceFees:       pricing.ServiceFees{},
			},
			passengers: pricing.PassengerCounts{
				Adults: 1,
			},
			prices: pricing.PassengerPrices{
				Adult: money.FromCents(110),
			},
			want: pricing.Breakdown{
				Base:       money.FromCents(110),
				Commission: money.FromCents(6),
				ServiceFee: money.FromCents(0),
				Total:      money.FromCents(116),
				Profit:     money.FromCents(6),
			},
		},
		{
			name: "commission rounds below half down",
			config: pricing.Config{
				CommissionPercent: 5,
				ServiceFees:       pricing.ServiceFees{},
			},
			passengers: pricing.PassengerCounts{
				Adults: 1,
			},
			prices: pricing.PassengerPrices{
				Adult: money.FromCents(109),
			},
			want: pricing.Breakdown{
				Base:       money.FromCents(109),
				Commission: money.FromCents(5),
				ServiceFee: money.FromCents(0),
				Total:      money.FromCents(114),
				Profit:     money.FromCents(5),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calculator := pricing.NewCalculator(tt.config)
			got := calculator.Calculate(tt.passengers, tt.prices)

			assertBreakdown(t, got, tt.want)
		})
	}
}

func assertBreakdown(t *testing.T, got pricing.Breakdown, want pricing.Breakdown) {
	t.Helper()

	if got.Base.Cents() != want.Base.Cents() {
		t.Fatalf("Base = %d, want %d", got.Base.Cents(), want.Base.Cents())
	}
	if got.Commission.Cents() != want.Commission.Cents() {
		t.Fatalf("Commission = %d, want %d", got.Commission.Cents(), want.Commission.Cents())
	}
	if got.ServiceFee.Cents() != want.ServiceFee.Cents() {
		t.Fatalf("ServiceFee = %d, want %d", got.ServiceFee.Cents(), want.ServiceFee.Cents())
	}
	if got.Total.Cents() != want.Total.Cents() {
		t.Fatalf("Total = %d, want %d", got.Total.Cents(), want.Total.Cents())
	}
	if got.Profit.Cents() != want.Profit.Cents() {
		t.Fatalf("Profit = %d, want %d", got.Profit.Cents(), want.Profit.Cents())
	}
}
