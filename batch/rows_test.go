package batch

import (
	"context"
	"reflect"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/function"
	"github.com/aosanya/mwanachama-backend-compute/problem"
)

func TestAllGoodRowsGiveOneValueEachInTheOrderSent(t *testing.T) {
	compiled := mustCompile(t, evenShare(), evenShareBody)
	rows := []function.Row{
		row("total", "100.00", "people", "3"),
		row("total", "90.00", "people", "4"),
		row("total", "50.00", "people", "2"),
	}
	result, err := compiled.RunBatch(context.Background(), rows, Options{})
	requireNoError(t, err)
	requireRowNumbers(t, result, 3)
	if got, want := answersOf(result), []string{"33.33", "22.50", "25.00"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("answers %v, want %v", got, want)
	}
	requireSummaryMatches(t, result)
	requireEqualsSingle(t, compiled, rows, nil, result)
}

func TestOneBadRowNeverSpoilsTheGoodRowsAroundIt(t *testing.T) {
	compiled := mustCompile(t, loanBalance(), loanBalanceBody)
	rows := []function.Row{
		row("balance", "120.50", "payment", "45.25"),
		row("balance", "80.00", "payment", "-5.00"),
		row("balance", "60.00", "payment", "70.00"),
	}
	result, err := compiled.RunBatch(context.Background(), rows, Options{})
	requireNoError(t, err)
	requireRowNumbers(t, result, 3)
	requireValue(t, result.Rows[0], "75.25")
	requireProblems(t, result.Rows[1], StageValidate, problem.CodeNoNegatives)
	requireValue(t, result.Rows[2], "0.00")
	requireEqualsSingle(t, compiled, rows, nil, result)
}

func TestMonthEndRunHasFourProblemsAcrossThreeStoppedRows(t *testing.T) {
	compiled := mustCompile(t, loanBalance(), loanBalanceBody)
	rows := []function.Row{
		row("balance", "120.50", "payment", "45.25"),
		row("balance", "80.00", "payment", "-5.00"),
		row("balance", "60.00", "payment", "70.00"),
		row("balance", "200.00"),
		row("balance", "-1.00", "payment", "5.125"),
	}
	result, err := compiled.RunBatch(context.Background(), rows, Options{})
	requireNoError(t, err)
	requireRowNumbers(t, result, 5)
	requireValue(t, result.Rows[0], "75.25")
	requireProblems(t, result.Rows[1], StageValidate, problem.CodeNoNegatives)
	requireValue(t, result.Rows[2], "0.00")
	requireProblems(t, result.Rows[3], StageValidate, problem.CodeRequired)
	requireProblems(t, result.Rows[4], StageValidate, problem.CodeNoNegatives, problem.CodeDecimalPlaces)
	summary := result.Summary
	if summary.Rows != 5 || summary.Valued != 2 || summary.Stopped != 3 || summary.NotRun != 0 {
		t.Fatalf("summary %+v, want 5 rows, 2 valued, 3 stopped, 0 not run", summary)
	}
	want := map[string]int{problem.CodeNoNegatives: 2, problem.CodeRequired: 1, problem.CodeDecimalPlaces: 1}
	if !reflect.DeepEqual(summary.ProblemsByCode, want) {
		t.Fatalf("problems by code %v, want %v", summary.ProblemsByCode, want)
	}
	requireSummaryMatches(t, result)
	requireEqualsSingle(t, compiled, rows, nil, result)
}

func TestARowThatFailsWhileCalculatedStopsOnlyItself(t *testing.T) {
	compiled := mustCompile(t, evenShare(), evenShareBody)
	rows := []function.Row{
		row("total", "100.00", "people", "3"),
		row("total", "10.00", "people", "0"),
		row("total", "90.00", "people", "4"),
	}
	result, err := compiled.RunBatch(context.Background(), rows, Options{})
	requireNoError(t, err)
	requireValue(t, result.Rows[0], "33.33")
	requireProblems(t, result.Rows[1], StageCalculate, "divide-by-zero")
	requireValue(t, result.Rows[2], "22.50")
	requireEqualsSingle(t, compiled, rows, nil, result)
}

func TestARowWithProblemsIsNeverCalculated(t *testing.T) {
	compiled := mustCompile(t, evenShare(), evenShareBody)
	rows := []function.Row{row("total", "10.00", "people", "0"), row("total", "-1.00", "people", "0")}
	result, err := compiled.RunBatch(context.Background(), rows, Options{})
	requireNoError(t, err)
	requireProblems(t, result.Rows[0], StageCalculate, "divide-by-zero")
	requireProblems(t, result.Rows[1], StageValidate, problem.CodeNoNegatives)
}

