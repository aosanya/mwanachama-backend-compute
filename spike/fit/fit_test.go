package fit

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"text/tabwriter"

	"github.com/aosanya/mwanachama-backend-compute/spike/engine/cel"
	"github.com/aosanya/mwanachama-backend-compute/spike/engine/exprlang"
	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
)

const resultsDirectory = "FIT_RESULTS_DIR"

var redundantBrackets = map[int]bool{18: true}

type outcome struct {
	engine     string
	number     int
	converts   bool
	sameAsOwn  bool
	idempotent bool
	canonical  int
	roundTrips bool
	readsAsOwn bool
	equalsPage bool
	remarks    []string
}

var outcomes []outcome

func keep(o outcome) {
	sectionLock.Lock()
	defer sectionLock.Unlock()
	outcomes = append(outcomes, o)
}

func TestMain(m *testing.M) {
	code := m.Run()
	if dir := os.Getenv(resultsDirectory); dir != "" && code == 0 {
		if err := os.WriteFile(filepath.Join(dir, "fit-table.txt"), []byte(renderTable()), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			code = 1
		}
	}
	os.Exit(code)
}

func assess(s subject, own subject, c harness.Case) outcome {
	o := outcome{engine: s.name, number: c.Number}
	tree, treeErr := s.neutral(c.Source)
	o.converts = treeErr == nil
	if treeErr != nil {
		o.remarks = append(o.remarks, "tree: "+treeErr.Error())
	}
	reference, _ := own.neutral(c.Source)
	o.sameAsOwn = o.converts && sameTree(tree, reference)
	if o.converts && !o.sameAsOwn {
		o.remarks = append(o.remarks, "tree differs from own: "+tree.String())
	}
	printed, printErr := s.printed(c.Source)
	if printErr != nil {
		o.remarks = append(o.remarks, "print: "+printErr.Error())
		return o
	}
	again, err := s.printed(printed)
	o.idempotent = err == nil && again == printed
	if !o.idempotent {
		o.remarks = append(o.remarks, fmt.Sprintf("second print: %q", again))
	}
	o.canonical = convergingVariants(s, c.Source, printed, &o)
	o.equalsPage = printed == c.Source
	if !o.equalsPage {
		o.remarks = append(o.remarks, fmt.Sprintf("prints %q", printed))
	}
	back, err := s.neutral(printed)
	o.roundTrips = o.converts && err == nil && sameTree(tree, back)
	if !o.roundTrips {
		o.remarks = append(o.remarks, "round trip changed the tree")
	}
	o.readsAsOwn = readsAsOwn(own, c.Source, printed)
	if !o.readsAsOwn {
		o.remarks = append(o.remarks, "own reads the printed text as another tree")
	}
	return o
}

func readsAsOwn(own subject, source, printed string) bool {
	want, err := own.neutral(source)
	if err != nil {
		return false
	}
	got, err := own.neutral(printed)
	return err == nil && sameTree(want, got)
}

func convergingVariants(s subject, source, standard string, o *outcome) int {
	converged := 0
	for style, variant := range variants(source) {
		text, err := s.printed(variant)
		switch {
		case err != nil:
			o.remarks = append(o.remarks, fmt.Sprintf("variant %d: %v", style, err))
		case text != standard:
			o.remarks = append(o.remarks, fmt.Sprintf("variant %d prints %q", style, text))
		default:
			converged++
		}
	}
	return converged
}

func ownSubject(t testing.TB) subject {
	for _, s := range subjects(t) {
		if s.name == "own" {
			return s
		}
	}
	t.Fatal("own engine missing")
	return subject{}
}

func TestFit(t *testing.T) {
	reference := ownSubject(t)
	for _, s := range subjects(t) {
		for _, c := range harness.Cases() {
			t.Run(fmt.Sprintf("%s/formula-%02d", s.name, c.Number), func(t *testing.T) {
				o := assess(s, reference, c)
				keep(o)
				t.Logf("converts=%s sameAsOwn=%s idempotent=%s canonical=%d/%d roundTrips=%s readsAsOwn=%s equalsPage=%s %s",
					yes(o.converts), yes(o.sameAsOwn), yes(o.idempotent), o.canonical, variantStyles,
					yes(o.roundTrips), yes(o.readsAsOwn), yes(o.equalsPage), strings.Join(o.remarks, "; "))
				if s.name == "own" {
					requireOwnStandard(t, o)
				}
			})
		}
	}
}

