package functions

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const (
	referencePage = "../documentation/4. qa/library-reference.md"
	sourcesPage   = "../documentation/1. requirements/sources.md"
	firstCaseHash = "345008ae79f497d9c4999c1075c9ba95b63f68866d2c3ab2bfa6ad9dd893daab"
	firstKey      = "fin.balance_after_payment@1"
	firstID       = "REF-fin.balance_after_payment-01"
)

func readPage(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("the page %s cannot be read: %v", path, err)
	}
	return string(raw)
}

func servedOrders() map[string][]string {
	served := map[string][]string{}
	for _, candidate := range All() {
		served[candidate.Key()] = candidate.ParamOrder()
	}
	return served
}

func realGuard(t *testing.T) Guard {
	t.Helper()
	manifest, err := LoadManifest(referenceDir)
	if err != nil {
		t.Fatal(err)
	}
	return Guard{
		Dir:      "functions/reference",
		Cases:    loadReference(t),
		Manifest: manifest,
		Served:   servedOrders(),
		Page:     PageCaseIDs(readPage(t, referencePage)),
		Sources:  SourceIDs(readPage(t, sourcesPage)),
	}
}

func publishedCopy(g Guard) Guard {
	published := Manifest{}
	for key, version := range g.Manifest {
		published[key] = ManifestVersion{State: StatePublished, Cases: version.Cases}
	}
	g.Manifest = published
	return g
}

func copyCases(versions map[string][]Case) map[string][]Case {
	copied := map[string][]Case{}
	for key, cases := range versions {
		copied[key] = append([]Case{}, cases...)
	}
	return copied
}

func TestTheGuardPassesOnTheRecordedCases(t *testing.T) {
	report := realGuard(t).Check()
	if len(report.Failures) > 0 || len(report.Notices) > 0 {
		t.Fatalf("failures %v, notices %v", report.Failures, report.Notices)
	}
}

func TestTheGuardPassesWhenEveryVersionIsPublished(t *testing.T) {
	report := publishedCopy(realGuard(t)).Check()
	if len(report.Failures) > 0 || len(report.Notices) > 0 {
		t.Fatalf("failures %v, notices %v", report.Failures, report.Notices)
	}
}

func TestTheManifestHoldsEveryRecordedCase(t *testing.T) {
	g := realGuard(t)
	for key, cases := range g.Cases {
		version, recorded := g.Manifest[key]
		if !recorded {
			t.Errorf("%s is not in the manifest", key)
			continue
		}
		if len(version.Cases) != len(cases) {
			t.Errorf("%s: %d cases in the manifest, %d in the data file", key, len(version.Cases), len(cases))
		}
	}
}

func TestTheCanonicalFormOfTheFirstCaseIsTheOneTheDesignPageHashes(t *testing.T) {
	g := realGuard(t)
	first := g.Cases[firstKey][0]
	want := `{"id":"REF-fin.balance_after_payment-01","kind":"normal","inputs":{"balance":"120.50","payment":"45.25"},"expected":"75.25","tolerance":"exact","source":"S-01"}`
	if got := Canonical(first, g.Served[firstKey]); got != want {
		t.Fatalf("canonical form\n got %s\nwant %s", got, want)
	}
	if got := Hash(first, g.Served[firstKey]); got != firstCaseHash {
		t.Fatalf("hash %s, want %s", got, firstCaseHash)
	}
}

func TestACanonicalFormKeepsTheParameterOrderAndEscapesOnlyWhatMustBe(t *testing.T) {
	c := Case{
		ID: "REF-x-01", Kind: "normal", Expected: `a"b\c<d>&`, Tolerance: ToleranceExact, Source: "S-03",
		Inputs: map[string]InputValue{"b": {Text: "2"}, "a": {List: true, Items: []string{"1", "2"}}},
	}
	want := `{"id":"REF-x-01","kind":"normal","inputs":{"b":"2","a":["1","2"]},"expected":"a\"b\\c<d>&","tolerance":"exact","source":"S-03"}`
	if got := Canonical(c, []string{"b", "a"}); got != want {
		t.Fatalf("got %s want %s", got, want)
	}
}

