package own

import (
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
)

func compileWithOptions(options Options, source string) error {
	c := harness.Case{Source: source, Result: integerKind, Params: []harness.Param{{Name: "n", Kind: integerKind}}}
	_, err := NewWith(options).Compile(c, nil)
	return err
}

func TestZeroOptionsMeanUnlimited(t *testing.T) {
	if err := compileWithOptions(Options{}, "1 + 2 * 3 + 4 + 5 + 6 + 7"); err != nil {
		t.Error(err)
	}
}

func TestMaxDepthRefusesDeeperFormulas(t *testing.T) {
	if err := compileWithOptions(Options{MaxDepth: 4}, "1 + 2 * 3 + 4"); err != nil {
		t.Errorf("depth 4 should fit: %v", err)
	}
	err := compileWithOptions(Options{MaxDepth: 3}, "1 + 2 * 3 + 4")
	if err == nil || err.Error() != "The formula is nested too deeply: 4 levels where 3 are allowed." {
		t.Errorf("got %v", err)
	}
}

func TestMaxNodesRefusesLargerFormulas(t *testing.T) {
	if err := compileWithOptions(Options{MaxNodes: 7}, "1 + 2 * 3 + 4"); err != nil {
		t.Errorf("7 nodes should fit: %v", err)
	}
	err := compileWithOptions(Options{MaxNodes: 6}, "1 + 2 * 3 + 4")
	if err == nil || err.Error() != "The formula is too large: 7 parts where 6 are allowed." {
		t.Errorf("got %v", err)
	}
}

func TestLimitsAreCheckedAfterReadingSoSyntaxErrorsComeFirst(t *testing.T) {
	err := compileWithOptions(Options{MaxDepth: 1}, "1 +")
	if err == nil || err.Error() != "The formula cannot be read: expected a value but found the end of the formula at line 1, column 4." {
		t.Errorf("got %v", err)
	}
}

func TestEngineIsNamedOwn(t *testing.T) {
	if New().Name() != "own" || NewWith(Options{MaxDepth: 2}).Name() != "own" {
		t.Error("name is not own")
	}
}