func requireOwnStandard(t *testing.T, o outcome) {
	t.Helper()
	if !o.converts || !o.idempotent || o.canonical != variantStyles || !o.roundTrips || !o.readsAsOwn {
		t.Errorf("own breaks the standard shape: %+v", o)
	}
	if !o.equalsPage && !redundantBrackets[o.number] {
		t.Errorf("own prints formula %d differently from the page: %v", o.number, o.remarks)
	}
}

func TestInterfaces(t *testing.T) {
	if got := len(subjects(t)); got != 3 {
		t.Fatalf("want 3 tree engines, got %d", got)
	}
}

type standardCase struct {
	source   string
	standard string
}

var bracketCases = []standardCase{
	{"(a + b) * c", "(a + b) * c"},
	{"a + (b * c)", "a + b * c"},
	{"a - (b - c)", "a - (b - c)"},
	{"(a - b) - c", "a - b - c"},
	{"a - b - c", "a - b - c"},
	{"(a * b) / c", "a * b / c"},
	{"a * (b / c)", "a * (b / c)"},
	{"a || (b && c)", "a || b && c"},
	{"(a && b) || c", "a && b || c"},
	{"(a || b) && c", "(a || b) && c"},
	{"a && (b && c)", "a && (b && c)"},
	{"!(a && b)", "!(a && b)"},
	{"!!a", "!!a"},
	{"-(a + b)", "-(a + b)"},
	{"-a * b", "-a * b"},
	{"((a))", "a"},
	{"(a ? b : c) ? d : e", "(a ? b : c) ? d : e"},
	{"a ? (b ? c : d) : e", "a ? (b ? c : d) : e"},
	{"a ? b : (c ? d : e)", "a ? b : c ? d : e"},
	{"a ? b : c ? d : e", "a ? b : c ? d : e"},
	{"(a ? b : c) + d", "(a ? b : c) + d"},
	{"a == b < c", "a == b < c"},
	{"a < b == c", "a < b == c"},
	{"a == (b < c)", "a == b < c"},
	{"(a == b) < c", "(a == b) < c"},
}

func TestBrackets(t *testing.T) {
	reference := ownSubject(t)
	for _, s := range subjects(t) {
		for _, bc := range bracketCases {
			t.Run(s.name+"/"+bc.source, func(t *testing.T) {
				printed, err := s.printed(bc.source)
				if err != nil {
					record("brackets", "%s\t%s\tERROR\t%v", s.name, bc.source, err)
					t.Logf("error: %v", err)
					return
				}
				tree, _ := s.neutral(bc.source)
				want, _ := reference.neutral(bc.source)
				back, _ := s.neutral(printed)
				same, trips, reads := sameTree(tree, want), sameTree(tree, back), readsAsOwn(reference, bc.source, printed)
				record("brackets", "%s\t%s\t%s\t%s\t%s\t%s\t%s", s.name, bc.source, printed, yes(printed == bc.standard), yes(same), yes(trips), yes(reads))
				t.Logf("prints %q standard %q standardText=%s sameTreeAsOwn=%s roundTrips=%s readsAsOwn=%s", printed, bc.standard, yes(printed == bc.standard), yes(same), yes(trips), yes(reads))
				if s.name == "own" && (printed != bc.standard || !same || !trips) {
					t.Errorf("own printed %q for %q, want %q", printed, bc.source, bc.standard)
				}
			})
		}
	}
}

var literalCases = []standardCase{
	{"12.50", "12.50"},
	{"0.010", "0.010"},
	{"0.90", "0.90"},
	{"1000.00", "1000.00"},
	{"x * 0.90", "x * 0.90"},
	{"round(x * 0.010, 2)", "round(x * 0.010, 2)"},
	{"[0.00, 1000.00]", "[0.00, 1000.00]"},
	{"7", "7"},
	{"0.75f", "0.75f"},
	{"2f", "2f"},
	{"2.0f", "2f"},
	{`"a \"q\" b"`, `"a \"q\" b"`},
	{`"two  spaces"`, `"two  spaces"`},
	{"true", "true"},
}

func TestLiteralFidelity(t *testing.T) {
	reference := ownSubject(t)
	for _, s := range subjects(t) {
		for _, lc := range literalCases {
			t.Run(s.name+"/"+lc.source, func(t *testing.T) {
				printed, err := s.printed(lc.source)
				if err != nil {
					record("literals", "%s\t%s\tERROR\t%v", s.name, lc.source, err)
					t.Logf("error: %v", err)
					return
				}
				tree, _ := s.neutral(lc.source)
				back, _ := s.neutral(printed)
				record("literals", "%s\t%s\t%s\t%s\t%s\t%s\t%s", s.name, lc.source, printed, yes(printed == lc.standard), yes(sameTree(tree, back)), yes(readsAsOwn(reference, lc.source, printed)), tree.String())
				t.Logf("prints %q standard %q keeps=%s tree=%s", printed, lc.standard, yes(printed == lc.standard), tree.String())
				if s.name == "own" && printed != lc.standard {
					t.Errorf("own printed %q for %q, want %q", printed, lc.source, lc.standard)
				}
			})
		}
	}
}