func TestAnInputThatIsNotGivenIsLeftOutOfTheCanonicalForm(t *testing.T) {
	c := Case{ID: "REF-x-02", Kind: "refusal", Expected: RefusedText, Tolerance: ToleranceExact, Source: "S-02",
		Inputs: map[string]InputValue{"due": {Text: "2026-03-31"}}}
	if got := Canonical(c, []string{"due", "filed"}); !strings.Contains(got, `"inputs":{"due":"2026-03-31"}`) {
		t.Fatalf("got %s", got)
	}
}

func TestAnEditedCaseOfAPublishedVersionFailsTheGuardWithTheSentence(t *testing.T) {
	g := publishedCopy(realGuard(t))
	g.Cases = copyCases(g.Cases)
	g.Cases[firstKey][0].Expected = "75.26"
	report := g.Check()
	want := firstID + ": this case of published " + firstKey + " has changed (recorded " + firstCaseHash[:12] + ", now "
	if len(report.Failures) != 1 || !strings.HasPrefix(report.Failures[0], want) {
		t.Fatalf("failures %v", report.Failures)
	}
	if !strings.HasSuffix(report.Failures[0], "). A published version is never edited. Ship the fix as the next version of fin.balance_after_payment.") {
		t.Fatalf("sentence %q", report.Failures[0])
	}
}

func TestAnEditedCaseOfADraftVersionIsNotAFailure(t *testing.T) {
	g := realGuard(t)
	g.Cases = copyCases(g.Cases)
	g.Cases[firstKey][0].Expected = "75.26"
	if report := g.Check(); len(report.Failures) > 0 {
		t.Fatalf("failures %v", report.Failures)
	}
}

func TestARemovedCaseOfAPublishedVersionFailsTheGuard(t *testing.T) {
	g := publishedCopy(realGuard(t))
	g.Cases = copyCases(g.Cases)
	g.Cases[firstKey] = g.Cases[firstKey][1:]
	g.Page = map[string][]string{}
	report := g.Check()
	want := firstID + ": this case of published " + firstKey + " is missing. A case of a published version is never removed."
	if !contains(report.Failures, want) {
		t.Fatalf("failures %v", report.Failures)
	}
}

func TestAPublishedVersionWithNoDataFileFailsTheGuard(t *testing.T) {
	g := publishedCopy(realGuard(t))
	g.Cases = copyCases(g.Cases)
	delete(g.Cases, firstKey)
	want := firstKey + " is published but functions/reference/" + firstKey + ".json does not exist."
	if report := g.Check(); !contains(report.Failures, want) {
		t.Fatalf("failures %v", report.Failures)
	}
}

func TestAPublishedVersionTheLibraryNoLongerHasFailsTheGuard(t *testing.T) {
	g := publishedCopy(realGuard(t))
	served := map[string][]string{}
	for key, order := range g.Served {
		if key != firstKey {
			served[key] = order
		}
	}
	g.Served = served
	want := firstKey + " is published but the library no longer has it. A published version is never removed."
	if report := g.Check(); !contains(report.Failures, want) {
		t.Fatalf("failures %v", report.Failures)
	}
}

func TestACaseWithNoSourceIsNotAccepted(t *testing.T) {
	g := realGuard(t)
	g.Cases = copyCases(g.Cases)
	g.Cases[firstKey][2].Source = ""
	want := g.Cases[firstKey][2].ID + ": this case has no source id. A case with no source is not accepted."
	if report := g.Check(); !contains(report.Failures, want) {
		t.Fatalf("failures %v", report.Failures)
	}
}

