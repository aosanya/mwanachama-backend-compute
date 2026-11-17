package library

import (
	"errors"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/formula"
)

func TestRefParsingAndWriting(t *testing.T) {
	good := []struct {
		text string
		want Ref
	}{
		{"fin.pmt@1", Ref{Library: "fin", Name: "pmt", Version: 1}},
		{"fin.pmt@12", Ref{Library: "fin", Name: "pmt", Version: 12}},
		{"fin.pmt", Ref{Library: "fin", Name: "pmt"}},
		{"loan_tools.late_fee@3", Ref{Library: "loan_tools", Name: "late_fee", Version: 3}},
	}
	for _, c := range good {
		got, problem := ParseRef(c.text)
		if problem != nil || got != c.want {
			t.Errorf("ParseRef(%q) = %v, %v; want %v", c.text, got, problem, c.want)
		}
	}
	if got := (Ref{Library: "fin", Name: "pmt", Version: 2}).String(); got != "fin.pmt@2" {
		t.Errorf("String with a version = %q", got)
	}
	if got := (Ref{Library: "fin", Name: "pmt"}).String(); got != "fin.pmt" {
		t.Errorf("String without a version = %q", got)
	}
	for _, text := range []string{"", "pmt", "fin.", ".pmt", "fin.pmt@", "fin.pmt@0", "fin.pmt@01", "fin.pmt@-1", "fin.pmt@x", "Fin.pmt@1", "fin.pmt@1@2", "fin.pmt.x@1", "fin pmt@1"} {
		if _, problem := ParseRef(text); problem == nil {
			t.Errorf("ParseRef(%q) was accepted", text)
		}
	}
}

func TestVersionsStartAtOneAndGoUpByOne(t *testing.T) {
	registry := NewRegistry()
	if err := registry.Publish(pmtVersion(t, 2)); !errors.Is(err, ErrOutOfOrder) {
		t.Fatalf("publishing 2 first gave %v, want ErrOutOfOrder", err)
	}
	if err := registry.Publish(pmtVersion(t, 1)); err != nil {
		t.Fatal(err)
	}
	if err := registry.Publish(pmtVersion(t, 3)); !errors.Is(err, ErrOutOfOrder) {
		t.Fatalf("publishing 3 after 1 gave %v, want ErrOutOfOrder", err)
	}
	if err := registry.Publish(pmtVersion(t, 2)); err != nil {
		t.Fatal(err)
	}
	if got := registry.Versions("fin", "pmt"); !reflect.DeepEqual(got, []int{1, 2}) {
		t.Fatalf("Versions = %v, want [1 2]", got)
	}
	if got := registry.Versions("fin", "none"); len(got) != 0 {
		t.Fatalf("Versions of an unknown function = %v", got)
	}
}

func TestAPublishedVersionIsNeverEditedOrReplaced(t *testing.T) {
	registry := finRegistry(t)
	before := runLoan(t, registry, "fin.pmt@1(principal, rate, months)")
	replacement := pmtVersion(t, 1)
	replacement.Apply = pmtRoundedToCents(t)
	if err := registry.Publish(replacement); !errors.Is(err, ErrAlreadyPublished) {
		t.Fatalf("publishing 1 again gave %v, want ErrAlreadyPublished", err)
	}
	if after := runLoan(t, registry, "fin.pmt@1(principal, rate, months)"); after != before {
		t.Fatalf("fin.pmt@1 changed from %s to %s", before, after)
	}
}

func TestRegistryOffersNoWayToStopRemoveOrEditAVersion(t *testing.T) {
	forbidden := []string{"Stop", "Remove", "Delete", "Unpublish", "Replace", "Edit", "Update", "Retire", "Disable", "Unmark", "Revoke"}
	registryType := reflect.TypeOf(&Registry{})
	for i := 0; i < registryType.NumMethod(); i++ {
		name := registryType.Method(i).Name
		for _, prefix := range forbidden {
			if strings.HasPrefix(name, prefix) {
				t.Errorf("Registry has the method %s", name)
			}
		}
	}
}