func TestARowHoldsAValueOrProblemsNeverBothAndNeverNeither(t *testing.T) {
	compiled := mustCompile(t, loanBalance(), loanBalanceBody)
	rows := []function.Row{
		row("balance", "120.50", "payment", "45.25"),
		row("balance", "80.00", "payment", "-5.00"),
		row("payment", "5.00"),
		{},
	}
	result, err := compiled.RunBatch(context.Background(), rows, Options{})
	requireNoError(t, err)
	for _, found := range result.Rows {
		hasValue := found.Value.String() != ""
		hasProblems := len(found.Problems) > 0
		if hasValue == hasProblems {
			t.Fatalf("row %d holds value %v and problems %v", found.Row, hasValue, hasProblems)
		}
		if hasProblems && found.StoppedAt == "" {
			t.Fatalf("row %d has problems and no stage", found.Row)
		}
	}
}

func TestEveryProblemOfARowComesBackAtOnce(t *testing.T) {
	compiled := mustCompile(t, loanBalance(), loanBalanceBody)
	rows := []function.Row{row("balance", "-1.00", "payment", "5.125", "extra", "1")}
	result, err := compiled.RunBatch(context.Background(), rows, Options{})
	requireNoError(t, err)
	requireProblems(t, result.Rows[0], StageValidate, problem.CodeNoNegatives, problem.CodeDecimalPlaces, problem.CodeUnknownInput)
}

func TestAnEmptyBatchGivesNoResultsNoErrorAndASummaryOfZeros(t *testing.T) {
	compiled := mustCompile(t, monthEnd(), monthEndBody)
	result, err := compiled.RunBatch(context.Background(), nil, Options{})
	requireNoError(t, err)
	if result.Rows == nil || len(result.Rows) != 0 {
		t.Fatalf("rows %v, want an empty list", result.Rows)
	}
	summary := result.Summary
	if summary.Rows != 0 || summary.Valued != 0 || summary.Stopped != 0 || summary.NotRun != 0 || len(summary.ProblemsByCode) != 0 {
		t.Fatalf("summary %+v, want zeros", summary)
	}
	if summary.ChunkSize != DefaultChunkSize || summary.Workers != DefaultWorkers {
		t.Fatalf("summary chunk size %d workers %d, want the defaults", summary.ChunkSize, summary.Workers)
	}
}

func TestSizesAroundAChunkAllGiveEveryRowItsOwnAnswer(t *testing.T) {
	compiled := mustCompile(t, monthEnd(), monthEndBody)
	for _, count := range []int{1, 3, 99, 100, 101, 200, 201} {
		rows := monthEndRows(count)
		result, err := compiled.RunBatch(context.Background(), rows, Options{ChunkSize: 100})
		requireNoError(t, err)
		requireRowNumbers(t, result, count)
		for i, found := range result.Rows {
			requireValue(t, found, monthEndAnswer(i+1))
		}
		if result.Summary.Valued != count || result.Summary.Stopped != 0 || result.Summary.NotRun != 0 {
			t.Fatalf("%d rows: summary %+v", count, result.Summary)
		}
	}
}

func TestTheMaximumNumberOfRowsIsAccepted(t *testing.T) {
	compiled := mustCompile(t, monthEnd(), monthEndBody)
	rows := monthEndRows(10000)
	result, err := compiled.RunBatch(context.Background(), rows, Options{Workers: 4})
	requireNoError(t, err)
	requireRowNumbers(t, result, 10000)
	requireValue(t, result.Rows[0], "0.05")
	requireValue(t, result.Rows[9999], "500.00")
	requireSummaryMatches(t, result)
}

func TestChunkEdgesGiveTheSameAnswersWhateverTheChunkSizeAndWorkers(t *testing.T) {
	compiled := mustCompile(t, monthEnd(), monthEndBody)
	rows := monthEndRows(7)
	want := []string{"0.05", "0.10", "0.15", "0.20", "0.25", "0.30", "0.35"}
	for _, chunk := range []int{1, 2, 3, 7, 8} {
		for _, workers := range []int{1, 4} {
			result, err := compiled.RunBatch(context.Background(), rows, Options{ChunkSize: chunk, Workers: workers})
			requireNoError(t, err)
			if got := answersOf(result); !reflect.DeepEqual(got, want) {
				t.Fatalf("chunk %d workers %d gave %v, want %v", chunk, workers, got, want)
			}
		}
	}
}

