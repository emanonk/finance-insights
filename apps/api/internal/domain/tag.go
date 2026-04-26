package domain

// Tag is a label applied to transactions via merchant rules.
type Tag struct {
	ID   int64
	Name string
	Type string // "primary" or "secondary"
}
