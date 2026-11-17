package functions

import (
	"testing"
)

const referenceDir = "reference"

func loadReference(t *testing.T) map[string][]Case {
	t.Helper()
	versions, err := LoadCases(referenceDir)
	if err != nil {
		t.Fatalf("the recorded cases cannot be read: %v", err)
	}
	return versions
}

func TestEveryRecordedCaseReplaysAgainstTheLibrary(t *testing.T) {
	versions := loadReference(t)
	if len(versions) != 10 {
		t.Fatalf("the reference page has 10 functions and %d data files were read", len(versions))
	}
	for _, key := range SortedKeys(versions) {
		function, found := Find(key)
		if !found {
			t.Errorf("%s is recorded and the library does not have it", key)
			continue
		}
		callable, err := function.Prepare()
		if err != nil {
			t.Errorf("%v", err)
			continue
		}
		for _, c := range versions[key] {
			c := c
			t.Run(c.ID, func(t *testing.T) {
				if failure := Judge(key, c, callable.Call(RowOf(c))); failure != "" {
					t.Error(failure)
				}
			})
		}
	}
}

func TestTheRecordedCasesAreTheEightOfEachFunction(t *testing.T) {
	for key, cases := range loadReference(t) {
		if len(cases) != 8 {
			t.Errorf("%s has %d cases and the page lists 8", key, len(cases))
		}
	}
}
