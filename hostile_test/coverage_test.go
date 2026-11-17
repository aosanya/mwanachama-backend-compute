package hostile

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

type heldCase struct {
	id     string
	reason string
}

const (
	heldValidate = "needs Validate, which reads input values and gives list-too-long and the text and number limits; ISS-26 builds it"
	heldBatch    = "needs batch evaluation, ISS-29"
)

var namedTests = map[string]string{
	"L08": "TestNothingIsLeftBehindAfterTheHostileCases",
	"L10": "TestTenHostileFormulasAtOnceOnOneEngine",
	"C10": "TestSelfCallsSlippedPastTheSaveCheckAreStopped",
	"C12": "TestSelfCallsSlippedPastTheSaveCheckAreStopped",
	"C13": "TestSelfCallsSlippedPastTheSaveCheckAreStopped",
	"O14": "TestSameResultOnEveryRunFreshAndBusy",
	"O16": "TestSumOfManyDecimalsIsTheSameInTwoProcesses",
}

var heldCases = []heldCase{
	{"L06", "withdrawn: there is no time limit (decision 19)"},
	{"L09", "open (decision 19): whether and how a host's own cancel is honoured"},
	{"C08", "a callee must already be published, so no chain past the call depth can be built to slip past the save check; C10, C12 and C13 cover a host that loops"},
	{"I02", heldValidate},
	{"I05", heldValidate},
	{"I12", heldValidate},
	{"I13", "the language has no join; the list length limit of a value belongs to Validate (ISS-26)"},
	{"B01", heldBatch},
	{"B02", heldBatch},
	{"B03", heldBatch},
	{"B04", heldBatch},
	{"B05", heldBatch},
}

var rowOfPage = regexp.MustCompile(`(?m)^\| ([PLCIOB][0-9]{2}) \|`)

func idsOfPage(t *testing.T) []string {
	t.Helper()
	text, err := os.ReadFile(filepath.Join("..", "documentation", "4. qa", "hostile-formulas.md"))
	if err != nil {
		t.Fatalf("the QA page cannot be read: %v", err)
	}
	var ids []string
	for _, match := range rowOfPage.FindAllStringSubmatch(string(text), -1) {
		ids = append(ids, match[1])
	}
	return ids
}

func idsBuilt() map[string]bool {
	built := map[string]bool{}
	for _, c := range allSaveCases() {
		built[c.id] = true
	}
	for _, c := range runCases {
		built[c.id] = true
	}
	for id := range namedTests {
		built[id] = true
	}
	return built
}

func TestEveryCaseOfThePageIsBuiltOrHeldWithAReason(t *testing.T) {
	built := idsBuilt()
	held := map[string]string{}
	for _, c := range heldCases {
		held[c.id] = c.reason
		if c.reason == "" {
			t.Errorf("%s is held with no reason", c.id)
		}
	}
	ids := idsOfPage(t)
	if len(ids) < 80 {
		t.Fatalf("only %d case rows were read from the page", len(ids))
	}
	var missing []string
	for _, id := range ids {
		if !built[id] && held[id] == "" {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		t.Fatalf("cases neither built nor held: %s", strings.Join(missing, ", "))
	}
}

func TestNoCaseIsBothBuiltAndHeldAsUnbuilt(t *testing.T) {
	built := idsBuilt()
	page := map[string]bool{}
	for _, id := range idsOfPage(t) {
		page[id] = true
	}
	for id := range built {
		if !page[id] {
			t.Errorf("%s is built but is not a row of the page", id)
		}
	}
	for _, c := range heldCases {
		if !page[c.id] {
			t.Errorf("%s is held but is not a row of the page", c.id)
		}
	}
}

func TestHeldCasesAreListedWithTheirReason(t *testing.T) {
	for _, c := range heldCases {
		c := c
		t.Run(c.id, func(t *testing.T) {
			t.Skip("held: " + c.reason)
		})
	}
}

func TestNoCaseReadsAClock(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil || len(files) == 0 {
		t.Fatalf("no files found: %v", err)
	}
	for _, name := range files {
		parsed, err := parser.ParseFile(token.NewFileSet(), name, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		for _, imported := range parsed.Imports {
			if imported.Path.Value == `"time"` {
				t.Errorf("%s imports time", name)
			}
		}
	}
}