func TestEveryVersionFromTwoCarriesAThreePartNote(t *testing.T) {
	cases := map[string]Note{
		"no changed part": {Matters: pmtNote.Matters, Example: pmtNote.Example},
		"no matters part": {Changed: pmtNote.Changed, Example: pmtNote.Example},
		"no example":      {Changed: pmtNote.Changed, Matters: pmtNote.Matters},
		"half an example": {Changed: pmtNote.Changed, Matters: pmtNote.Matters, Example: Example{Input: "12,000", Old: "564"}},
	}
	for name, note := range cases {
		registry := NewRegistry()
		if err := registry.Publish(pmtVersion(t, 1)); err != nil {
			t.Fatal(err)
		}
		release := pmtVersion(t, 2)
		release.Note = note
		if err := registry.Publish(release); !errors.Is(err, ErrNoteIncomplete) {
			t.Errorf("%s: publishing 2 gave %v, want ErrNoteIncomplete", name, err)
		}
		if got := registry.Versions("fin", "pmt"); !reflect.DeepEqual(got, []int{1}) {
			t.Errorf("%s: a refused version was kept: %v", name, got)
		}
	}
	registry := finRegistry(t)
	note, found := registry.NoteOf("fin", "pmt", 2)
	if !found || note != pmtNote {
		t.Errorf("NoteOf fin.pmt@2 = %v, %v", note, found)
	}
	if _, found := registry.NoteOf("fin", "pmt", 1); found {
		t.Error("fin.pmt@1 has no earlier version, so it carries no note")
	}
}

func TestAReleaseNeedsANameASignatureAndAnImplementation(t *testing.T) {
	bad := map[string]func(*Release){
		"no library":        func(r *Release) { r.Library = "" },
		"no name":           func(r *Release) { r.Name = "" },
		"bad library name":  func(r *Release) { r.Library = "Fin" },
		"no implementation": func(r *Release) { r.Apply = nil },
		"no result type":    func(r *Release) { r.Signature.Result = "" },
	}
	for name, change := range bad {
		release := pmtVersion(t, 1)
		change(&release)
		if err := NewRegistry().Publish(release); !errors.Is(err, ErrBadRelease) {
			t.Errorf("%s: got %v, want ErrBadRelease", name, err)
		}
	}
}

func TestASignatureHandedOutCannotEditTheRegistry(t *testing.T) {
	registry := finRegistry(t)
	signature, found := registry.Function("fin", "pmt", 1)
	if !found || len(signature.Params) != 3 {
		t.Fatalf("Function = %v, %v", signature, found)
	}
	signature.Params[0].Name = "changed"
	signature.Result = formula.TypeText
	again, _ := registry.Function("fin", "pmt", 1)
	if again.Params[0].Name != "principal" || again.Result != formula.TypeDecimal {
		t.Fatalf("a caller edited the stored signature: %v", again)
	}
	if _, found := registry.Function("fin", "pmt", 7); found {
		t.Error("Function found fin.pmt@7")
	}
}

func TestResolvingACallByNameAndVersion(t *testing.T) {
	registry := finRegistry(t)
	release, problem := registry.Resolve(Ref{Library: "fin", Name: "pmt", Version: 2})
	if problem != nil || release.Version != 2 || release.Apply == nil {
		t.Fatalf("Resolve fin.pmt@2 = %v, %v", release, problem)
	}

	cases := []struct {
		ref      Ref
		code     string
		sentence string
	}{
		{Ref{Library: "fin", Name: "pmt"}, "version-missing", "fin.pmt needs a version. Available: fin.pmt@1, fin.pmt@2."},
		{Ref{Library: "fin", Name: "pmt", Version: 7}, "version-not-found", "`fin.pmt@7` does not exist. Latest is `fin.pmt@2`."},
		{Ref{Library: "fin", Name: "later", Version: 1}, "unknown-function", "The function fin.later is not known."},
	}
	for _, c := range cases {
		_, problem := registry.Resolve(c.ref)
		if problem == nil || problem.Code != c.code || problem.Sentence != c.sentence {
			t.Errorf("Resolve(%v) = %v, want %s: %s", c.ref, problem, c.code, c.sentence)
		}
	}
}

