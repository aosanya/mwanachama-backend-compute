package examplecases

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const header = "| name | formula | inputs | expected | rule | page |\n|---|---|---|---|---|---|\n"
const pagePath = "documentation/2. design/formula-language.md"

func row(name string, formula string, inputs string, expected string, rule string, page string) string {
	return "| " + strings.Join([]string{name, formula, inputs, expected, rule, page}, " | ") + " |\n"
}

func goodRow(name string) string {
	return row(name, "qty * price", "qty=int:3, price=dec:12.50", "result: dec:37.50", "Multiplying a whole number by a decimal keeps the decimal exact.", pagePath)
}

func parseOne(t *testing.T, text string) ([]Case, error) {
	t.Helper()
	return Parse("formula-language.cases.md", text)
}

func TestParseReadsTheSixColumnsAndTheTypedInputs(t *testing.T) {
	cases, err := parseOne(t, "# Cases\n\n"+header+goodRow("line-total"))
	if err != nil {
		t.Fatal(err)
	}
	if len(cases) != 1 {
		t.Fatalf("got %d cases", len(cases))
	}
	c := cases[0]
	if c.Name != "line-total" || c.Formula != "qty * price" || c.Page != pagePath {
		t.Fatalf("columns: %+v", c)
	}
	if c.File != "formula-language.cases.md" || c.Line != 5 {
		t.Fatalf("where: %s:%d", c.File, c.Line)
	}
	if len(c.Inputs) != 2 || c.Inputs[0].Name != "qty" || c.Inputs[0].Value.Type != "int" || c.Inputs[0].Value.Text != "3" {
		t.Fatalf("inputs: %+v", c.Inputs)
	}
	if c.Inputs[1].Value.Type != "dec" || c.Inputs[1].Value.Text != "12.50" {
		t.Fatalf("decimal kept as written: %+v", c.Inputs[1])
	}
	if c.Expected.Result == nil || c.Expected.Result.Type != "dec" || c.Expected.Result.Text != "37.50" {
		t.Fatalf("expected: %+v", c.Expected)
	}
}

func TestParseKeepsRowsInTheOrderOfTheFile(t *testing.T) {
	cases, err := parseOne(t, header+goodRow("first-case")+goodRow("second-case")+goodRow("third-case"))
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, c := range cases {
		names = append(names, c.Name)
	}
	if strings.Join(names, ",") != "first-case,second-case,third-case" {
		t.Fatalf("order: %v", names)
	}
}

func TestParseReadsEveryValueNotation(t *testing.T) {
	inputs := `a=int:-7, b=dec:0.05, c=float:0.37, d=text:"say \"hi\", \\ bye", e=bool:true, f=date:2026-03-12, g=list<int>:[72, 85], h=list<dec>:[200.00, 250.00], i=list<text>:["72", "104"], j=list<dec>:[], k=list<text>:[]`
	cases, err := parseOne(t, header+row("every-type", "a", inputs, "result: bool:true", "Every type is read.", pagePath))
	if err != nil {
		t.Fatal(err)
	}
	got := cases[0].Inputs
	if len(got) != 11 {
		t.Fatalf("got %d inputs: %+v", len(got), got)
	}
	if got[3].Value.Text != `say "hi", \ bye` {
		t.Fatalf("text escapes: %q", got[3].Value.Text)
	}
	if got[6].Value.Type != "list<int>" || strings.Join(got[6].Value.Items, "|") != "72|85" {
		t.Fatalf("list of integer: %+v", got[6].Value)
	}
	if strings.Join(got[8].Value.Items, "|") != "72|104" {
		t.Fatalf("list of text: %+v", got[8].Value)
	}
	if len(got[9].Value.Items) != 0 || len(got[10].Value.Items) != 0 {
		t.Fatalf("empty lists: %+v %+v", got[9].Value, got[10].Value)
	}
}

func TestParseReadsAnEscapedPipeInAFormula(t *testing.T) {
	cases, err := parseOne(t, header+row("either-one", `a \|\| b`, "a=bool:true, b=bool:false", "result: bool:true", "A pipe in a formula is written with a backslash.", pagePath))
	if err != nil {
		t.Fatal(err)
	}
	if cases[0].Formula != "a || b" {
		t.Fatalf("formula: %q", cases[0].Formula)
	}
}

