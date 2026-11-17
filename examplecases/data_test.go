package examplecases

import (
	"os"
	"strings"
	"testing"
)

const (
	caseDir     = "../testdata"
	repoRoot    = ".."
	qaPagePath  = "../documentation/4. qa/formula-examples.md"
	caseDirName = "testdata"
)

func loadAll(t *testing.T) []Case {
	t.Helper()
	cases, err := Load(caseDir)
	if err != nil {
		t.Fatal(err)
	}
	return cases
}

func TestTheCaseFileLoadsAllOneHundredAndElevenCases(t *testing.T) {
	cases := loadAll(t)
	var lang, rules int
	for _, c := range cases {
		switch {
		case strings.HasPrefix(c.Name, "lang-"):
			lang++
		case strings.HasPrefix(c.Name, "rules-"):
			rules++
		}
	}
	if len(cases) != 111 || lang != 55 || rules != 56 {
		t.Fatalf("cases %d, lang %d, rules %d", len(cases), lang, rules)
	}
}

func TestTheCaseFileHoldsTheRowsOfTheQAPageAndNoOther(t *testing.T) {
	page, err := os.ReadFile(qaPagePath)
	if err != nil {
		t.Fatal(err)
	}
	file, err := os.ReadFile(caseDir + "/formula-language.cases.md")
	if err != nil {
		t.Fatal(err)
	}
	want := rowsOf(string(page))
	got := rowsOf(string(file))
	if len(want) != len(got) {
		t.Fatalf("the QA page has %d case rows and the case file %d", len(want), len(got))
	}
	for i := range want {
		if want[i] != got[i] {
			t.Fatalf("row %d differs:\n  page: %s\n  file: %s", i+1, want[i], got[i])
		}
	}
}

func rowsOf(text string) []string {
	var rows []string
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, "| lang-") || strings.HasPrefix(line, "| rules-") {
			if strings.HasSuffix(line, "| "+pagePath+" |") {
				rows = append(rows, line)
			}
		}
	}
	return rows
}

func TestEveryExampleOnTheDesignPageHasARowAndEveryRowAnExample(t *testing.T) {
	failures := CheckPages(repoRoot, caseDirName, loadAll(t))
	if len(failures) != 0 {
		t.Fatalf("%d failures, the first: %s", len(failures), failures[0])
	}
}