func TestTheEngineRunsTheVersioningCases(t *testing.T) {
	registry := finRegistry(t)
	cases := []struct {
		name    string
		formula string
		want    string
	}{
		{"pmt-version-1-loan", "fin.pmt@1(principal, rate, months)", "564"},
		{"pmt-version-2-loan", "fin.pmt@2(principal, rate, months)", "564.88"},
		{"same-host-newer-minus-older", "fin.pmt@2(principal, rate, months) - fin.pmt@1(principal, rate, months)", "0.88"},
		{"same-host-older-minus-newer", "fin.pmt@1(principal, rate, months) - fin.pmt@2(principal, rate, months)", "-0.88"},
	}
	for _, c := range cases {
		if got := runLoan(t, registry, c.formula); got != c.want {
			t.Errorf("%s: got %s, want %s", c.name, got, c.want)
		}
	}
}

func TestTheEngineRefusesAnUnversionedCallAndAVersionThatIsNotThere(t *testing.T) {
	registry := finRegistry(t)
	cases := []struct {
		name    string
		formula string
		code    string
	}{
		{"unversioned-call-refused", "fin.pmt(principal, rate, months)", "version-missing"},
		{"unknown-version-refused", "fin.pmt@7(principal, rate, months)", "version-not-found"},
	}
	for _, c := range cases {
		_, problems := planned(t, registry, c.formula)
		if len(problems) != 1 || problems[0].Code != c.code {
			t.Errorf("%s: problems = %v, want one %s", c.name, problems, c.code)
		}
	}
	_, problems := planned(t, registry, "fin.pmt(principal, rate, months)")
	want := "fin.pmt needs a version. Available: fin.pmt@1, fin.pmt@2."
	if len(problems) != 1 || problems[0].Sentence != want {
		t.Errorf("the sentence is %v, want %s", problems, want)
	}
}

func TestTwoVersionsOfOneFunctionRunSideBySideWithTheirOwnResults(t *testing.T) {
	registry := finRegistry(t)
	older, problems := planned(t, registry, "fin.pmt@1(principal, rate, months)")
	if len(problems) > 0 {
		t.Fatal(problems)
	}
	newer, problems := planned(t, registry, "fin.pmt@2(principal, rate, months)")
	if len(problems) > 0 {
		t.Fatal(problems)
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			value, problem := older.Run(loanValues(t))
			if problem != nil || value.String() != "564" {
				t.Errorf("fin.pmt@1 gave %v, %v", value, problem)
			}
		}()
		go func() {
			defer wg.Done()
			value, problem := newer.Run(loanValues(t))
			if problem != nil || value.String() != "564.88" {
				t.Errorf("fin.pmt@2 gave %v, %v", value, problem)
			}
		}()
	}
	wg.Wait()
}

func TestReadingWhilePublishingIsSafe(t *testing.T) {
	registry := NewRegistry()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, version := range []int{1, 2} {
			if err := registry.Publish(pmtVersion(t, version)); err != nil {
				t.Error(err)
			}
		}
	}()
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				registry.Versions("fin", "pmt")
				registry.Function("fin", "pmt", 1)
				registry.Status("fin", "pmt", 1)
			}
		}()
	}
	wg.Wait()
	if got := registry.Versions("fin", "pmt"); !reflect.DeepEqual(got, []int{1, 2}) {
		t.Fatalf("Versions = %v", got)
	}
}

