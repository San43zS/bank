package service

import (
	"testing"

	"banking-platform/internal/model"
)

func TestAmountToCents(t *testing.T) {
	tests := []struct {
		name    string
		in      float64
		want    int64
		wantErr bool
	}{
		{name: "zero", in: 0, want: 0, wantErr: false},
		{name: "one_cent", in: 0.01, want: 1, wantErr: false},
		{name: "ten", in: 10.00, want: 1000, wantErr: false},
		{name: "ten_point_one", in: 10.1, want: 1010, wantErr: false},
		{name: "two_decimals", in: 10.12, want: 1012, wantErr: false},
		{name: "three_decimals_rejected", in: 0.001, wantErr: true},
		{name: "three_decimals_rejected2", in: 10.129, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := amountToCents(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err=%v wantErr=%v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got != tt.want {
				t.Fatalf("got=%d want=%d", got, tt.want)
			}
		})
	}
}

func TestConvertExchange(t *testing.T) {
	rateNum := int64(92)
	rateDen := int64(100)
	rate := float64(rateNum) / float64(rateDen)

	tests := []struct {
		name     string
		amount   int64
		from     model.Currency
		to       model.Currency
		wantRate float64
		want     int64
		wantErr  bool
	}{
		{name: "usd_to_eur_100", amount: 10000, from: model.CurrencyUSD, to: model.CurrencyEUR, wantRate: rate, want: 9200},
		{name: "usd_to_eur_rounds_half_up", amount: 2, from: model.CurrencyUSD, to: model.CurrencyEUR, wantRate: rate, want: 2},
		{name: "usd_to_eur_1_cent", amount: 1, from: model.CurrencyUSD, to: model.CurrencyEUR, wantRate: rate, want: 1},
		{name: "eur_to_usd_92", amount: 9200, from: model.CurrencyEUR, to: model.CurrencyUSD, wantRate: 1.0 / rate, want: 10000},
		{name: "eur_to_usd_rounds_half_up", amount: 1, from: model.CurrencyEUR, to: model.CurrencyUSD, wantRate: 1.0 / rate, want: 1},
		{name: "unsupported_pair", amount: 1, from: model.CurrencyUSD, to: "GBP", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRate, got, err := convertExchange(tt.amount, tt.from, tt.to, rateNum, rateDen)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err=%v wantErr=%v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if gotRate != tt.wantRate {
				t.Fatalf("rate=%v want=%v", gotRate, tt.wantRate)
			}
			if got != tt.want {
				t.Fatalf("got=%d want=%d", got, tt.want)
			}
		})
	}
}
