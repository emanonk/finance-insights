package parser

// ParsedTransaction is the raw, unnormalized output of the PDF parser.
// Fields are strings because parsing preserves the source representation;
// conversion to domain types happens in the service layer.
type ParsedTransaction struct {
	AccountID               string
	Date                    string
	BankReferenceNumber     string
	Justification           string
	Indicator               string
	MerchantIdentifier      string
	Amount1                 string
	MCCCode                 string
	CardMasked              string
	Reference               string
	Description             string
	PaymentMethod           string
	Direction               string
	Amount                  string
	BalanceAfterTransaction string
}
