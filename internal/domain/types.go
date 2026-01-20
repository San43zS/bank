package domain

type Currency string

const (
	CurrencyUSD Currency = "USD"
	CurrencyEUR Currency = "EUR"
)

type TransactionType string

const (
	TransactionTypeTransfer TransactionType = "transfer"
	TransactionTypeExchange TransactionType = "exchange"
)

