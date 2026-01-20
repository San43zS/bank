package domain

import "testing"

func TestCentsToDecimalString(t *testing.T) {
	testCases := []struct {
		name string
		in   int64
		want string
	}{
		{name: "zero", in: 0, want: "0.00"},
		{name: "one_cent", in: 1, want: "0.01"},
		{name: "ten_cents", in: 10, want: "0.10"},
		{name: "one_dollar", in: 100, want: "1.00"},
		{name: "ten_25", in: 1025, want: "10.25"},
		{name: "negative", in: -1025, want: "-10.25"},
		{name: "negative_one_cent", in: -1, want: "-0.01"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := CentsToDecimalString(tc.in); got != tc.want {
				t.Fatalf("got=%q want=%q", got, tc.want)
			}
		})
	}
}

func TestDecimalStringToCents(t *testing.T) {
	testCases := []struct {
		name    string
		in      string
		want    int64
		wantErr bool
	}{
		{name: "zero", in: "0", want: 0},
		{name: "zero_with_spaces", in: "  0  ", want: 0},
		{name: "plus_sign", in: "+10.12", want: 1012},
		{name: "negative", in: "-10.12", want: -1012},
		{name: "integer", in: "10", want: 1000},
		{name: "one_decimal", in: "10.1", want: 1010},
		{name: "two_decimals", in: "10.12", want: 1012},
		{name: "trailing_dot", in: "10.", want: 1000},
		{name: "leading_dot", in: ".5", want: 50},
		{name: "reject_empty", in: "", wantErr: true},
		{name: "reject_spaces", in: "   ", wantErr: true},
		{name: "reject_sign_only", in: "-", wantErr: true},
		{name: "reject_non_number", in: "abc", wantErr: true},
		{name: "reject_multiple_dots", in: "1.2.3", wantErr: true},
		{name: "reject_too_many_decimals", in: "0.001", wantErr: true},
		{name: "reject_negative_whole_part_in_body", in: "--1", wantErr: true},
		{name: "reject_signed_int_part", in: "-1.-2", wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := DecimalStringToCents(tc.in)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err=%v wantErr=%v", err, tc.wantErr)
			}
			if tc.wantErr {
				return
			}
			if got != tc.want {
				t.Fatalf("got=%d want=%d", got, tc.want)
			}
		})
	}
}