func TestACaseWhoseSourceHasNoRowIsNotAccepted(t *testing.T) {
	g := realGuard(t)
	g.Cases = copyCases(g.Cases)
	g.Cases[firstKey][2].Source = "S-09"
	want := g.Cases[firstKey][2].ID + ": source S-09 has no row in documentation/1. requirements/sources.md."
	if report := g.Check(); !contains(report.Failures, want) {
		t.Fatalf("failures %v", report.Failures)
	}
}

func TestACaseOnThePageAndNotInTheDataFileFailsTheGuard(t *testing.T) {
	g := realGuard(t)
	g.Cases = copyCases(g.Cases)
	g.Cases[firstKey] = g.Cases[firstKey][:7]
	want := "REF-fin.balance_after_payment-08 is on the page and not in " + firstKey + ".json."
	if report := g.Check(); !contains(report.Failures, want) {
		t.Fatalf("failures %v", report.Failures)
	}
}

func TestACaseInTheDataFileAndNotOnThePageFailsTheGuard(t *testing.T) {
	g := realGuard(t)
	g.Page = map[string][]string{}
	for key, ids := range PageCaseIDs(readPage(t, referencePage)) {
		g.Page[key] = ids
	}
	g.Page[firstKey] = g.Page[firstKey][1:]
	want := firstID + " is in " + firstKey + ".json and not on the page."
	if report := g.Check(); !contains(report.Failures, want) {
		t.Fatalf("failures %v", report.Failures)
	}
}

func TestACaseAddedToAPublishedVersionIsANoticeAndNotAFailure(t *testing.T) {
	g := publishedCopy(realGuard(t))
	g.Cases = copyCases(g.Cases)
	extra := g.Cases[firstKey][0]
	extra.ID = "REF-fin.balance_after_payment-09"
	g.Cases[firstKey] = append(g.Cases[firstKey], extra)
	g.Page = map[string][]string{}
	for key, ids := range PageCaseIDs(readPage(t, referencePage)) {
		g.Page[key] = ids
	}
	g.Page[firstKey] = append(append([]string{}, g.Page[firstKey]...), extra.ID)
	report := g.Check()
	if len(report.Failures) > 0 {
		t.Fatalf("failures %v", report.Failures)
	}
	want := "REF-fin.balance_after_payment-09: new case of published " + firstKey + " is not in the manifest yet."
	if !contains(report.Notices, want) {
		t.Fatalf("notices %v", report.Notices)
	}
}

func TestTheNextVersionWithItsOwnCasesPassesBesideThePublishedOne(t *testing.T) {
	g := publishedCopy(realGuard(t))
	g.Cases = copyCases(g.Cases)
	next := Case{ID: "REF-fin.balance_after_payment-09", Kind: "normal", Expected: "75.25", Tolerance: ToleranceExact, Source: "S-03",
		Inputs: map[string]InputValue{"balance": {Text: "120.50"}, "payment": {Text: "45.25"}}}
	g.Cases["fin.balance_after_payment@2"] = []Case{next}
	g.Page = map[string][]string{}
	for key, ids := range PageCaseIDs(readPage(t, referencePage)) {
		g.Page[key] = ids
	}
	g.Page["fin.balance_after_payment@2"] = []string{next.ID}
	g.Manifest["fin.balance_after_payment@2"] = ManifestVersion{State: StateDraft, Cases: map[string]string{}}
	if report := g.Check(); len(report.Failures) > 0 {
		t.Fatalf("failures %v", report.Failures)
	}
}

