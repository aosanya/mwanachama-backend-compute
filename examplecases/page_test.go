package examplecases

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type pageFixture struct {
	root string
}

func newPageFixture(t *testing.T) pageFixture {
	t.Helper()
	root := t.TempDir()
	for _, folder := range []string{"documentation/2. design", "testdata"} {
		if err := os.MkdirAll(filepath.Join(root, folder), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return pageFixture{root: root}
}

func (f pageFixture) write(t *testing.T, path string, text string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(f.root, path), []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

func example(formula string, name string) string {
	return "```formula\n" + formula + "\n```\ncase: " + name + "\n\n"
}

func (f pageFixture) failures(t *testing.T) []string {
	t.Helper()
	cases, err := Load(filepath.Join(f.root, "testdata"))
	if err != nil {
		t.Fatal(err)
	}
	return CheckPages(f.root, "testdata", cases)
}

func TestCheckPagesPassesWhenEveryExampleHasARowAndEveryRowAnExample(t *testing.T) {
	f := newPageFixture(t)
	f.write(t, pagePath, "# Language\n\n"+example("qty * price", "line-total")+example("a + b", "sum-two"))
	f.write(t, "testdata/formula-language.cases.md", header+goodRow("line-total")+goodRow("sum-two"))
	if got := f.failures(t); len(got) != 0 {
		t.Fatalf("failures: %v", got)
	}
}

func TestCheckPagesReportsAnExampleWithNoRow(t *testing.T) {
	f := newPageFixture(t)
	f.write(t, pagePath, "# Language\n\n"+example("qty * price", "line-total")+example("a + b", "sum-two"))
	f.write(t, "testdata/formula-language.cases.md", header+goodRow("line-total"))
	got := f.failures(t)
	want := pagePath + `:8: example "sum-two" has no row in testdata/formula-language.cases.md`
	if len(got) != 1 || got[0] != want {
		t.Fatalf("got %v, want %q", got, want)
	}
}

func TestCheckPagesReportsAnExampleWithNoCaseLine(t *testing.T) {
	f := newPageFixture(t)
	f.write(t, pagePath, "# Language\n\n```formula\nqty * price\n```\n\nprose\n")
	f.write(t, "testdata/formula-language.cases.md", header+goodRow("line-total"))
	got := f.failures(t)
	if len(got) < 1 || got[0] != pagePath+":3: example has no case line" {
		t.Fatalf("got %v", got)
	}
}

func TestCheckPagesReportsARowWhoseExampleIsGone(t *testing.T) {
	f := newPageFixture(t)
	f.write(t, pagePath, "# Language\n\n"+example("qty * price", "line-total"))
	f.write(t, "testdata/formula-language.cases.md", header+goodRow("line-total")+goodRow("gone-case"))
	got := f.failures(t)
	want := `testdata/formula-language.cases.md:4: row "gone-case" names page "` + pagePath + `" which has no example marked "case: gone-case"`
	if len(got) != 1 || got[0] != want {
		t.Fatalf("got %v, want %q", got, want)
	}
}

func TestCheckPagesReportsARowWhoseDesignPageDoesNotExist(t *testing.T) {
	f := newPageFixture(t)
	f.write(t, "testdata/formula-language.cases.md", header+goodRow("line-total"))
	got := f.failures(t)
	want := `testdata/formula-language.cases.md:3: row "line-total" names page "` + pagePath + `" which does not exist`
	if len(got) != 1 || got[0] != want {
		t.Fatalf("got %v, want %q", got, want)
	}
}

func TestCheckPagesPassesWhenNeitherThePageNorItsCaseFileExists(t *testing.T) {
	f := newPageFixture(t)
	if got := f.failures(t); len(got) != 0 {
		t.Fatalf("failures: %v", got)
	}
}

func TestCheckPagesReportsEveryExampleWhenThePageExistsAndItsCaseFileDoesNot(t *testing.T) {
	f := newPageFixture(t)
	f.write(t, pagePath, "# Language\n\n"+example("qty * price", "line-total")+example("a + b", "sum-two"))
	got := f.failures(t)
	if len(got) != 2 || !strings.Contains(got[0], `"line-total" has no row`) || !strings.Contains(got[1], `"sum-two" has no row`) {
		t.Fatalf("got %v", got)
	}
}

func TestCheckPagesIgnoresFencesThatAreNotFormulaBlocks(t *testing.T) {
	f := newPageFixture(t)
	f.write(t, pagePath, "# Language\n\n```go\nx := 1\n```\n\n"+example("qty * price", "line-total"))
	f.write(t, "testdata/formula-language.cases.md", header+goodRow("line-total"))
	if got := f.failures(t); len(got) != 0 {
		t.Fatalf("failures: %v", got)
	}
}

func TestScanPageRefusesAMarkerUsedTwice(t *testing.T) {
	f := newPageFixture(t)
	f.write(t, pagePath, "# Language\n\n"+example("qty * price", "line-total")+example("a + b", "line-total"))
	f.write(t, "testdata/formula-language.cases.md", header+goodRow("line-total"))
	got := f.failures(t)
	if len(got) == 0 || !strings.Contains(got[0], `marker "line-total" is used twice`) {
		t.Fatalf("got %v", got)
	}
}
