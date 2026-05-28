package dto_test

import (
	"testing"
	"time"

	"flysoft-flight-service/internal/apperror"
	"flysoft-flight-service/internal/dto"
)

func TestSearchRequestValidate(t *testing.T) {
	valid := func() dto.SearchRequest {
		return dto.SearchRequest{
			From:          "DYU",
			To:            "IST",
			DepartureDate: dateFromNow(1),
			ReturnDate:    dateFromNow(10),
			Adults:        1,
			Children:      1,
			Infants:       0,
			Currency:      "USD",
		}
	}

	tests := []struct {
		name    string
		request dto.SearchRequest
		wantErr bool
	}{
		{
			name:    "valid request",
			request: valid(),
		},
		{
			name: "from required",
			request: func() dto.SearchRequest {
				req := valid()
				req.From = ""
				return req
			}(),
			wantErr: true,
		},
		{
			name: "to required",
			request: func() dto.SearchRequest {
				req := valid()
				req.To = " "
				return req
			}(),
			wantErr: true,
		},
		{
			name: "from and to must differ",
			request: func() dto.SearchRequest {
				req := valid()
				req.To = "dyu"
				return req
			}(),
			wantErr: true,
		},
		{
			name: "departure date parseable",
			request: func() dto.SearchRequest {
				req := valid()
				req.DepartureDate = "2026/06/15"
				return req
			}(),
			wantErr: true,
		},
		{
			name: "departure date not in past",
			request: func() dto.SearchRequest {
				req := valid()
				req.DepartureDate = dateFromNow(-1)
				return req
			}(),
			wantErr: true,
		},
		{
			name: "total passengers greater than zero",
			request: func() dto.SearchRequest {
				req := valid()
				req.Adults = 0
				req.Children = 0
				req.Infants = 0
				return req
			}(),
			wantErr: true,
		},
		{
			name: "passenger counts non-negative",
			request: func() dto.SearchRequest {
				req := valid()
				req.Children = -1
				return req
			}(),
			wantErr: true,
		},
		{
			name: "infants do not exceed adults",
			request: func() dto.SearchRequest {
				req := valid()
				req.Adults = 1
				req.Infants = 2
				return req
			}(),
			wantErr: true,
		},
		{
			name: "currency required",
			request: func() dto.SearchRequest {
				req := valid()
				req.Currency = ""
				return req
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			assertValidationResult(t, err, tt.wantErr)
		})
	}
}

func TestBookingRequestValidate(t *testing.T) {
	valid := func() dto.BookingRequest {
		return dto.BookingRequest{
			OfferID: "OF-123456",
			Passengers: []dto.PassengerDTO{
				{
					Type:           "adult",
					FirstName:      "Alisher",
					LastName:       "Sabirov",
					DocumentNumber: "A1234567",
				},
			},
		}
	}

	tests := []struct {
		name    string
		request dto.BookingRequest
		wantErr bool
	}{
		{
			name:    "valid request",
			request: valid(),
		},
		{
			name: "offer_id required",
			request: func() dto.BookingRequest {
				req := valid()
				req.OfferID = " "
				return req
			}(),
			wantErr: true,
		},
		{
			name: "passengers non-empty",
			request: func() dto.BookingRequest {
				req := valid()
				req.Passengers = nil
				return req
			}(),
			wantErr: true,
		},
		{
			name: "passenger type valid",
			request: func() dto.BookingRequest {
				req := valid()
				req.Passengers[0].Type = "senior"
				return req
			}(),
			wantErr: true,
		},
		{
			name: "passenger first_name required",
			request: func() dto.BookingRequest {
				req := valid()
				req.Passengers[0].FirstName = ""
				return req
			}(),
			wantErr: true,
		},
		{
			name: "passenger last_name required",
			request: func() dto.BookingRequest {
				req := valid()
				req.Passengers[0].LastName = " "
				return req
			}(),
			wantErr: true,
		},
		{
			name: "passenger document_number required",
			request: func() dto.BookingRequest {
				req := valid()
				req.Passengers[0].DocumentNumber = ""
				return req
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			assertValidationResult(t, err, tt.wantErr)
		})
	}
}

func assertValidationResult(t *testing.T, err *apperror.AppError, wantErr bool) {
	t.Helper()

	if !wantErr {
		if err != nil {
			t.Fatalf("Validate() error = %v, want nil", err)
		}
		return
	}

	if err == nil {
		t.Fatal("Validate() error = nil, want validation error")
	}
	if err.Code != apperror.CodeValidationError {
		t.Fatalf("Validate() code = %s, want %s", err.Code, apperror.CodeValidationError)
	}
	if err.HTTPStatus != 400 {
		t.Fatalf("Validate() HTTPStatus = %d, want 400", err.HTTPStatus)
	}
}

func dateFromNow(days int) string {
	return time.Now().AddDate(0, 0, days).Format("2006-01-02")
}
