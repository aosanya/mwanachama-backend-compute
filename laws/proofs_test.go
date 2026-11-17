package laws

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/formula"
)

func firstFailure(law func(Engine, int64) *Failure, engine Engine) *Failure {
	for seed := int64(1); seed <= 500; seed++ {
		if failure := law(engine, seed); failure != nil {
			return failure
		}
	}
	return nil
}

func proveLawFails(t *testing.T, law func(Engine, int64) *Failure, broken Engine, prefix string) *Failure {
	t.Helper()
	failure := firstFailure(law, broken)
	if failure == nil {
		t.Fatal("the broken engine passed all 500 cases")
	}
	if !strings.HasPrefix(failure.Message, prefix) {
		t.Fatalf("message %q does not start with %q", failure.Message, prefix)
	}
	if !strings.Contains(failure.Message, "seed ") {
		t.Fatalf("message %q carries no seed", failure.Message)
	}
	again := law(broken, failure.Seed)
	if again == nil || again.Message != failure.Message {
		t.Fatalf("seed %d did not reproduce the failure: first %q, again %v", failure.Seed, failure.Message, again)
	}
	if healthy := law(RealEngine(), failure.Seed); healthy != nil {
		t.Fatalf("the real engine fails seed %d too: %s", failure.Seed, healthy.Message)
	}
	return failure
}

func TestBrokenFormatThatDropsBracketsFailsTheRoundTripLaw(t *testing.T) {
	broken := RealEngine()
	real := broken.Format
	broken.Format = func(tree formula.Tree) string {
		return strings.NewReplacer("(", "", ")", "").Replace(real(tree))
	}
	failure := proveLawFails(t, RoundTrip, broken, "round trip failed, seed ")
	if !strings.Contains(failure.Message, ", shrunk to ") {
		t.Fatalf("message %q has no shrunk tree", failure.Message)
	}
	if !strings.Contains(failure.Message, "Parse(Format(tree)) ") {
		t.Fatalf("message %q does not say what read back", failure.Message)
	}
}

var trailingZero = regexp.MustCompile(`(\d+\.\d*)0\b`)

func TestBrokenFormatThatKeepsTrimmingFailsTheStandardShapeLaw(t *testing.T) {
	broken := RealEngine()
	real := broken.Format
	broken.Format = func(tree formula.Tree) string {
		return trailingZero.ReplaceAllString(real(tree), "$1")
	}
	proveLawFails(t, StandardShape, broken, "standard shape moved, seed ")
}

func TestBrokenHashOfTheTypedTextFailsTheSpacingLaw(t *testing.T) {
	broken := RealEngine()
	broken.Hash = func(text string) string {
		sum := sha256.Sum256([]byte(text))
		return hex.EncodeToString(sum[:])
	}
	failure := proveLawFails(t, SpacingKeepsHash, broken, "hash differs, seed ")
	if !strings.Contains(failure.Message, "2 spellings, 2 hashes") {
		t.Fatalf("message %q lacks the page's closing words", failure.Message)
	}
}

func TestBrokenRunThatMovesBetweenRunsFailsTheSameOutputLaw(t *testing.T) {
	broken := RealEngine()
	real := broken.Run
	calls := 0
	broken.Run = func(tree formula.Tree, inputs []formula.Input, values map[string]formula.Value) string {
		calls++
		return real(tree, inputs, values) + "#" + string(rune('a'+calls%2))
	}
	proveLawFails(t, SameOutput, broken, "output moved between runs, seed ")
}

func TestBrokenRunThatCalculatesAMissingInputFailsTheValidationFirstLaw(t *testing.T) {
	broken := RealEngine()
	real := broken.Run
	broken.Run = func(tree formula.Tree, inputs []formula.Input, values map[string]formula.Value) string {
		outcome := real(tree, inputs, values)
		if strings.HasPrefix(outcome, "problem:bad-input") {
			return "value:integer:0"
		}
		return outcome
	}
	failure := proveLawFails(t, NoRowWithProblemsIsCalculated, broken, "row ")
	if !strings.Contains(failure.Message, " and was calculated") {
		t.Fatalf("message %q lacks the page's closing words", failure.Message)
	}
}
