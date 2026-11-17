package function

import (
	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/numtext"
)

const maxListItems = 1000

type typeInfo struct {
	kind    numtext.Type
	element numtext.Type
	list    bool
	numeric bool
	words   string
}

var typeInfos = map[formula.Type]typeInfo{
	formula.TypeInteger: {kind: numtext.Integer, numeric: true, words: "a whole number with at most 18 digits"},
	formula.TypeDecimal: {kind: numtext.Decimal, numeric: true, words: "a decimal number with at most 30 digits, such as 12.50"},
	formula.TypeFloat:   {kind: numtext.Float, numeric: true, words: "a number with at most 30 digits, such as 0.37"},
	formula.TypeText:    {kind: numtext.Text, words: "text of at most 200 characters"},
	formula.TypeBoolean: {kind: numtext.Boolean, words: "true or false"},
	formula.TypeDate:    {kind: numtext.Date, words: "a date written like 2026-03-31"},
	formula.TypeListInteger: {
		kind: numtext.IntegerList, element: numtext.Integer, list: true, numeric: true, words: "a list of whole numbers",
	},
	formula.TypeListDecimal: {
		kind: numtext.DecimalList, element: numtext.Decimal, list: true, numeric: true, words: "a list of decimal numbers",
	},
}

func infoOf(kind formula.Type) (typeInfo, bool) {
	info, known := typeInfos[kind]
	return info, known
}

func (i typeInfo) itemWords() string {
	return typeInfos[elementType(i.element)].words
}

func elementType(element numtext.Type) formula.Type {
	if element == numtext.Integer {
		return formula.TypeInteger
	}
	return formula.TypeDecimal
}
