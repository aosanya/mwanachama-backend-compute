package functions

import (
	"math"
	"strconv"
	"strings"

	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/function"
	"github.com/aosanya/mwanachama-backend-compute/numtext"
)

func RowOf(c Case) function.Row {
	row := function.Row{}
	for name, value := range c.Inputs {
		if value.List {
			row[name] = numtext.Many(value.Items...)
		} else {
			row[name] = numtext.One(value.Text)
		}
	}
	return row
}

func Judge(key string, c Case, out Outcome) string {
	if c.Expected == RefusedText {
		if !out.Refused() {
			return c.ID + ": " + key + " gave " + shown(out) + ", the case records refused."
		}
		return ""
	}
	if out.Refused() {
		return c.ID + ": " + key + " gave no result (" + out.Sentence() + "), the case records " + c.Expected + "."
	}
	if matches(c, out) {
		return ""
	}
	return c.ID + ": " + key + " gave " + shown(out) + ", the case records " + c.Expected + "."
}

func shown(out Outcome) string {
	if out.Refused() {
		return "no result"
	}
	return out.Value.String()
}

func matches(c Case, out Outcome) bool {
	if out.Value.Type() != formula.TypeFloat {
		return out.Value.String() == c.Expected
	}
	expected, err := strconv.ParseFloat(strings.TrimSuffix(c.Expected, "f"), 64)
	if err != nil {
		return false
	}
	got, err := strconv.ParseFloat(out.Value.String(), 64)
	if err != nil {
		return false
	}
	if c.Tolerance != ToleranceRelative || expected == 0 {
		return got == expected
	}
	return math.Abs(got-expected) <= 1e-9*math.Abs(expected)
}
