package parsers

import "fmt"

// ParseResult is returned by Registry.Parse on success.
type ParseResult struct {
	Transactions  []ParsedTransaction
	ParserVersion string // e.g. "piraeus/v1"
	AccountNumber string // bank account number extracted from the statement
}

// Registry holds all registered bank parsers grouped by bank name.
// Versions are tried in registration order; the first that succeeds wins.
type Registry struct {
	parsers map[string][]BankParser
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry {
	return &Registry{parsers: make(map[string][]BankParser)}
}

// Register adds p to the registry under its bank name.
func (r *Registry) Register(p BankParser) {
	name := p.BankName()
	r.parsers[name] = append(r.parsers[name], p)
}

// Parse tries each registered version for bankName in order and returns
// the first successful result. If no parser succeeds, an error is returned.
func (r *Registry) Parse(bankName, pdfPath string) (ParseResult, error) {
	versions, ok := r.parsers[bankName]
	if !ok {
		return ParseResult{}, fmt.Errorf("no parsers registered for bank %q", bankName)
	}
	for _, p := range versions {
		txs, err := p.Parse(pdfPath)
		if err != nil {
			fmt.Printf("[registry] parser %s/%s failed: %v\n", p.BankName(), p.Version(), err)
			continue
		}
		result := ParseResult{
			Transactions:  txs,
			ParserVersion: fmt.Sprintf("%s/%s", p.BankName(), p.Version()),
		}
		for _, tx := range txs {
			if tx.AccountID != "" {
				result.AccountNumber = tx.AccountID
				break
			}
		}
		return result, nil
	}
	return ParseResult{}, fmt.Errorf("all parsers for bank %q failed on %q", bankName, pdfPath)
}
