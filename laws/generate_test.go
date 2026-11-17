package laws

import (
	"reflect"
	"strings"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/limit"
)

func TestSameSeedGivesTheSameCaseTwice(t *testing.T) {
	for seed := int64(1); seed <= 200; seed++ {
		first := NewCase(seed)
		second := NewCase(seed)
		if !reflect.DeepEqual(first.Tree, second.Tree) {
			t.Fatalf("seed %d: two trees differ: %q and %q", seed, formula.Format(first.Tree), formula.Format(second.Tree))
		}
		if !reflect.DeepEqual(first.Inputs, second.Inputs) {
			t.Fatalf("seed %d: two input lists differ", seed)
		}
		if first.Text() != second.Text() {
			t.Fatalf("seed %d: two texts differ", seed)
		}
	}
}

func TestEveryGeneratedTreePassesCheckWithinTheProposedLimits(t *testing.T) {
	limits := limit.Proposed()
	for seed := int64(1); seed <= 1000; seed++ {
		c := NewCase(seed)
		if problems := formula.Check(c.Tree, c.Inputs, nil); len(problems) != 0 {
			t.Fatalf("seed %d: %q does not pass Check: %+v", seed, c.Text(), problems)
		}
		if len(c.Inputs) > limits.Inputs {
			t.Fatalf("seed %d: %d inputs, limit %d", seed, len(c.Inputs), limits.Inputs)
		}
		if got := len([]rune(c.Text())); got > limits.FormulaLength {
			t.Fatalf("seed %d: text of %d characters, limit %d", seed, got, limits.FormulaLength)
		}
		if got := formula.Cost(c.Tree, nil); got > limits.Cost {
			t.Fatalf("seed %d: cost %d, limit %d", seed, got, limits.Cost)
		}
	}
}

func TestGeneratedTreesCoverEveryKindOfNode(t *testing.T) {
	seen := map[formula.NodeKind]int{}
	calls := map[string]int{}
	types := map[formula.Type]int{}
	for seed := int64(1); seed <= 1000; seed++ {
		c := NewCase(seed)
		walk(c.Tree.Root, func(n formula.Node) {
			seen[n.Kind()]++
			if call, isCall := n.(formula.Call); isCall {
				calls[call.Name]++
			}
			if literal, isLiteral := n.(formula.Literal); isLiteral {
				types[literal.Type]++
			}
		})
	}
	for _, kind := range []formula.NodeKind{formula.KindLiteral, formula.KindName, formula.KindUnary, formula.KindBinary, formula.KindConditional, formula.KindCall, formula.KindList} {
		if seen[kind] == 0 {
			t.Errorf("no generated tree has a %s node", kind)
		}
	}
	for _, name := range []string{"round", "sum", "count", "avg", "min", "max", "to_float", "to_decimal", "days_between"} {
		if calls[name] == 0 {
			t.Errorf("no generated tree calls %s", name)
		}
	}
	for _, kind := range []formula.Type{formula.TypeInteger, formula.TypeDecimal, formula.TypeText, formula.TypeBoolean} {
		if types[kind] == 0 {
			t.Errorf("no generated tree has a %s literal", kind)
		}
	}
}

func TestSpacingGeneratorKeepsTheStandardTextOfTheTree(t *testing.T) {
	changed := 0
	for seed := int64(1); seed <= 500; seed++ {
		c := NewCase(seed)
		standard := c.Text()
		spelled := SpellDifferently(standard, seed)
		if spelled != standard {
			changed++
		}
		tree, problems := formula.Parse(spelled)
		if len(problems) != 0 {
			t.Fatalf("seed %d: Parse(%q) gave %+v", seed, spelled, problems)
		}
		if formula.Format(tree) != standard {
			t.Fatalf("seed %d: %q formats as %q, want %q", seed, spelled, formula.Format(tree), standard)
		}
		if strings.Join(strings.Fields(spelled), "") != strings.Join(strings.Fields(standard), "") {
			t.Fatalf("seed %d: spacing changed the characters of %q", seed, standard)
		}
	}
	if changed < 400 {
		t.Fatalf("only %d of 500 spellings differ from the standard text", changed)
	}
}

func TestSpacingGeneratorNeverBreaksANumberOrANameOrAText(t *testing.T) {
	for seed := int64(1); seed <= 300; seed++ {
		standard := `max(rate_2, 12.50) + name_x == "a  b, c"`
		spelled := SpellDifferently(standard, seed)
		for _, whole := range []string{"rate_2", "12.50", "name_x", `"a  b, c"`, "max", "=="} {
			if !strings.Contains(spelled, whole) {
				t.Fatalf("seed %d: %q was split in %q", seed, whole, spelled)
			}
		}
	}
}

func TestRowGeneratorMakesOneToTwoHundredRowsWithTheirPlantedProblems(t *testing.T) {
	for seed := int64(1); seed <= 100; seed++ {
		c := NewCase(seed)
		rows := NewRows(seed, c.Inputs)
		if len(rows) < 1 || len(rows) > 200 {
			t.Fatalf("seed %d: %d rows", seed, len(rows))
		}
		again := NewRows(seed, c.Inputs)
		if !reflect.DeepEqual(rows, again) {
			t.Fatalf("seed %d: two row sets differ", seed)
		}
		planted := 0
		for _, row := range rows {
			planted += len(row.Planted)
			if len(row.Planted) > 5 {
				t.Fatalf("seed %d: %d planted problems in one row", seed, len(row.Planted))
			}
			for _, input := range c.Inputs {
				value, given := row.Values[input.Name]
				if !given {
					if !row.Plants(input.Name, "missing") {
						t.Fatalf("seed %d: %s is missing and not planted", seed, input.Name)
					}
					continue
				}
				if value.Type() != input.Type && !row.Plants(input.Name, "wrong-type") {
					t.Fatalf("seed %d: %s has the wrong type and it is not planted", seed, input.Name)
				}
			}
		}
		if len(rows) > 20 && planted == 0 {
			t.Fatalf("seed %d: %d rows and no planted problem", seed, len(rows))
		}
	}
}

func walk(n formula.Node, visit func(formula.Node)) {
	visit(n)
	for _, child := range n.Children() {
		walk(child, visit)
	}
}
