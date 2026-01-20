package domain

import (
	"fmt"
	"strconv"
	"strings"
)

// Format cents as "12.34".
func CentsToDecimalString(cents int64) string {
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	return fmt.Sprintf("%s%d.%02d", sign, cents/100, cents%100)
}

// Parse a decimal string into cents (strict: max 2 decimals).
func DecimalStringToCents(s string) (int64, error) {
	raw := strings.TrimSpace(s)
	if raw == "" {
		return 0, fmt.Errorf("empty amount")
	}

	sign := int64(1)
	if raw[0] == '-' {
		sign = -1
		raw = strings.TrimSpace(raw[1:])
		if raw == "" {
			return 0, fmt.Errorf("invalid amount")
		}
	} else if raw[0] == '+' {
		raw = strings.TrimSpace(raw[1:])
		if raw == "" {
			return 0, fmt.Errorf("invalid amount")
		}
	}

	parts := strings.Split(raw, ".")
	if len(parts) > 2 {
		return 0, fmt.Errorf("invalid amount")
	}

	intPart := parts[0]
	if intPart == "" {
		intPart = "0"
	}
	if strings.HasPrefix(intPart, "+") || strings.HasPrefix(intPart, "-") {
		return 0, fmt.Errorf("invalid amount")
	}
	whole, err := strconv.ParseInt(intPart, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid amount")
	}

	var frac int64
	if len(parts) == 2 {
		fp := parts[1]
		if strings.HasPrefix(fp, "+") || strings.HasPrefix(fp, "-") {
			return 0, fmt.Errorf("invalid amount")
		}
		if len(fp) > 2 {
			return 0, fmt.Errorf("amount has more than 2 decimals")
		}
		if fp == "" {
			fp = "0"
		}
		if len(fp) == 1 {
			fp = fp + "0"
		}
		frac, err = strconv.ParseInt(fp, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid amount")
		}
	} else {
		frac = 0
	}

	if whole < 0 {
		return 0, fmt.Errorf("invalid amount")
	}

	return sign * (whole*100 + frac), nil
}

