package own

import (
	"strings"
	"testing"
)

func kindsAndTexts(t *testing.T, source string) []string {
	t.Helper()
	tokens, err := tokenize(source)
	if err != nil {
		t.Fatalf("tokenize %q: %v", source, err)
	}
	var out []string
	for _, tok := range tokens {
		out = append(out, strings.TrimSpace(tokenNames[tok.kind]+" "+tok.text))
	}
	return out
}

var tokenNames = map[tokenKind]string{
	tokenEnd: "end", tokenInteger: "int", tokenDecimal: "dec", tokenFloat: "float",
	tokenDate: "date", tokenText: "text", tokenName: "name", tokenSymbol: "sym",
}

func TestLexerReadsEveryKindOfToken(t *testing.T) {
	got := strings.Join(kindsAndTexts(t, `12 12.50 0.75f 2026-03-31 "a\"b" late_days@1 >= !`), "|")
	want := `int 12|dec 12.50|float 0.75|date 2026-03-31|text a"b|name late_days@1|sym >=|sym !|end`
	if got != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}
}

func TestLexerKeepsTheWrittenScaleOfDecimals(t *testing.T) {
	tree, err := Parse("0.010")
	if err != nil {
		t.Fatal(err)
	}
	if text := tree.(DecimalLiteral).Value.String(); text != "0.010" {
		t.Errorf("decimal is %s", text)
	}
}

func TestLexerTracksLinesAndColumns(t *testing.T) {
	tokens, err := tokenize("a +\n  bb\n\t* c")
	if err != nil {
		t.Fatal(err)
	}
	wants := []Pos{{1, 1}, {1, 3}, {2, 3}, {3, 2}, {3, 4}, {3, 5}}
	for i, want := range wants {
		if tokens[i].pos != want {
			t.Errorf("token %d (%q) is at %+v, want %+v", i, tokens[i].text, tokens[i].pos, want)
		}
	}
}

func TestSubtractionOfNumbersIsNotADate(t *testing.T) {
	got := strings.Join(kindsAndTexts(t, "2026-3-31"), "|")
	if got != "int 2026|sym -|int 3|sym -|int 31|end" {
		t.Errorf("got %s", got)
	}
}

func TestLexerRefusesWhatItCannotRead(t *testing.T) {
	cases := []struct{ source, message string }{
		{"a = b", "The formula cannot be read: the character '=' is not allowed at line 1, column 3."},
		{"a & b", "The formula cannot be read: the character '&' is not allowed at line 1, column 3."},
		{"12abc", "The formula cannot be read: a number cannot be followed by letters at line 1, column 3."},
		{"\"open", "The formula cannot be read: the text is not closed at line 1, column 1."},
		{"a\n  \"open\nmore", "The formula cannot be read: the text is not closed at line 2, column 3."},
		{`"a\q"`, `The formula cannot be read: the escape \q is not known at line 1, column 4.`},
		{"late_days@", "The formula cannot be read: a version number must follow @ at line 1, column 10."},
		{"x $ y", "The formula cannot be read: the character '$' is not allowed at line 1, column 3."},
	}
	for _, c := range cases {
		_, err := tokenize(c.source)
		if err == nil || err.Error() != c.message {
			t.Errorf("%q: got %v, want %s", c.source, err, c.message)
		}
	}
}
