package domain

// TagSpend is a single row in a spend-by-tag report.
type TagSpend struct {
	TagName string
	Total   string // numeric string, e.g. "1234.56"
	Count   int
}

// MerchantMonthRow is one (merchant, month) bucket from the monthly merchant report.
type MerchantMonthRow struct {
	Identifier string
	Month      string // "YYYY-MM"
	Total      string
	MaxAmount  string
	AvgAmount  string
	Count      int
}

// DailySpend is a single day's debit total.
type DailySpend struct {
	Date  string // "YYYY-MM-DD"
	Total string
}

// RecurringCharge is a detected recurring debit pattern for a single merchant.
type RecurringCharge struct {
	CalendarYear int
	Identifier   string
	Amount       string
	Occurrences  int
	TotalDebited string
	FirstDate    string // "YYYY-MM-DD"
	LastDate     string // "YYYY-MM-DD"
}
