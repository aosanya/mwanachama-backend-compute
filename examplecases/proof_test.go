package examplecases

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func readFile(t *testing.T, path string) string {
	t.Helper()
	text, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(text)
}

func copyOfTheCaseFile(t *testing.T, change func(string) string) (string, []Case) {
	t.Helper()
	dir := t.TempDir()
	text := change(readFile(t, caseDir+"/formula-language.cases.md"))
	if err := os.WriteFile(filepath.Join(dir, "formula-language.cases.md"), []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	cases, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	return dir, cases
}

func TestChangingOneExpectedResultInACopyMakesExactlyThatCaseFail(t *testing.T) {
	held, err := LoadHeld(heldPath)
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := NewStandIn()
	if err != nil {
		t.Fatal(err)
	}
	_, untouched := copyOfTheCaseFile(t, func(text string) string { return text })
	if failing := Failing(untouched, held, runtime); len(failing) != 0 {
		t.Fatalf("the unchanged copy fails: %v", failing)
	}
	changes := map[string][2]string{
		"lang-f01-balance-after-payment":     {"result: dec:75.25 | A conditional gives", "result: dec:75.26 | A conditional gives"},
		"lang-f19-seed-bags-left-over":       {"result: int:37 | The remainder operator", "result: int:38 | The remainder operator"},
		"lang-problem-divide-by-zero":        {"problems: divide-by-zero | Dividing by zero is a problem and never", "problems: unknown-name | Dividing by zero is a problem and never"},
		"lang-decimal-equal-by-value":        {"result: bool:true | Comparison reads the value", "result: bool:false | Comparison reads the value"},
		"lang-f13-collateral-coverage-ratio": {"float:1.363", "float:1.364"},
	}
	for name, change := range changes {
		t.Run(name, func(t *testing.T) {
			_, cases := copyOfTheCaseFile(t, func(text string) string {
				if strings.Count(text, change[0]) != 1 {
					t.Fatalf("%q is not found once in the case file", change[0])
				}
				return strings.Replace(text, change[0], change[1], 1)
			})
			failing := Failing(cases, held, runtime)
			if len(failing) != 1 || failing[0] != name {
				t.Fatalf("failing cases: %v, want only %s", failing, name)
			}
		})
	}
}

func copyOfTheDesignPage(t *testing.T, change func(string) string) string {
	t.Helper()
	root := t.TempDir()
	folder := filepath.Join(root, "documentation", "2. design")
	if err := os.MkdirAll(folder, 0o755); err != nil {
		t.Fatal(err)
	}
	text := change(readFile(t, repoRoot+"/"+pagePath))
	if err := os.WriteFile(filepath.Join(folder, "formula-language.md"), []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestAFormulaAddedToACopyOfTheDesignPageIsReportedNotSkipped(t *testing.T) {
	cases := loadAll(t)
	unmarked := "\n```formula\ntotal * 2\n```\n\nprose without a case line\n"
	root := copyOfTheDesignPage(t, func(text string) string { return text + unmarked })
	failures := CheckPages(root, caseDirName, cases)
	if len(failures) != 1 || !strings.HasSuffix(failures[0], ": example has no case line") {
		t.Fatalf("failures: %v", failures)
	}
	marked := "\n```formula\ntotal * 2\n```\ncase: lang-added-example\n"
	root = copyOfTheDesignPage(t, func(text string) string { return text + marked })
	failures = CheckPages(root, caseDirName, cases)
	if len(failures) != 1 || !strings.Contains(failures[0], `example "lang-added-example" has no row in testdata/formula-language.cases.md`) {
		t.Fatalf("failures: %v", failures)
	}
}

func TestAnExampleRemovedFromACopyOfTheDesignPageLeavesARowToReport(t *testing.T) {
	cases := loadAll(t)
	root := copyOfTheDesignPage(t, func(text string) string {
		block := "```formula\nkilos % bag_size\n```\ncase: lang-f19-seed-bags-left-over\n"
		if strings.Count(text, block) != 1 {
			t.Fatal("the example of formula 19 is not found once on the page")
		}
		return strings.Replace(text, block, "", 1)
	})
	failures := CheckPages(root, caseDirName, cases)
	if len(failures) != 1 || !strings.Contains(failures[0], `row "lang-f19-seed-bags-left-over" names page`) || !strings.Contains(failures[0], `has no example marked "case: lang-f19-seed-bags-left-over"`) {
		t.Fatalf("failures: %v", failures)
	}
}
