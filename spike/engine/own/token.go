package own

import "fmt"

type tokenKind int

const (
	tokenEnd tokenKind = iota
	tokenInteger
	tokenDecimal
	tokenFloat
	tokenDate
	tokenText
	tokenName
	tokenSymbol
)

type token struct {
	kind tokenKind
	text string
	pos  Pos
}

func describe(t token) string {
	switch t.kind {
	case tokenEnd:
		return "the end of the formula"
	case tokenText:
		return fmt.Sprintf("the text %q", t.text)
	}
	return fmt.Sprintf("%q", t.text)
}
