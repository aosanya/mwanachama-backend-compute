package basicruns

import (
	"testing"
)

func loadResolution(t *testing.T) (Page, Resolution) {
	t.Helper()
	page := loadPage(readFile(t, basicPagePath))
	examples := loadExamples(readFile(t, examplesPagePath))
	return page, resolve(page, examples)
}

func TestEveryCaseOfThePageRuns(t *testing.T) {
	_, resolution := loadResolution(t)
	runtime := newRuntime(t)
	if len(resolution.Run) == 0 {
		t.Fatal("the page gave no case to run")
	}
	for _, c := range resolution.Run {
		t.Run(c.Name, func(t *testing.T) {
			outcome, err := execute(c, runtime)
			if err != nil {
				t.Fatalf("case %s cannot be read: %v", c.Name, err)
			}
			if err := judge(c, outcome); err != nil {
				t.Fatalf("%s: %s\nformula: %s\ninputs: %s", c.Name, err, c.Formula, c.Inputs)
			}
		})
	}
}

func TestHeldCasesAreListedWithTheirReason(t *testing.T) {
	_, resolution := loadResolution(t)
	if len(resolution.Held) == 0 {
		t.Fatal("the page has held cases and none was listed")
	}
	for _, held := range resolution.Held {
		t.Run(held.Case.Name, func(t *testing.T) {
			if held.Reason == "" {
				t.Fatal("a held case has no reason")
			}
			t.Skipf("held: %s", held.Reason)
		})
	}
}

func TestAliasesPointAtARowOfThePage(t *testing.T) {
	_, resolution := loadResolution(t)
	for _, name := range resolution.Missing {
		t.Errorf("%s is named in a section with no formula and has no row of its own", name)
	}
}

func TestEveryNamedCaseHasARow(t *testing.T) {
	_, resolution := loadResolution(t)
	for _, name := range resolution.NoRowFor {
		t.Errorf("%s is named by the page and has no row on formula-examples.md or another QA page", name)
	}
}

func TestCaseNamesAreUnique(t *testing.T) {
	page, _ := loadResolution(t)
	seen := map[string]string{}
	for _, c := range page.Cases {
		if c.Kind != KindOwn {
			continue
		}
		if first, found := seen[c.Name]; found {
			t.Errorf("%s has its own row in section %s and in section %s", c.Name, first, c.Section)
		}
		seen[c.Name] = c.Section
	}
}

func TestGapRowsNameAGapOfThePage(t *testing.T) {
	page, resolution := loadResolution(t)
	if len(page.Gaps) == 0 {
		t.Fatal("the page lists no gaps")
	}
	for _, c := range resolution.Gaps {
		if c.Note == "" {
			t.Errorf("a gap row in section %s has no rule text naming its gap", c.Section)
		}
	}
}

func caseNamed(t *testing.T, resolution Resolution, name string) Case {
	t.Helper()
	for _, c := range resolution.Run {
		if c.Name == name {
			return c
		}
	}
	t.Fatalf("no case named %s runs", name)
	return Case{}
}

func TestAChangedExpectedResultFailsThatCase(t *testing.T) {
	_, resolution := loadResolution(t)
	runtime := newRuntime(t)
	original := caseNamed(t, resolution, "basic-add-decimals")
	changed := original
	changed.Expect = "result: dec:165.71"
	outcome, err := execute(changed, runtime)
	if err != nil {
		t.Fatal(err)
	}
	if judge(changed, outcome) == nil {
		t.Fatal("a changed expected result did not fail the case")
	}
	outcome, err = execute(original, runtime)
	if err != nil || judge(original, outcome) != nil {
		t.Fatalf("the untouched case no longer passes: %v %v", err, judge(original, outcome))
	}
}

func TestAChangedProblemCodeFailsThatCase(t *testing.T) {
	_, resolution := loadResolution(t)
	runtime := newRuntime(t)
	changed := caseNamed(t, resolution, "basic-divide-integers-refused")
	changed.Expect = "problems: divide-by-zero"
	outcome, err := execute(changed, runtime)
	if err != nil {
		t.Fatal(err)
	}
	if judge(changed, outcome) == nil {
		t.Fatal("a changed problem code did not fail the case")
	}
}
