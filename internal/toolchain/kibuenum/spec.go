// Package kibuenum proves call-expression based enum discovery without executing
// user code. It is intentionally not wired into the production generators yet.
package kibuenum

import "go/token"

// Result is a target-neutral, serializable candidate for future Package.Enums
// integration in modspecv2. Invalid declarations are omitted; consumers must also
// honor diagnostics before generating artifacts.
type Result struct{ Enums []Enum }

type Symbol struct {
	PackagePath string
	Name        string
	Position    token.Position
}

type Span struct{ Start, End token.Position }

type Enum struct {
	Declaration Symbol
	Type        Symbol
	Underlying  string
	Source      Span
	Members     []Member
}

type Member struct {
	Constant Symbol
	// WireValue is the decoded string for string enums and exact base-10 text
	// for integer enums. No float64/JavaScript number conversion is involved.
	WireValue         string
	Label             string
	Description       string
	Source            Span
	ValueSource       Span
	LabelSource       Span
	DescriptionSource Span
}