func TestExtendingTheManifestAddsNewCasesAndNeverChangesAnEntry(t *testing.T) {
	g := realGuard(t)
	extended, failures, added := ExtendManifest(g.Manifest, g.Cases, g.Served)
	if len(failures) > 0 || len(added) > 0 {
		t.Fatalf("failures %v added %v", failures, added)
	}
	cases := copyCases(g.Cases)
	cases[firstKey] = append(cases[firstKey], Case{ID: "REF-fin.balance_after_payment-09", Kind: "normal", Expected: "1.00", Tolerance: ToleranceExact, Source: "S-03",
		Inputs: map[string]InputValue{"balance": {Text: "2.00"}, "payment": {Text: "1.00"}}})
	extended, failures, added = ExtendManifest(extended, cases, g.Served)
	if len(failures) > 0 || len(added) != 1 || added[0] != "REF-fin.balance_after_payment-09" {
		t.Fatalf("failures %v added %v", failures, added)
	}
	if extended[firstKey].Cases[firstID] != firstCaseHash {
		t.Fatal("an entry changed")
	}
	cases[firstKey][0].Expected = "75.26"
	_, failures, _ = ExtendManifest(extended, cases, g.Served)
	if len(failures) != 1 || !strings.HasPrefix(failures[0], firstID+": this case of "+firstKey+" has changed (recorded "+firstCaseHash[:12]) {
		t.Fatalf("failures %v", failures)
	}
}

func TestEditingARecordedFileOnDiskFailsTheGuardOnceTheVersionIsPublished(t *testing.T) {
	directory := t.TempDir()
	entries, err := os.ReadDir(referenceDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		raw, err := os.ReadFile(filepath.Join(referenceDir, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(directory, entry.Name()), raw, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	g := realGuard(t)
	published := publishedCopy(g).Manifest
	if err := published.Write(directory); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, firstKey+".json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	edited := strings.Replace(string(raw), `"expected": "75.25"`, `"expected": "75.26"`, 1)
	if edited == string(raw) {
		t.Fatal("the test copy was not edited")
	}
	if err := os.WriteFile(path, []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}
	cases, err := LoadCases(directory)
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := LoadManifest(directory)
	if err != nil {
		t.Fatal(err)
	}
	g.Cases, g.Manifest = cases, manifest
	report := g.Check()
	if len(report.Failures) != 1 || !strings.HasPrefix(report.Failures[0], firstID+": this case of published "+firstKey+" has changed") {
		t.Fatalf("failures %v", report.Failures)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	cases, _ = LoadCases(directory)
	g.Cases = cases
	if report := g.Check(); !contains(report.Failures, firstKey+" is published but functions/reference/"+firstKey+".json does not exist.") {
		t.Fatalf("failures %v", report.Failures)
	}
}

var signatureLine = regexp.MustCompile("(?m)^Signature: `([a-z]+\\.[a-z_]+@\\d+)\\(([^)]*)\\) -> ([a-z ]+)`")

func TestEachDefinitionHasTheSignatureOfThePage(t *testing.T) {
	matches := signatureLine.FindAllStringSubmatch(readPage(t, referencePage), -1)
	if len(matches) != 10 {
		t.Fatalf("the page has %d signatures, want 10", len(matches))
	}
	for _, match := range matches {
		found, ok := Find(match[1])
		if !ok {
			t.Errorf("%s is on the page and not in the library", match[1])
			continue
		}
		var names, types []string
		for _, part := range strings.Split(match[2], ", ") {
			pair := strings.SplitN(part, ": ", 2)
			names = append(names, pair[0])
			types = append(types, pair[1])
		}
		if strings.Join(found.ParamOrder(), ",") != strings.Join(names, ",") {
			t.Errorf("%s: parameters %v, the page says %v", match[1], found.ParamOrder(), names)
		}
		for i, input := range found.Definition.Inputs {
			if string(input.Type) != types[i] {
				t.Errorf("%s: input %s is %s, the page says %s", match[1], input.Name, input.Type, types[i])
			}
		}
		if string(found.Result) != match[3] {
			t.Errorf("%s: result %s, the page says %s", match[1], found.Result, match[3])
		}
	}
}

func contains(list []string, want string) bool {
	for _, item := range list {
		if item == want {
			return true
		}
	}
	return false
}
