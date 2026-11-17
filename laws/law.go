package laws

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"

	"github.com/aosanya/mwanachama-backend-compute/formula"
)

const (
	spacingSeedOffset = 1 << 20
	rowsPerCase       = 20
)

type Failure struct {
	Law     string
	Seed    int64
	Shrunk  string
	Message string
}

func (f *Failure) Error() string {
	return f.Message
}

func validIn(engine Engine, inputs []formula.Input) func(formula.Tree) bool {
	return func(tree formula.Tree) bool { return len(engine.Check(tree, inputs)) == 0 }
}

func RoundTrip(engine Engine, seed int64) *Failure {
	c := NewCase(seed)
	readBack := func(tree formula.Tree) (string, bool) {
		again, problems := engine.Parse(engine.Format(tree))
		if len(problems) > 0 {
			return "Parse(Format(tree)) was refused: " + problems[0].Sentence, true
		}
		if !reflect.DeepEqual(tree, again) {
			return "Parse(Format(tree)) read back as " + Show(again), true
		}
		return "", false
	}
	fails := func(tree formula.Tree) bool {
		_, failed := readBack(tree)
		return failed
	}
	if !fails(c.Tree) {
		return nil
	}
	shrunk := Shrink(c.Tree, validIn(engine, c.Inputs), fails)
	detail, _ := readBack(shrunk)
	return failure("round trip", seed, shrunk, fmt.Sprintf("round trip failed, seed %d, shrunk to %s: %s", seed, Show(shrunk), detail))
}

func StandardShape(engine Engine, seed int64) *Failure {
	c := NewCase(seed)
	twice := func(tree formula.Tree) (string, string, bool) {
		first := engine.Format(tree)
		again, problems := engine.Parse(first)
		if len(problems) > 0 {
			return first, "a refusal: " + problems[0].Sentence, true
		}
		second := engine.Format(again)
		return first, second, first != second
	}
	fails := func(tree formula.Tree) bool {
		_, _, failed := twice(tree)
		return failed
	}
	if !fails(c.Tree) {
		return nil
	}
	shrunk := Shrink(c.Tree, validIn(engine, c.Inputs), fails)
	first, second, _ := twice(shrunk)
	return failure("standard shape", seed, shrunk, fmt.Sprintf("standard shape moved, seed %d, shrunk to %s: first Format gave %s, second gave %s", seed, Show(shrunk), first, second))
}

func SpacingKeepsHash(engine Engine, seed int64) *Failure {
	c := NewCase(seed)
	spellings := func(tree formula.Tree) (string, string, bool) {
		standard := formula.Format(tree)
		one := SpellDifferently(standard, seed)
		two := SpellDifferently(standard, seed+spacingSeedOffset)
		return one, two, engine.Hash(one) != engine.Hash(two)
	}
	fails := func(tree formula.Tree) bool {
		_, _, failed := spellings(tree)
		return failed
	}
	if !fails(c.Tree) {
		return nil
	}
	shrunk := Shrink(c.Tree, validIn(engine, c.Inputs), fails)
	one, two, _ := spellings(shrunk)
	return failure("spacing and hash", seed, shrunk, fmt.Sprintf("hash differs, seed %d, shrunk to %s and %s: 2 spellings, 2 hashes", seed, visible(one), visible(two)))
}

func visible(text string) string {
	return strings.NewReplacer("\n", `\n`, "\t", `\t`, "\r", `\r`).Replace(text)
}

func runRows(engine Engine, c Case, tree formula.Tree) []string {
	rows := NewRows(c.Seed, c.Inputs)
	if len(rows) > rowsPerCase {
		rows = rows[:rowsPerCase]
	}
	outcomes := make([]string, len(rows))
	for i, row := range rows {
		outcomes[i] = engine.Run(tree, c.Inputs, row.Values)
	}
	return outcomes
}

func SameOutput(engine Engine, seed int64) *Failure {
	c := NewCase(seed)
	moved := func(tree formula.Tree) (string, string, bool) {
		first := runRows(engine, c, tree)
		second := runRows(engine, c, tree)
		for i := range first {
			if first[i] != second[i] {
				return first[i], second[i], true
			}
		}
		return "", "", false
	}
	fails := func(tree formula.Tree) bool {
		_, _, failed := moved(tree)
		return failed
	}
	if !fails(c.Tree) {
		return nil
	}
	shrunk := Shrink(c.Tree, validIn(engine, c.Inputs), fails)
	one, two, _ := moved(shrunk)
	return failure("same output", seed, shrunk, fmt.Sprintf("output moved between runs, seed %d, shrunk to %s: run 1 gave %s, run 2 gave %s", seed, Show(shrunk), one, two))
}

func plural(count int) string {
	if count == 1 {
		return "1 problem"
	}
	return fmt.Sprintf("%d problems", count)
}

func NoRowWithProblemsIsCalculated(engine Engine, seed int64) *Failure {
	c := NewCase(seed)
	rows := NewRows(seed, c.Inputs)
	if len(rows) > rowsPerCase {
		rows = rows[:rowsPerCase]
	}
	calculated := func(tree formula.Tree) (int, bool) {
		for i, row := range rows {
			if len(row.Planted) > 0 && strings.HasPrefix(engine.Run(tree, c.Inputs, row.Values), "value:") {
				return i, true
			}
		}
		return 0, false
	}
	index, failed := calculated(c.Tree)
	if !failed {
		return nil
	}
	fails := func(tree formula.Tree) bool {
		_, failed := calculated(tree)
		return failed
	}
	shrunk := Shrink(c.Tree, validIn(engine, c.Inputs), fails)
	index, _ = calculated(shrunk)
	return failure("validation first", seed, shrunk, fmt.Sprintf("row %d has %s and was calculated, seed %d, shrunk to %s", index+1, plural(len(rows[index].Planted)), seed, Show(shrunk)))
}

func failure(law string, seed int64, shrunk formula.Tree, message string) *Failure {
	return &Failure{Law: law, Seed: seed, Shrunk: Show(shrunk), Message: message}
}

func OutputDigest(engine Engine, base int64, count int) string {
	hash := sha256.New()
	for i := int64(1); i <= int64(count); i++ {
		c := NewCase(base + i)
		fmt.Fprintf(hash, "%d %s\n", c.Seed, c.Text())
		for _, outcome := range runRows(engine, c, c.Tree) {
			fmt.Fprintf(hash, "%s\n", outcome)
		}
	}
	return hex.EncodeToString(hash.Sum(nil))
}