func TestParseReadsAnEmptyInputsCellAsNoInputs(t *testing.T) {
	cases, err := parseOne(t, header+row("no-inputs", "1 + 2", "  ", "result: int:3", "A formula that reads nothing has no inputs.", pagePath))
	if err != nil {
		t.Fatal(err)
	}
	if len(cases[0].Inputs) != 0 {
		t.Fatalf("inputs: %+v", cases[0].Inputs)
	}
}

func TestParseReadsProblemsInOrderWithRepeats(t *testing.T) {
	cases, err := parseOne(t, header+row("many-problems", "qty", "qty=int:1", "problems: no-negatives, no-negatives, decimal-places", "Problems keep their order.", pagePath))
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Join(cases[0].Expected.Problems, ",")
	if got != "no-negatives,no-negatives,decimal-places" || cases[0].Expected.Result != nil {
		t.Fatalf("expected: %+v", cases[0].Expected)
	}
}

func TestParseRefusesWithTheFileTheLineAndTheReason(t *testing.T) {
	long := strings.Repeat("x", 121)
	refused := []struct {
		label  string
		text   string
		reason string
		line   int
	}{
		{"header with other titles", "| name | formula | inputs | expected | why | page |\n|---|---|---|---|---|---|\n", "bad header", 1},
		{"header missing", "no table here\n", "bad header", 1},
		{"separator missing", "| name | formula | inputs | expected | rule | page |\n" + goodRow("line-total"), "bad header", 2},
		{"five cells", header + "| line-total | a | | result: int:1 | Rule. |\n", "wrong column count", 3},
		{"seven cells", header + "| line-total | a | | result: int:1 | Rule. | " + pagePath + " | extra |\n", "wrong column count", 3},
		{"a line that is not a row", header + goodRow("line-total") + "some prose\n", "wrong column count", 4},
		{"empty name", header + row("", "a", "", "result: int:1", "Rule.", pagePath), "empty cell", 3},
		{"empty formula", header + row("line-total", "", "", "result: int:1", "Rule.", pagePath), "empty cell", 3},
		{"empty expected", header + row("line-total", "a", "", "", "Rule.", pagePath), "empty cell", 3},
		{"empty rule", header + row("line-total", "a", "", "result: int:1", "", pagePath), "empty cell", 3},
		{"empty page", header + row("line-total", "a", "", "result: int:1", "Rule.", ""), "empty cell", 3},
		{"capital in a name", header + row("Line-total", "a", "", "result: int:1", "Rule.", pagePath), "bad name", 3},
		{"name too short", header + row("ab", "a", "", "result: int:1", "Rule.", pagePath), "bad name", 3},
		{"name starts with a digit", header + row("1-line", "a", "", "result: int:1", "Rule.", pagePath), "bad name", 3},
		{"input name with a hyphen", header + row("line-total", "a", "my-qty=int:1", "result: int:1", "Rule.", pagePath), "bad name", 3},
		{"duplicate case name", header + goodRow("line-total") + goodRow("line-total"), "duplicate name", 4},
		{"duplicate input name", header + row("line-total", "a", "qty=int:1, qty=int:2", "result: int:1", "Rule.", pagePath), "duplicate name", 3},
		{"input without a type", header + row("line-total", "a", "qty=3", "result: int:1", "Rule.", pagePath), "bad input", 3},
		{"unknown type", header + row("line-total", "a", "qty=money:3", "result: int:1", "Rule.", pagePath), "bad input", 3},
		{"integer with a dot", header + row("line-total", "a", "qty=int:3.5", "result: int:1", "Rule.", pagePath), "bad input", 3},
		{"decimal with an exponent", header + row("line-total", "a", "qty=dec:1e3", "result: int:1", "Rule.", pagePath), "bad input", 3},
		{"bool with a capital", header + row("line-total", "a", "ok=bool:True", "result: int:1", "Rule.", pagePath), "bad input", 3},
		{"date that is not a day", header + row("line-total", "a", "on=date:2026-02-30", "result: int:1", "Rule.", pagePath), "bad input", 3},
		{"text without quotes", header + row("line-total", "a", "word=text:red", "result: int:1", "Rule.", pagePath), "bad input", 3},
		{"list without brackets", header + row("line-total", "a", "xs=list<int>:1, 2", "result: int:1", "Rule.", pagePath), "bad input", 3},
		{"list item of the wrong type", header + row("line-total", "a", "xs=list<int>:[1, 2.5]", "result: int:1", "Rule.", pagePath), "bad input", 3},
		{"neither result nor problems", header + row("line-total", "a", "", "dec:1", "Rule.", pagePath), "bad expected", 3},
		{"result with a bad value", header + row("line-total", "a", "", "result: dec:abc", "Rule.", pagePath), "bad expected", 3},
		{"problem code with a capital", header + row("line-total", "a", "", "problems: Bad-Code", "Rule.", pagePath), "bad expected", 3},
		{"problems with nothing after", header + row("line-total", "a", "", "problems: ", "Rule.", pagePath), "bad expected", 3},
		{"result and problems together", header + row("line-total", "a", "", "result: int:1, problems: a-b", "Rule.", pagePath), "bad expected", 3},
		{"another page", header + row("line-total", "a", "", "result: int:1", "Rule.", "documentation/2. design/batch.md"), "wrong page", 3},
		{"rule too long", header + row("line-total", "a", "", "result: int:1", long, pagePath), "too long", 3},
		{"formula too long", header + row("line-total", strings.Repeat("a", 501), "", "result: int:1", "Rule.", pagePath), "too long", 3},
		{"integer with 19 digits", header + row("line-total", "a", "qty=int:1234567890123456789", "result: int:1", "Rule.", pagePath), "too long", 3},
		{"text with 201 characters", header + row("line-total", "a", `word=text:"`+strings.Repeat("x", 201)+`"`, "result: int:1", "Rule.", pagePath), "too long", 3},
	}
	for _, tc := range refused {
		t.Run(tc.label, func(t *testing.T) {
			_, err := parseOne(t, tc.text)
			if err == nil {
				t.Fatal("no refusal")
			}
			want := "formula-language.cases.md:" + strconv.Itoa(tc.line) + ": " + tc.reason
			if !strings.HasPrefix(err.Error(), want) {
				t.Fatalf("got %q, want prefix %q", err.Error(), want)
			}
		})
	}
}

