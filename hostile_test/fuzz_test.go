package hostile

import (
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/aosanya/mwanachama-backend-compute/formula"
)

var readSentence = regexp.MustCompile(`^The formula cannot be read: .+ at line ([1-9][0-9]*), column ([1-9][0-9]*)\.$`)

func seedTexts() []string {
	return []string{
		"payment >= balance ? 0.00 : balance - payment",
		"round(total / people, 2)",
		"sum(rows) + sum(rows)",
		"fin.pmt@1(total, 0.05, 12)",
		"step_e@1(total)",
		"people == 0 ? 0.00 : total / people",
		"!flag && people > 2",
		`label == "farm"`,
		"[1, 2, 3]",
		"price * quantity +",
		"sum(price, quantity",
		`"hello\`,
		"1 + \x00 2",
		"tot​al + 1",
		"\U0001F600 * 1",
		"\xff\xfe\xc0",
		strings.Repeat("(", 8),
		nested(13, "(", ")", "1"),
		nested(13, "[", "]", "1"),
		nested(12, "!(", ")", "flag"),
		strings.Repeat("a", 600) + " + 1",
		strings.Repeat("1", 19),
		"0." + strings.Repeat("1", 11),
		manyItems(99),
		balancedSum(32, "sum(rows)"),
		"now()",
		`read_file("/etc/passwd")`,
		"[1, 2, 3][99]",
		"__proto__",
	}
}

func outcomeOf(t *testing.T, text string) string {
	t.Helper()
	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("panic on %q: %v", text, recovered)
		}
	}()
	tree, problems := formula.Parse(text)
	for _, problem := range problems {
		requirePlain(t, problem)
	}
	if len(problems) > 0 {
		first := problems[0]
		if utf8.ValidString(text) && utf8.RuneCountInString(text) > 500 && first.Code != "too-long" {
			t.Fatalf("a text of %d characters gave %s", utf8.RuneCountInString(text), shown(first))
		}
		if first.Code == "bad-syntax" {
			match := readSentence.FindStringSubmatch(first.Sentence)
			if match == nil {
				t.Fatalf("a syntax refusal without line and column: %q", first.Sentence)
			}
			line, _ := strconv.Atoi(match[1])
			column, _ := strconv.Atoi(match[2])
			if line != first.Line || column != first.Column {
				t.Fatalf("sentence %q and position %d:%d differ", first.Sentence, first.Line, first.Column)
			}
		}
		return "parse " + shown(first)
	}
	problems = formula.Check(tree, standardInputs(), hostileLibrary())
	if len(problems) > 20 {
		t.Fatalf("Check listed %d problems", len(problems))
	}
	for _, problem := range problems {
		requirePlain(t, problem)
	}
	if len(problems) > 0 {
		return "check " + shown(problems[0])
	}
	if cost := formula.Cost(tree, hostileLibrary()); cost > 20000 {
		t.Fatalf("Check passed a formula of cost %d", cost)
	}
	program, problems := formula.Plan(tree, standardInputs(), hostileLibrary(), formula.DefaultOptions())
	if len(problems) > 0 {
		t.Fatalf("Plan refused what Check passed: %+v", problems)
	}
	meter := formula.NewMeter(20000)
	value, problem := program.RunMetered(standardValues(200), meter)
	if problem != nil {
		requirePlain(t, *problem)
		return "run " + shown(*problem)
	}
	if meter.Spent() > meter.Limit() {
		t.Fatalf("a run that gave a value spent %d of %d", meter.Spent(), meter.Limit())
	}
	return "value " + string(value.Type()) + " " + value.String()
}

func holdsEveryProperty(t *testing.T, text string) {
	t.Helper()
	first := outcomeOf(t, text)
	if second := outcomeOf(t, text); first != second {
		t.Fatalf("two runs of %.80q differ: %q and %q", text, first, second)
	}
}

func FuzzParse(f *testing.F) {
	for _, seed := range seedTexts() {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, text string) {
		holdsEveryProperty(t, text)
	})
}

var pieces = []string{
	"(", ")", "[", "]", ",", "+", "-", "*", "/", "%", "!", "&&", "||", "==", "!=", "<", "<=", ">", ">=", "?", ":", "@", ".",
	"total", "people", "flag", "rows", "label", "sum", "count", "avg", "min", "max", "round", "to_float", "to_decimal",
	"days_between", "fin.pmt@1", "step_e@1", "0", "1", "2.50", "99999999999999999999", "0.12345678901", `"abc"`, `"`, `\`,
	" ", "\t", "\n", "\r", "\x00", "\a", "​", "‮", "\U0001F600", "e", "__proto__", "true", "false",
}

func hostileText(random *rand.Rand) string {
	var text strings.Builder
	count := 1 + random.Intn(60)
	for i := 0; i < count; i++ {
		switch random.Intn(12) {
		case 0:
			text.WriteString(strings.Repeat("(", 1+random.Intn(30)))
		case 1:
			text.WriteString(strings.Repeat("a", 1+random.Intn(700)))
		case 2:
			text.WriteString(strings.Repeat("1", 1+random.Intn(40)))
		default:
			text.WriteString(pieces[random.Intn(len(pieces))])
		}
	}
	return text.String()
}

func seedOfTheDay(t *testing.T) int64 {
	t.Helper()
	if given := os.Getenv("HOSTILE_SEED"); given != "" {
		seed, err := strconv.ParseInt(given, 10, 64)
		if err != nil {
			t.Fatalf("HOSTILE_SEED %q: %v", given, err)
		}
		return seed
	}
	return 20261117
}

func TestGeneratedHostileTextHoldsEveryProperty(t *testing.T) {
	seed := seedOfTheDay(t)
	t.Logf("seed %d; set HOSTILE_SEED to replay or to try another", seed)
	random := rand.New(rand.NewSource(seed))
	current := ""
	defer func() {
		if t.Failed() {
			t.Logf("seed %d, text %q", seed, current)
		}
	}()
	for i := 0; i < 3000; i++ {
		current = hostileText(random)
		holdsEveryProperty(t, current)
	}
}