func TestMarkingAVersionKnownWrongNamesItsReplacementAndKeepsItRunning(t *testing.T) {
	registry := finRegistry(t)
	if err := registry.MarkKnownWrong("fin", "pmt", 1, 2); err != nil {
		t.Fatal(err)
	}
	status := registry.Status("fin", "pmt", 1)
	if !status.Exists || !status.KnownWrong || status.Replacement != 2 || status.Latest != 2 || !status.Newer {
		t.Fatalf("Status = %+v", status)
	}
	for round := 0; round < 3; round++ {
		if got := runLoan(t, registry, "fin.pmt@1(principal, rate, months)"); got != "564" {
			t.Fatalf("round %d: a known-wrong version gave %s, want 564", round, got)
		}
	}
	if registry.Status("fin", "pmt", 2).KnownWrong {
		t.Error("marking fin.pmt@1 marked fin.pmt@2")
	}
}

func TestAKnownWrongMarkIsRefusedWhenItDoesNotMakeSense(t *testing.T) {
	registry := finRegistry(t)
	cases := map[string]func() error{
		"no such version":          func() error { return registry.MarkKnownWrong("fin", "pmt", 9, 2) },
		"no such replacement":      func() error { return registry.MarkKnownWrong("fin", "pmt", 1, 9) },
		"replacement is not newer": func() error { return registry.MarkKnownWrong("fin", "pmt", 2, 1) },
		"replacement is itself":    func() error { return registry.MarkKnownWrong("fin", "pmt", 1, 1) },
		"no such function":         func() error { return registry.MarkKnownWrong("fin", "none", 1, 2) },
	}
	for name, mark := range cases {
		if err := mark(); !errors.Is(err, ErrBadMark) {
			t.Errorf("%s: got %v, want ErrBadMark", name, err)
		}
	}
	if err := registry.MarkKnownWrong("fin", "pmt", 1, 2); err != nil {
		t.Fatal(err)
	}
	if err := registry.MarkKnownWrong("fin", "pmt", 1, 2); !errors.Is(err, ErrBadMark) {
		t.Errorf("marking twice gave %v, want ErrBadMark", err)
	}
}

func TestStatusSaysWhatANoticeNeeds(t *testing.T) {
	registry := finRegistry(t)
	older := registry.Status("fin", "pmt", 1)
	if !older.Exists || older.Latest != 2 || !older.Newer || older.KnownWrong || older.Stopped {
		t.Fatalf("Status of fin.pmt@1 = %+v", older)
	}
	if older.Note != pmtNote {
		t.Errorf("the notice carries %v, want the note of the newer version", older.Note)
	}
	newest := registry.Status("fin", "pmt", 2)
	if !newest.Exists || newest.Newer || newest.Latest != 2 {
		t.Fatalf("Status of fin.pmt@2 = %+v", newest)
	}
	missing := registry.Status("fin", "pmt", 7)
	if missing.Exists || missing.Latest != 2 {
		t.Fatalf("Status of fin.pmt@7 = %+v", missing)
	}
	none := registry.Status("fin", "none", 1)
	if none.Exists || none.Latest != 0 {
		t.Fatalf("Status of an unknown function = %+v", none)
	}
}

func TestOnlyTheStopSourceStopsAVersionAndEveryAgencyFollowsAtOnce(t *testing.T) {
	registry := finRegistry(t)
	stops := newStopTable()
	registry.ReadStopsFrom(stops)

	riverbend, problems := planned(t, registry, "fin.pmt@1(principal, rate, months)")
	if len(problems) > 0 {
		t.Fatal(problems)
	}
	lakeside, problems := planned(t, registry, "fin.pmt@1(principal, rate, months)")
	if len(problems) > 0 {
		t.Fatal(problems)
	}
	if value, problem := riverbend.Run(loanValues(t)); problem != nil || value.String() != "564" {
		t.Fatalf("before the stop: %v, %v", value, problem)
	}

	stops.put(Ref{Library: "fin", Name: "pmt", Version: 1}, "it pays out too little")

	want := "`fin.pmt@1` is stopped by the platform owner and cannot run. Reason: it pays out too little."
	for name, program := range map[string]*formula.Program{"riverbend": riverbend, "lakeside": lakeside} {
		_, problem := program.Run(loanValues(t))
		if problem == nil || problem.Code != "version-stopped" || problem.Sentence != want {
			t.Errorf("%s after the stop: %v", name, problem)
		}
	}
	if got := runLoan(t, registry, "fin.pmt@2(principal, rate, months)"); got != "564.88" {
		t.Errorf("the other version gave %s", got)
	}
	status := registry.Status("fin", "pmt", 1)
	if !status.Stopped || status.StopReason != "it pays out too little" {
		t.Errorf("Status = %+v", status)
	}
}

