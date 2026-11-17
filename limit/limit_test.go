package limit

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

type limitRow struct {
	name  string
	set   func(*Measure, int)
	limit Name
	value int
}

func limitRows() []limitRow {
	return []limitRow{
		{"formula length 500 proposed", func(m *Measure, v int) { m.FormulaLength = v }, FormulaLengthLimit, 500},
		{"tree depth 12 proposed", func(m *Measure, v int) { m.Depth = v }, DepthLimit, 12},
		{"brackets open 12 proposed", func(m *Measure, v int) { m.BracketsOpen = v }, DepthLimit, 12},
		{"nodes 100 proposed", func(m *Measure, v int) { m.Nodes = v }, NodeLimit, 100},
		{"inputs 8 proposed", func(m *Measure, v int) { m.Inputs = v }, InputLimit, 8},
		{"text literal 200 proposed", func(m *Measure, v int) { m.TextLiteral = v }, TextLengthLimit, 200},
		{"integer literal 18 digits proposed", func(m *Measure, v int) { m.IntegerDigits = v }, NumberMagnitudeLimit, 18},
		{"decimal literal 30 digits proposed", func(m *Measure, v int) { m.DecimalDigits = v }, NumberMagnitudeLimit, 30},
		{"places of a decimal literal 10 proposed", func(m *Measure, v int) { m.DecimalPlaces = v }, DecimalPlacesLimit, 10},
		{"name length 40 proposed", func(m *Measure, v int) { m.NameLength = v }, NameLengthLimit, 40},
		{"call depth 4 proposed", func(m *Measure, v int) { m.CallDepth = v }, CallDepthLimit, 4},
		{"cost 20000 proposed", func(m *Measure, v int) { m.Cost = v }, CostLimit, 20000},
		{"list value 1000 items proposed", func(m *Measure, v int) { m.ListItems = v }, ListLengthLimit, 1000},
	}
}

func TestEveryLimitAcceptsTheValueAndRefusesOneOver(t *testing.T) {
	settings := Proposed()
	for _, row := range limitRows() {
		t.Run(row.name, func(t *testing.T) {
			var atLimit Measure
			row.set(&atLimit, row.value)
			if refusal, refused := settings.Refusal(atLimit); refused {
				t.Fatalf("at the limit %d was refused: %s", row.value, refusal.Sentence())
			}
			var overLimit Measure
			row.set(&overLimit, row.value+1)
			refusal, refused := settings.Refusal(overLimit)
			if !refused {
				t.Fatalf("one over the limit %d was accepted", row.value)
			}
			if refusal.Limit != row.limit || refusal.Value != row.value {
				t.Fatalf("refusal named %q %d, want %q %d", refusal.Limit, refusal.Value, row.limit, row.value)
			}
			want := fmt.Sprintf("This formula goes over the %s limit of %d.", row.limit, row.value)
			if refusal.Sentence() != want {
				t.Fatalf("sentence %q, want %q", refusal.Sentence(), want)
			}
		})
	}
}

func TestAnEmptyMeasureIsNeverRefused(t *testing.T) {
	if refusal, refused := Proposed().Refusal(Measure{}); refused {
		t.Fatalf("empty measure refused: %s", refusal.Sentence())
	}
}

func TestTheProposedValuesAreTheLanguagePageTable(t *testing.T) {
	got := Proposed()
	want := Settings{
		FormulaLength: 500,
		Depth:         12,
		BracketsOpen:  12,
		Nodes:         100,
		Inputs:        8,
		CallArguments: 8,
		TextLiteral:   200,
		IntegerDigits: 18,
		DecimalDigits: 30,
		DecimalPlaces: 10,
		NameLength:    40,
		CallDepth:     4,
		Cost:          20000,
		ListItems:     1000,
	}
	if got != want {
		t.Fatalf("proposed settings %+v, want %+v", got, want)
	}
}

func TestTheFirstLimitInTableOrderIsNamed(t *testing.T) {
	refusal, refused := Proposed().Refusal(Measure{FormulaLength: 501, Depth: 13, Cost: 20001})
	if !refused || refusal.Limit != FormulaLengthLimit {
		t.Fatalf("got %+v refused %v, want the formula length limit first", refusal, refused)
	}
	refusal, refused = Proposed().Refusal(Measure{Depth: 13, Cost: 20001})
	if !refused || refusal.Limit != DepthLimit {
		t.Fatalf("got %+v refused %v, want the depth limit first", refusal, refused)
	}
}

func TestLimitsAreSettingsNotConstants(t *testing.T) {
	settings := Proposed()
	settings.Depth = 3
	settings.Cost = 50
	if _, refused := settings.Refusal(Measure{Depth: 3}); refused {
		t.Fatal("depth 3 refused under a depth setting of 3")
	}
	refusal, refused := settings.Refusal(Measure{Depth: 4})
	if !refused || refusal.Sentence() != "This formula goes over the depth limit of 3." {
		t.Fatalf("got %+v refused %v", refusal, refused)
	}
	refusal, refused = settings.Refusal(Measure{Cost: 51})
	if !refused || refusal.Sentence() != "This formula goes over the cost limit of 50." {
		t.Fatalf("got %+v refused %v", refusal, refused)
	}
}

func TestArgumentsInOneCallAreCheckedAgainstTheSetting(t *testing.T) {
	settings := Proposed()
	if !settings.AllowsArguments(8) {
		t.Fatal("8 arguments refused")
	}
	if settings.AllowsArguments(9) {
		t.Fatal("9 arguments accepted")
	}
}

func TestARefusalIsAnErrorWithTheSameSentence(t *testing.T) {
	refusal := Refusal{Limit: CostLimit, Value: 20000}
	var err error = refusal
	if err.Error() != refusal.Sentence() {
		t.Fatalf("error %q, sentence %q", err.Error(), refusal.Sentence())
	}
	if !strings.HasSuffix(err.Error(), "20000.") {
		t.Fatalf("sentence %q does not end with the value", err.Error())
	}
}

func TestNothingInThePackageReadsAClock(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		source, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), file, source, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, spec := range parsed.Imports {
			path, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				t.Fatal(err)
			}
			if path == "time" || path == "math/rand" || path == "context" {
				t.Fatalf("%s imports %s", file, path)
			}
		}
	}
}