func TestDefaultsAreReportedInTheSummary(t *testing.T) {
	compiled := mustCompile(t, monthEnd(), monthEndBody)
	result, err := compiled.RunBatch(context.Background(), monthEndRows(250), Options{})
	requireNoError(t, err)
	if result.Summary.ChunkSize != 100 || result.Summary.Workers != 1 {
		t.Fatalf("summary chunk size %d workers %d, want 100 and 1", result.Summary.ChunkSize, result.Summary.Workers)
	}
	explicit, err := compiled.RunBatch(context.Background(), monthEndRows(250), Options{ChunkSize: 100, Workers: 1})
	requireNoError(t, err)
	if !reflect.DeepEqual(answersOf(result), answersOf(explicit)) {
		t.Fatalf("the default run differs from the explicit run")
	}
}

func TestSharedInputsApplyToEveryRowAndTheRowsOwnValueWins(t *testing.T) {
	compiled := mustCompile(t, loanBalance(), loanBalanceBody)
	rows := []function.Row{
		row("balance", "120.50"),
		row("balance", "60.00", "payment", "99.00"),
		row("balance", "30.00"),
	}
	shared := row("payment", "10.00")
	result, err := compiled.RunBatch(context.Background(), rows, Options{Shared: shared})
	requireNoError(t, err)
	requireValue(t, result.Rows[0], "110.50")
	requireValue(t, result.Rows[1], "0.00")
	requireValue(t, result.Rows[2], "20.00")
	requireEqualsSingle(t, compiled, rows, shared, result)
}

func TestMergeLeavesBothMapsAsTheyWere(t *testing.T) {
	shared := row("payment", "10.00")
	own := row("balance", "5.00", "payment", "1.00")
	merged := Merge(shared, own)
	if merged["payment"].Text != "1.00" || merged["balance"].Text != "5.00" {
		t.Fatalf("merged %v", merged)
	}
	if shared["payment"].Text != "10.00" || len(shared) != 1 || len(own) != 2 {
		t.Fatalf("Merge changed an argument: shared %v own %v", shared, own)
	}
}

func TestATotalPassedInGivesTheSameAnswersAsTheFormulaGivenTheList(t *testing.T) {
	passedIn := mustCompile(t, fundShare(), fundShareBody)
	fromList := mustCompile(t, fundShareFromList(), fundShareFromListBody)
	contributions := []string{"250.00", "500.00", "250.00"}
	rows := make([]function.Row, len(contributions))
	for i, contribution := range contributions {
		rows[i] = row("contribution", contribution)
	}
	shared := row("total_contributions", "1000.00", "payout", "400.00")
	result, err := passedIn.RunBatch(context.Background(), rows, Options{Shared: shared})
	requireNoError(t, err)
	if got, want := answersOf(result), []string{"100.00", "200.00", "100.00"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("passed-in total gave %v, want %v", got, want)
	}
	list := function.Row{"contributions": listOf(contributions...), "payout": row("payout", "400.00")["payout"]}
	listRows := make([]function.Row, len(contributions))
	for i, contribution := range contributions {
		listRows[i] = row("contribution", contribution)
	}
	listResult, err := fromList.RunBatch(context.Background(), listRows, Options{Shared: list})
	requireNoError(t, err)
	if !reflect.DeepEqual(answersOf(result), answersOf(listResult)) {
		t.Fatalf("passed-in total %v differs from the list %v", answersOf(result), answersOf(listResult))
	}
	for i := range result.Rows {
		if result.Rows[i].Value.Type() != listResult.Rows[i].Value.Type() {
			t.Fatalf("row %d type differs: %s against %s", i+1, result.Rows[i].Value.Type(), listResult.Rows[i].Value.Type())
		}
	}
}

func TestAListThatIsNotGivenIsAnEmptyListForTheFormula(t *testing.T) {
	definition := function.Definition{Inputs: []function.InputDecl{
		{Name: "amounts", Type: formula.TypeListDecimal, Rules: nil},
	}}
	compiled := mustCompile(t, definition, "sum(amounts)")
	result, err := compiled.RunBatch(context.Background(), []function.Row{{}}, Options{})
	requireNoError(t, err)
	requireValue(t, result.Rows[0], "0")
}

func TestTheSummaryNamesTheFormulaThatWasRun(t *testing.T) {
	compiled := mustCompile(t, monthEnd(), monthEndBody)
	same := mustCompile(t, monthEnd(), "round(  balance*rate ,2)")
	other := mustCompile(t, monthEnd(), "round(balance * rate, 3)")
	first, err := compiled.RunBatch(context.Background(), monthEndRows(2), Options{})
	requireNoError(t, err)
	respaced, err := same.RunBatch(context.Background(), monthEndRows(2), Options{})
	requireNoError(t, err)
	different, err := other.RunBatch(context.Background(), monthEndRows(2), Options{})
	requireNoError(t, err)
	if first.Summary.FormulaHash == "" {
		t.Fatal("the summary has no formula hash")
	}
	if first.Summary.FormulaHash != respaced.Summary.FormulaHash {
		t.Fatalf("a respaced formula changed the hash: %s against %s", first.Summary.FormulaHash, respaced.Summary.FormulaHash)
	}
	if first.Summary.FormulaHash == different.Summary.FormulaHash {
		t.Fatal("two different formulas have the same hash")
	}
}

