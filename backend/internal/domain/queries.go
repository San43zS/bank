package domain

type TransactionFilter struct {
	Type  TransactionType
	Page  int
	Limit int
}

type TransactionWithEmails struct {
	Transaction   Transaction
	FromUserEmail *string
	ToUserEmail   *string
}

