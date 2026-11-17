package formula

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
)

type generator struct {
	rng *rand.Rand
}

var generatedNames = []string{"a", "b", "balance", "payment", "rate_2", "x"}
var generatedFunctions = []string{"round", "sum", "max", "late_days", "pmt"}
var generatedLibraries = []string{"fin", "tax"}
var generatedBinary = []string{"+", "-", "*", "/", "%", "<", "<=", ">", ">=", "==", "!=", "&&", "||"}

func (g generator) digits(n int) string {
	out := make([]byte, n)
	for i := range out {
		out[i] = byte('0' + g.rng.Intn(10))
	}
	return string(out)
}

func (g generator) literal() Node {
	switch g.rng.Intn(5) {
	case 0:
		if g.rng.Intn(4) == 0 {
			return Literal{Type: TypeInteger, Value: "0"}
		}
		return Literal{Type: TypeInteger, Value: string(rune('1'+g.rng.Intn(9))) + g.digits(g.rng.Intn(5))}
	case 1:
		whole := "0"
		if g.rng.Intn(2) == 0 {
			whole = string(rune('1'+g.rng.Intn(9))) + g.digits(g.rng.Intn(4))
		}
		return Literal{Type: TypeDecimal, Value: whole + "." + g.digits(1+g.rng.Intn(4))}
	case 2:
		pool := []string{"", "active", "say \"hi\"", "back\\slash", "two words", "a, b", "(x)"}
		return Literal{Type: TypeText, Value: pool[g.rng.Intn(len(pool))]}
	case 3:
		return Literal{Type: TypeBoolean, Value: "true"}
	default:
		return Literal{Type: TypeBoolean, Value: "false"}
	}
}

func (g generator) node(depth int) Node {
	if depth <= 1 {
		if g.rng.Intn(2) == 0 {
			return g.literal()
		}
		return Name{Name: generatedNames[g.rng.Intn(len(generatedNames))]}
	}
	switch g.rng.Intn(8) {
	case 0:
		return g.literal()
	case 1:
		return Name{Name: generatedNames[g.rng.Intn(len(generatedNames))]}
	case 2:
		return Unary{Op: "!", Operand: g.node(depth - 1)}
	case 3, 4:
		return Binary{Op: generatedBinary[g.rng.Intn(len(generatedBinary))], Left: g.node(depth - 1), Right: g.node(depth - 1)}
	case 5:
		return Conditional{Condition: g.node(depth - 1), Then: g.node(depth - 1), Else: g.node(depth - 1)}
	case 6:
		call := Call{Name: generatedFunctions[g.rng.Intn(len(generatedFunctions))]}
		if g.rng.Intn(2) == 0 {
			version := 1 + g.rng.Intn(999999)
			call.Version = &version
			if g.rng.Intn(2) == 0 {
				library := generatedLibraries[g.rng.Intn(len(generatedLibraries))]
				call.Library = &library
			}
		}
		for i := 1 + g.rng.Intn(3); i > 0; i-- {
			call.Args = append(call.Args, g.node(depth-1))
		}
		return call
	default:
		list := List{}
		for i := 1 + g.rng.Intn(3); i > 0; i-- {
			list.Items = append(list.Items, g.node(depth-1))
		}
		return list
	}
}

func TestLawRoundTripOnGeneratedTrees(t *testing.T) {
	for seed := int64(1); seed <= 500; seed++ {
		g := generator{rng: rand.New(rand.NewSource(seed))}
		tree := Tree{Root: g.node(2 + g.rng.Intn(4))}
		text := Format(tree)
		again, problems := Parse(text)
		if len(problems) != 0 {
			t.Fatalf("seed %d: Parse(%q) gave %+v", seed, text, problems)
		}
		if !reflect.DeepEqual(tree, again) {
			t.Fatalf("seed %d: round trip failed on %q, read back as %q", seed, text, Format(again))
		}
	}
}

func TestLawStandardShapeIsStable(t *testing.T) {
	for seed := int64(1); seed <= 500; seed++ {
		g := generator{rng: rand.New(rand.NewSource(seed))}
		tree := Tree{Root: g.node(2 + g.rng.Intn(4))}
		first := Format(tree)
		second := Format(mustParse(t, first))
		if first != second {
			t.Fatalf("seed %d: first Format gave %q, second gave %q", seed, first, second)
		}
	}
}

func TestLawSpacingDoesNotChangeTheStandardText(t *testing.T) {
	for seed := int64(1); seed <= 300; seed++ {
		g := generator{rng: rand.New(rand.NewSource(seed))}
		standard := Format(Tree{Root: g.node(2 + g.rng.Intn(4))})
		typed := respaced(standard, g.rng)
		tree, problems := Parse(typed)
		if len(problems) != 0 {
			t.Fatalf("seed %d: Parse(%q) gave %+v", seed, typed, problems)
		}
		if got := Format(tree); got != standard {
			t.Fatalf("seed %d: %q gave %q, want %q", seed, typed, got, standard)
		}
	}
}

func TestLawCheckGivesTheSameAnswerEveryTime(t *testing.T) {
	library := testLibrary()
	for seed := int64(1); seed <= 100; seed++ {
		g := generator{rng: rand.New(rand.NewSource(seed))}
		tree := Tree{Root: g.node(2 + g.rng.Intn(4))}
		in := inputs("a", "integer", "b", "decimal", "balance", "decimal", "payment", "decimal", "rate_2", "float", "x", "boolean")
		first := Check(tree, in, library)
		for i := 0; i < 5; i++ {
			if again := Check(tree, in, library); !reflect.DeepEqual(first, again) {
				t.Fatalf("seed %d: Check moved between runs: %s", seed, fmt.Sprint(first))
			}
		}
	}
}