func TestCompileRefusesABadDefinitionAndABadFormula(t *testing.T) {
	if _, problems := Compile(monthEnd(), "round(balance * missing, 2)", nil, defaultOptions()); len(problems) == 0 {
		t.Fatal("a formula that reads an unknown name was accepted")
	}
	if _, problems := Compile(monthEnd(), "round(balance *", nil, defaultOptions()); len(problems) == 0 {
		t.Fatal("a formula that cannot be read was accepted")
	}
	bad := function.Definition{Inputs: []function.InputDecl{integerInput("people"), integerInput("people")}}
	if _, problems := Compile(bad, "people", nil, defaultOptions()); len(problems) == 0 {
		t.Fatal("a definition with a repeated input was accepted")
	}
}

func TestACheckOfTheDefinitionRunsOnEveryRow(t *testing.T) {
	definition := monthEnd()
	definition.Checks = []function.CheckDecl{{Formula: "rate <= 0.1000", Message: "The rate cannot be over ten percent"}}
	compiled := mustCompile(t, definition, monthEndBody)
	rows := []function.Row{
		row("balance", "100.00", "rate", "0.0500"),
		row("balance", "100.00", "rate", "0.2500"),
		row("balance", "100.00", "rate", "0.1000"),
	}
	result, err := compiled.RunBatch(context.Background(), rows, Options{Workers: 2, ChunkSize: 1})
	requireNoError(t, err)
	requireValue(t, result.Rows[0], "5.00")
	requireProblems(t, result.Rows[1], StageValidate, problem.CodeCheckFailed)
	requireValue(t, result.Rows[2], "10.00")
	requireEqualsSingle(t, compiled, rows, nil, result)
}

func TestARowOverItsOwnCostLimitStopsAtCalculateAndTheRestCarryOn(t *testing.T) {
	definition := function.Definition{Inputs: []function.InputDecl{
		{Name: "amounts", Type: formula.TypeListDecimal, Rules: nil},
	}}
	options := formula.DefaultOptions()
	options.Limits.Cost = 500
	compiled, problems := Compile(definition, "sum(amounts)", nil, options)
	if len(problems) > 0 {
		t.Fatalf("Compile refused: %v", problems)
	}
	big := make([]string, 1000)
	for i := range big {
		big[i] = "1.00"
	}
	small := function.Row{"amounts": listOf("1.00", "2.00")}
	rows := []function.Row{small, {"amounts": listOf(big...)}, small, {"amounts": listOf(big...)}, small}
	result, err := compiled.RunBatch(context.Background(), rows, Options{ChunkSize: 2, Workers: 2})
	requireNoError(t, err)
	requireValue(t, result.Rows[0], "3.00")
	requireProblems(t, result.Rows[1], StageCalculate, "too-costly")
	requireValue(t, result.Rows[2], "3.00")
	requireProblems(t, result.Rows[3], StageCalculate, "too-costly")
	requireValue(t, result.Rows[4], "3.00")
	if result.Rows[1].Problems[0].Sentence != "This formula goes over the cost limit of 500." {
		t.Fatalf("sentence %q", result.Rows[1].Problems[0].Sentence)
	}
	requireEqualsSingle(t, compiled, rows, nil, result)
}

func TestAListOfMoreThanAThousandItemsStopsAtValidateAndNothingElse(t *testing.T) {
	definition := function.Definition{Inputs: []function.InputDecl{
		{Name: "amounts", Type: formula.TypeListDecimal, Rules: nil},
	}}
	compiled := mustCompile(t, definition, "sum(amounts)")
	long := make([]string, 1001)
	for i := range long {
		long[i] = "1.00"
	}
	rows := []function.Row{{"amounts": listOf("1.00")}, {"amounts": listOf(long...)}, {"amounts": listOf("2.00")}}
	result, err := compiled.RunBatch(context.Background(), rows, Options{})
	requireNoError(t, err)
	requireValue(t, result.Rows[0], "1.00")
	requireProblems(t, result.Rows[1], StageValidate, problem.CodeListTooLong)
	requireValue(t, result.Rows[2], "2.00")
}