type part struct {
	name  string
	probe string
}

var parts = []part{
	{"integer", "7"}, {"decimal", "12.50"}, {"float", "0.75f"}, {"text", `"A"`}, {"boolean", "true"},
	{"name", "a"}, {"list", "[1, 2, 3]"},
	{"+", "a + b"}, {"-", "a - b"}, {"*", "a * b"}, {"/", "a / b"}, {"%", "a % b"},
	{">=", "a >= b"}, {"<=", "a <= b"}, {">", "a > b"}, {"<", "a < b"}, {"==", "a == b"}, {"!=", "a != b"},
	{"&&", "a && b"}, {"||", "a || b"}, {"!", "!a"}, {"unary minus", "-a"},
	{"brackets", "(a)"}, {"conditional", "c ? a : b"}, {"nested conditional", "c ? a : d ? b : e"},
	{"round", "round(x, 2)"}, {"sum", "sum(x)"}, {"count", "count(x)"}, {"avg", "avg(x)"},
	{"min", "min(x)"}, {"max", "max(x)"}, {"to_float", "to_float(x)"}, {"to_decimal", "to_decimal(x, 2)"},
	{"days_between", "days_between(a, b)"}, {"months_between", "months_between(a, b)"},
	{"years_between", "years_between(a, b)"}, {"versioned call", "late_days@1(a, b)"},
}

func TestParts(t *testing.T) {
	for _, s := range subjects(t) {
		for _, p := range parts {
			t.Run(s.name+"/"+p.name, func(t *testing.T) {
				raw, err := s.engine.Tree(harness.Case{Source: p.probe})
				if err != nil {
					record("parts", "%s\t%s\t%s\tERROR\t%v", s.name, p.name, p.probe, err)
					t.Logf("error: %v", err)
					return
				}
				label := rawLabel(raw)
				neutral, err := s.convert(raw)
				shown := neutral.String()
				if err != nil {
					shown = "NOT CONVERTIBLE: " + err.Error()
				}
				record("parts", "%s\t%s\t%s\t%s\t%s", s.name, p.name, p.probe, label, shown)
				t.Logf("raw=%s neutral=%s", label, shown)
				if s.name == "own" && err != nil {
					t.Errorf("own part %s has no neutral kind", p.name)
				}
			})
		}
	}
}

func TestRawKindsOfTheNineteen(t *testing.T) {
	for _, s := range subjects(t) {
		seen := map[string]bool{}
		rawTotal, neutralTotal := 0, 0
		for _, c := range harness.Cases() {
			raw, err := s.engine.Tree(c)
			if err != nil {
				t.Errorf("%s formula %d: %v", s.name, c.Number, err)
				continue
			}
			for _, kind := range rawKinds(raw) {
				seen[kind] = true
			}
			neutral, _ := s.convert(raw)
			rawTotal += rawCount(raw)
			neutralTotal += neutral.Count()
		}
		record("folds", "%s\t%d\t%d\t%d", s.name, rawTotal, neutralTotal, rawTotal-neutralTotal)
		kinds := make([]string, 0, len(seen))
		for kind := range seen {
			kinds = append(kinds, kind)
		}
		sort.Strings(kinds)
		record("kinds", "%s\t%d\t%s", s.name, len(kinds), strings.Join(kinds, ", "))
		t.Logf("%s raw node kinds (%d): %s", s.name, len(kinds), strings.Join(kinds, ", "))
	}
}

type reverseStep struct {
	engine      string
	translate   func(harness.Case) (string, bool)
	reverse     func(string) string
	translation string
}

