package pricing

import "flysoft-flight-service/internal/money"

type Calculator struct {
	CommissionPercent int64
	ServiceFees       ServiceFees
}

type Config struct {
	CommissionPercent int64
	ServiceFees       ServiceFees
}

type ServiceFees struct {
	Adult  money.Money
	Child  money.Money
	Infant money.Money
}

type PassengerCounts struct {
	Adults   int64
	Children int64
	Infants  int64
}

type PassengerPrices struct {
	Adult  money.Money
	Child  money.Money
	Infant money.Money
}

type Breakdown struct {
	Base       money.Money
	Commission money.Money
	ServiceFee money.Money
	Total      money.Money
	Profit     money.Money
}

func NewCalculator(cfg Config) Calculator {
	return Calculator{
		CommissionPercent: cfg.CommissionPercent,
		ServiceFees:       cfg.ServiceFees,
	}
}

func (c Calculator) Calculate(passengers PassengerCounts, prices PassengerPrices) Breakdown {
	base := passengers.Adults*prices.Adult.Cents() +
		passengers.Children*prices.Child.Cents() +
		passengers.Infants*prices.Infant.Cents()

	commission := (base*c.CommissionPercent + 50) / 100

	serviceFee := passengers.Adults*c.ServiceFees.Adult.Cents() +
		passengers.Children*c.ServiceFees.Child.Cents() +
		passengers.Infants*c.ServiceFees.Infant.Cents()

	total := base + commission + serviceFee
	profit := commission + serviceFee

	return Breakdown{
		Base:       money.FromCents(base),
		Commission: money.FromCents(commission),
		ServiceFee: money.FromCents(serviceFee),
		Total:      money.FromCents(total),
		Profit:     money.FromCents(profit),
	}
}
