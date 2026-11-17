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
	definitions := loadDefinitions(t)
	_, untouched := copyOfTheCaseFile(t, func(text string) string { return text })
	if failing := Failing(untouched, held, definitions, runtime); len(failing) != 0 {
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
			failing := Failing(cases, held, definitions, runtime)
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

func TestChangingOneExpectedResultOfAValidateRowInACopyMakesExactlyThatCaseFail(t *testing.T) {
	held, err := LoadHeld(heldPath)
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := NewStandIn()
	if err != nil {
		t.Fatal(err)
	}
	definitions := loadDefinitions(t)
	changes := map[string][2]string{
		"rules-required-valid":            {"result: dec:75.25 | A clean row is calculated", "result: dec:75.26 | A clean row is calculated"},
		"rules-three-problems-at-once":    {"problems: no-negatives, no-negatives, decimal-places |", "problems: no-negatives, decimal-places, no-negatives |"},
		"rules-check-failed":              {"problems: check-failed | A check that runs and is false", "problems: check-error | A check that runs and is false"},
		"rules-definition-range-reversed": {"problems: range-reversed |", "problems: places-too-many |"},
	}
	for name, change := range changes {
		t.Run(name, func(t *testing.T) {
			_, cases := copyOfTheCaseFile(t, func(text string) string {
				if strings.Count(text, change[0]) != 1 {
					t.Fatalf("%q is not found once in the case file", change[0])
				}
				return strings.Replace(text, change[0], change[1], 1)
			})
			failing := Failing(cases, held, definitions, runtime)
			if len(failing) != 1 || failing[0] != name {
				t.Fatalf("failing cases: %v, want only %s", failing, name)
			}
		})
	}
}

func TestChangingARuleOfADefinitionInACopyMakesTheCasesThatUseItFail(t *testing.T) {
	held, err := LoadHeld(heldPath)
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := NewStandIn()
	if err != nil {
		t.Fatal(err)
	}
	original := "| V1 | input | payment | decimal: required; no-negatives; decimal-places 2 | |"
	changed := "| V1 | input | payment | decimal: required; no-negatives; decimal-places 3 | |"
	text := readFile(t, definitionsPath)
	if strings.Count(text, original) != 1 {
		t.Fatalf("%q is not found once in the definitions file", original)
	}
	definitions, err := ParseDefinitions("copy.definitions.md", strings.Replace(text, original, changed, 1))
	if err != nil {
		t.Fatal(err)
	}
	failing := Failing(loadAll(t), held, definitions, runtime)
	if len(failing) == 0 {
		t.Fatal("no case fails")
	}
	failed := map[string]bool{}
	for _, name := range failing {
		failed[name] = true
		binding, bound := definitions.Bindings[name]
		if !bound || !strings.HasPrefix(binding.Definition, "V1") {
			t.Errorf("%s fails and does not use V1", name)
		}
	}
	if !failed["rules-decimal-places-invalid"] {
		t.Error("rules-decimal-places-invalid still passes with three places allowed")
	}
}