func TestReverseStep(t *testing.T) {
	steps := []reverseStep{
		{"cel-go", cel.Translations, cel.Reverse, "decimal(\"..\"), name_v1"},
		{"expr-lang", exprlang.Translations, exprlang.Reverse, "dec(\"..\"), name_v1"},
	}
	extra := append(append([]standardCase{}, literalCases...), bracketCases...)
	for _, step := range steps {
		exact := 0
		for _, c := range harness.Cases() {
			text, _ := step.translate(c)
			if back := step.reverse(text); back == c.Source {
				exact++
			} else {
				t.Errorf("%s formula %d: reverse(translate(source)) = %q", step.engine, c.Number, back)
			}
		}
		lossy := []string{}
		for _, e := range extra {
			text, _ := step.translate(harness.Case{Source: e.source})
			if back := step.reverse(text); back != e.source {
				lossy = append(lossy, fmt.Sprintf("%s -> %s", e.source, back))
			}
		}
		record("reverse", "%s\t%d/19\t%s", step.engine, exact, strings.Join(lossy, "; "))
		t.Logf("%s reverse(translate(page text)) exact on %d of 19; extra probes changed: %v", step.engine, exact, lossy)
	}
}

func renderTable() string {
	var out strings.Builder
	writer := tabwriter.NewWriter(&out, 0, 0, 2, ' ', 0)
	fmt.Fprintf(writer, "Fit table: tree for the block view, and print-back in the standard shape\n\n")
	fmt.Fprintf(writer, "per formula (variants: %d respacings of each formula)\n", variantStyles)
	fmt.Fprintf(writer, "engine\tformula\ttree converts\ttree equals own\tidempotent\tcanonical\tround-trips\treads as own\tequals page text\tremarks\n")
	sort.SliceStable(outcomes, func(i, j int) bool {
		if outcomes[i].engine != outcomes[j].engine {
			return outcomes[i].engine < outcomes[j].engine
		}
		return outcomes[i].number < outcomes[j].number
	})
	for _, o := range outcomes {
		fmt.Fprintf(writer, "%s\t%d\t%s\t%s\t%s\t%d/%d\t%s\t%s\t%s\t%s\n", o.engine, o.number, yes(o.converts), yes(o.sameAsOwn),
			yes(o.idempotent), o.canonical, variantStyles, yes(o.roundTrips), yes(o.readsAsOwn), yes(o.equalsPage), strings.Join(o.remarks, "; "))
	}
	fmt.Fprintf(writer, "\ntotals per engine\n")
	fmt.Fprintf(writer, "engine\tconverts\ttree equals own\tidempotent\tcanonical formulas\tcanonical variants\tround-trips\treads as own\tequals page text\n")
	for _, engine := range engineNames() {
		fmt.Fprintf(writer, "%s\t%s\n", engine, totals(engine))
	}
	for _, block := range []struct{ title, header, key string }{
		{"reverse step (translate then reverse on the page text)", "engine\texact\tprobes that changed", "reverse"},
		{"raw nodes against neutral nodes in the 19 formulas", "engine\traw nodes\tneutral nodes\tfolded away", "folds"},
		{"raw node kinds met in the 19 formulas", "engine\tcount\tkinds", "kinds"},
		{"brackets (standard text per design page)", "engine\tsource\tprinted\tstandard text\ttree equals own\tround-trips\treads as own", "brackets"},
		{"literal fidelity", "engine\tsource\tprinted\tkept as standard\tround-trips\treads as own\tneutral tree", "literals"},
		{"language parts the examples need", "engine\tpart\tprobe\traw top node\tneutral node", "parts"},
	} {
		fmt.Fprintf(writer, "\n%s\n%s\n", block.title, block.header)
		lines := append([]string{}, sections[block.key]...)
		sort.Strings(lines)
		for _, line := range lines {
			fmt.Fprintln(writer, line)
		}
	}
	writer.Flush()
	return out.String()
}

func engineNames() []string {
	seen := map[string]bool{}
	var names []string
	for _, o := range outcomes {
		if !seen[o.engine] {
			seen[o.engine] = true
			names = append(names, o.engine)
		}
	}
	return names
}

func totals(engine string) string {
	var converts, same, idem, canonicalFormulas, canonicalVariants, trips, reads, page, count int
	for _, o := range outcomes {
		if o.engine != engine {
			continue
		}
		count++
		converts += b2i(o.converts)
		same += b2i(o.sameAsOwn)
		idem += b2i(o.idempotent)
		canonicalFormulas += b2i(o.canonical == variantStyles)
		canonicalVariants += o.canonical
		trips += b2i(o.roundTrips)
		reads += b2i(o.readsAsOwn)
		page += b2i(o.equalsPage)
	}
	return fmt.Sprintf("%d/%d\t%d/%d\t%d/%d\t%d/%d\t%d/%d\t%d/%d\t%d/%d\t%d/%d", converts, count, same, count, idem, count,
		canonicalFormulas, count, canonicalVariants, count*variantStyles, trips, count, reads, count, page, count)
}

func b2i(ok bool) int {
	if ok {
		return 1
	}
	return 0
}