func TestAStopWithNoReasonStillGivesAPlainSentence(t *testing.T) {
	registry := finRegistry(t)
	stops := newStopTable()
	registry.ReadStopsFrom(stops)
	stops.put(Ref{Library: "fin", Name: "pmt", Version: 2}, "")
	program, problems := planned(t, registry, "fin.pmt@2(principal, rate, months)")
	if len(problems) > 0 {
		t.Fatal(problems)
	}
	_, problem := program.Run(loanValues(t))
	want := "`fin.pmt@2` is stopped by the platform owner and cannot run."
	if problem == nil || problem.Sentence != want {
		t.Fatalf("got %v, want %s", problem, want)
	}
}

func TestApplyToAVersionThatIsNotThereGivesThePageSentence(t *testing.T) {
	registry := finRegistry(t)
	_, problem := registry.Apply("fin", "pmt", 7, nil)
	want := "`fin.pmt@7` does not exist. Latest is `fin.pmt@2`."
	if problem == nil || problem.Code != "version-not-found" || problem.Sentence != want {
		t.Fatalf("got %v", problem)
	}
}

func TestUsesListsEveryLibraryCallWithItsVersionAndPlace(t *testing.T) {
	tree := parsed(t, "fin.pmt@1(principal, rate, months) + fin.pmt@1(principal, rate, months) + fin.pmt(principal, rate, months) + round(rate, 2)")
	uses := Uses(tree)
	if len(uses) != 3 {
		t.Fatalf("Uses = %v, want 3 library calls and no built-in", uses)
	}
	versions := []int{uses[0].Version, uses[1].Version, uses[2].Version}
	if !reflect.DeepEqual(versions, []int{1, 1, 0}) {
		t.Fatalf("versions = %v", versions)
	}
	for _, use := range uses {
		if use.Library != "fin" || use.Name != "pmt" {
			t.Errorf("use %v", use)
		}
	}
	if reflect.DeepEqual(uses[0].Path, uses[1].Path) {
		t.Errorf("two calls share the place %v", uses[0].Path)
	}
	part, found := formula.PartAt(tree, uses[1].Path)
	if !found || part.Text != "fin.pmt@1(principal, rate, months)" {
		t.Errorf("the place of the second use gives %v, %v", part, found)
	}
	if got := Uses(parsed(t, "rate * 2")); len(got) != 0 {
		t.Errorf("a formula with no library call gave %v", got)
	}
}

func TestThePackageReadsNoClockAndNoRandomNumbers(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		if _, err := os.Stat(file); err != nil {
			t.Fatal(err)
		}
		parsedFile, err := parser.ParseFile(token.NewFileSet(), file, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, spec := range parsedFile.Imports {
			path := strings.Trim(spec.Path.Value, `"`)
			switch path {
			case "time", "math/rand", "math/rand/v2", "context", "os", "net/http", "database/sql":
				t.Errorf("%s imports %s", file, path)
			}
			if strings.Contains(path, "mwanachama") && !strings.HasPrefix(path, "github.com/aosanya/mwanachama-backend-compute") {
				t.Errorf("%s imports %s", file, path)
			}
		}
	}
}
