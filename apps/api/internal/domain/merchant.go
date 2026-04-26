package domain

// Merchant links a bank-assigned identifier to user-defined tags and a display title.
type Merchant struct {
	ID             int64
	IdentifierName string
	PrimaryTag     Tag
	SecondaryTags  []Tag
	DefaultTitle   *string
}

// IdentifierCount pairs a merchant identifier with how many times it appears
// in the transactions table and the optional Merchant record for it.
type IdentifierCount struct {
	Identifier string
	Count      int
	Merchant   *Merchant
}
