package mockavia

import (
	"context"
	"strings"

	"flysoft-flight-service/internal/providers"
)

const supplierName = "MockAvia"

type Provider struct{}

type supplierPayload struct {
	Supplier string          `json:"supplier"`
	Routes   []supplierRoute `json:"routes"`
}

type supplierRoute struct {
	ID          string `json:"id"`
	Dep         string `json:"dep"`
	Arr         string `json:"arr"`
	Carrier     string `json:"carrier"`
	Flt         string `json:"flt"`
	PriceAdult  int64  `json:"price_adult"`
	PriceChild  int64  `json:"price_child"`
	PriceInfant int64  `json:"price_infant"`
	Currency    string `json:"currency"`
}

var _ providers.FlightProvider = Provider{}

var fixture = supplierPayload{
	Supplier: supplierName,
	Routes: []supplierRoute{
		{
			ID:          "SUP-001",
			Dep:         "DYU",
			Arr:         "IST",
			Carrier:     "TK",
			Flt:         "255",
			PriceAdult:  30000,
			PriceChild:  20000,
			PriceInfant: 0,
			Currency:    "USD",
		},
		{
			ID:          "SUP-002",
			Dep:         "IST",
			Arr:         "DYU",
			Carrier:     "TK",
			Flt:         "254",
			PriceAdult:  32000,
			PriceChild:  22000,
			PriceInfant: 0,
			Currency:    "USD",
		},
		{
			ID:          "SUP-003",
			Dep:         "DYU",
			Arr:         "DXB",
			Carrier:     "FZ",
			Flt:         "778",
			PriceAdult:  25000,
			PriceChild:  18000,
			PriceInfant: 5000,
			Currency:    "USD",
		},
	},
}

func New() Provider {
	return Provider{}
}

func (Provider) Name() string {
	return supplierName
}

func (Provider) Search(ctx context.Context, req providers.ProviderSearch) ([]providers.ProviderOffer, error) {
	offers := make([]providers.ProviderOffer, 0, len(fixture.Routes))

	for _, route := range fixture.Routes {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if !matchesRoute(route, req) {
			continue
		}

		offers = append(offers, normalize(route))
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return offers, nil
}

func matchesRoute(route supplierRoute, req providers.ProviderSearch) bool {
	from := strings.TrimSpace(req.From)
	to := strings.TrimSpace(req.To)

	if from != "" && !strings.EqualFold(from, route.Dep) {
		return false
	}
	if to != "" && !strings.EqualFold(to, route.Arr) {
		return false
	}
	return true
}

func normalize(route supplierRoute) providers.ProviderOffer {
	return providers.ProviderOffer{
		ProviderRef:  route.ID,
		From:         route.Dep,
		To:           route.Arr,
		Airline:      route.Carrier,
		FlightNumber: route.Carrier + route.Flt,
		Currency:     route.Currency,
		PriceAdult:   route.PriceAdult,
		PriceChild:   route.PriceChild,
		PriceInfant:  route.PriceInfant,
	}
}