func TestParseRefusesMoreThanTwoHundredRows(t *testing.T) {
	var b strings.Builder
	b.WriteString(header)
	for i := 0; i < 201; i++ {
		b.WriteString(goodRow("case-" + strings.Repeat("a", i%20+1) + "-" + strconv.Itoa(i)))
	}
	_, err := parseOne(t, b.String())
	if err == nil || !strings.Contains(err.Error(), "too many rows") {
		t.Fatalf("got %v", err)
	}
}

func TestLoadReadsEveryFileOfAFolderInNameOrderAndNeverSkipsOne(t *testing.T) {
	dir := t.TempDir()
	write := func(name string, text string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("formula-language.cases.md", header+goodRow("from-language"))
	write("batch.cases.md", header+strings.ReplaceAll(goodRow("from-batch"), "formula-language", "batch"))
	cases, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(cases) != 2 || cases[0].Name != "from-batch" || cases[1].Name != "from-language" {
		t.Fatalf("cases: %+v", cases)
	}
	write("posting.cases.md", "not a table\n")
	if _, err := Load(dir); err == nil || !strings.HasPrefix(err.Error(), "posting.cases.md:1: bad header") {
		t.Fatalf("got %v", err)
	}
}

func TestLoadRefusesTheSameNameInTwoFiles(t *testing.T) {
	dir := t.TempDir()
	one := header + goodRow("shared-name")
	two := header + strings.ReplaceAll(goodRow("shared-name"), "formula-language", "batch")
	if err := os.WriteFile(filepath.Join(dir, "formula-language.cases.md"), []byte(one), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "batch.cases.md"), []byte(two), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(dir); err == nil || !strings.Contains(err.Error(), "duplicate name") {
		t.Fatalf("got %v", err)
	}
}
