package service

import "testing"

func TestParseRateToFraction(t *testing.T) {
	testCases := []struct {
		name    string
		in      string
		wantNum int64
		wantDen int64
	}{
		{name: "empty_defaults", in: "", wantNum: 92, wantDen: 100},
		{name: "spaces_defaults", in: "   ", wantNum: 92, wantDen: 100},
		{name: "invalid_defaults", in: "nope", wantNum: 92, wantDen: 100},
		{name: "zero_defaults", in: "0", wantNum: 92, wantDen: 100},
		{name: "negative_defaults", in: "-1", wantNum: 92, wantDen: 100},
		{name: "simple_decimal_scaled", in: "0.92", wantNum: 920000, wantDen: 1000000},
		{name: "integer_scaled", in: "1", wantNum: 1000000, wantDen: 1000000},
		{name: "six_decimals_scaled", in: "0.123456", wantNum: 123456, wantDen: 1000000},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotNum, gotDen := parseRateToFraction(tc.in)
			if gotNum != tc.wantNum || gotDen != tc.wantDen {
				t.Fatalf("got=%d/%d want=%d/%d", gotNum, gotDen, tc.wantNum, tc.wantDen)
			}
		})
	}
}

func TestConvertExchange_InvalidRate(t *testing.T) {
	testCases := []struct {
		name    string
		num     int64
		den     int64
		wantErr bool
	}{
		{name: "zero_num", num: 0, den: 100, wantErr: true},
		{name: "zero_den", num: 92, den: 0, wantErr: true},
		{name: "negative_num", num: -1, den: 100, wantErr: true},
		{name: "negative_den", num: 92, den: -1, wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := convertExchange(100, "USD", "EUR", tc.num, tc.den)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err=%v wantErr=%v", err, tc.wantErr)
			}
		})
	}
}

