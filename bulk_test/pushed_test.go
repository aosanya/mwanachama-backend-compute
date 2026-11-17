package bulkruns

import (
	"context"
	"math/big"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/batch"
	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/function"
	"github.com/aosanya/mwanachama-backend-compute/number"
	"github.com/aosanya/mwanachama-backend-compute/numtext"
	"github.com/aosanya/mwanachama-backend-compute/problem"
)

type standInDatabase struct {
	rats []*big.Rat
}

func newStandIn(t *testing.T, contributions ...string) standInDatabase {
	t.Helper()
	database := standInDatabase{}
	for _, text := range contributions {
		parsed, ok := new(big.Rat).SetString(text)
		require(t, ok, "%q is not a number", text)
		database.rats = append(database.rats, parsed)
	}
	return database
}

func (d standInDatabase) sum() *big.Rat {
	total := new(big.Rat)
	for _, each := range d.rats {
		total.Add(total, each)
	}
	return total
}

func (d standInDatabase) extreme(wantLarger bool) *big.Rat {
	best := d.rats[0]
	for _, each := range d.rats[1:] {
		if (each.Cmp(best) > 0) == wantLarger && each.Cmp(best) != 0 {
			best = each
		}
	}
	return best
}

func (d standInDatabase) count() int {
	return len(d.rats)
}

func shareRows(contributions ...string) []function.Row {
	rows := make([]function.Row, len(contributions))
	for i, text := range contributions {
		rows[i] = row("contribution", text)
	}
	return rows
}

func sharedTotals(payout string, total string) function.Row {
	return row("payout", payout, "total_contributions", total)
}

func leftOver(t *testing.T, payout string, rows []batch.RowResult) string {
	t.Helper()
	remaining := decimalOf(t, payout)
	for _, found := range rows {
		require(t, len(found.Problems) == 0, "row %d has problems", found.Row)
		remaining = remaining.Sub(found.Value.Decimal())
	}
	return remaining.String()
}

func requireDecimalText(t *testing.T, got string, want string) {
	t.Helper()
	require(t, decimalOf(t, got).Equal(decimalOf(t, want)), "got %s, want %s", got, want)
}

func TestBR50ASplitThatDividesEvenlyLeavesNothing(t *testing.T) {
	compiled := compileF(t, formula.DefaultOptions())
	rows := shareRows("250.00", "500.00", "250.00")
	shared := sharedTotals("400.00", "1000.00")
	result := runAndCheck(t, compiled, rows, shared, batch.Options{}, []answer{value("100.00"), value("200.00"), value("100.00")})
	requireSum(t, result.Rows, 0, 3, "400.00")
	requireDecimalText(t, leftOver(t, "400.00", result.Rows), "0.00")
}

func TestBR51ALeftOverCentIsInNoRow(t *testing.T) {
	compiled := compileF(t, formula.DefaultOptions())
	rows := shareRows("100.00", "100.00", "100.00")
	shared := sharedTotals("100.00", "300.00")
	result := runAndCheck(t, compiled, rows, shared, batch.Options{}, []answer{value("33.33"), value("33.33"), value("33.33")})
	requireSum(t, result.Rows, 0, 3, "99.99")
	requireDecimalText(t, leftOver(t, "100.00", result.Rows), "0.01")
}

func TestBR52OneThousandEqualContributions(t *testing.T) {
	compiled := compileF(t, formula.DefaultOptions())
	contributions := make([]string, 1000)
	want := make([]answer, 1000)
	for i := range contributions {
		contributions[i] = "10.00"
		want[i] = value("5.00")
	}
	result := runAndCheck(t, compiled, shareRows(contributions...), sharedTotals("5000.00", "10000.00"), batch.Options{ChunkSize: 100}, want)
	requireSum(t, result.Rows, 0, 1000, "5000.00")
	requireDecimalText(t, leftOver(t, "5000.00", result.Rows), "0.00")
}

func listDefinition(name string) function.Definition {
	return function.Definition{Inputs: []function.InputDecl{
		{Name: name, Type: formula.TypeListDecimal, Rules: []function.Rule{{Name: problem.RuleRequired}}},
	}}
}

func runOverList(t *testing.T, body string, contributions []string) batch.RowResult {
	t.Helper()
	compiled := compile(t, listDefinition("contributions"), body, nil, formula.DefaultOptions())
	rows := []function.Row{{"contributions": numtext.Many(contributions...)}}
	result, err := compiled.RunBatch(context.Background(), rows, batch.Options{})
	requireNoError(t, err)
	requireShape(t, result, 1)
	require(t, len(result.Rows[0].Problems) == 0, "%s gave problems %v", body, codesOf(result.Rows[0].Problems))
	return result.Rows[0]
}

func TestBR53TheAggregatesMatchTheRowsAndTheStandInDatabase(t *testing.T) {
	contributions := []string{"250.00", "500.00", "250.00", "125.50"}
	database := newStandIn(t, contributions...)
	sum := runOverList(t, "sum(contributions)", contributions)
	require(t, sum.Value.String() == database.sum().FloatString(2), "sum is %s, the stand-in database says %s", sum.Value.String(), database.sum().FloatString(2))
	require(t, sum.Value.String() == "1125.50", "sum is %s, want 1125.50", sum.Value.String())
	count := runOverList(t, "count(contributions)", contributions)
	require(t, count.Value.String() == "4" && database.count() == 4, "count is %s, the stand-in database says %d", count.Value.String(), database.count())
	smallest := runOverList(t, "min(contributions)", contributions)
	require(t, smallest.Value.String() == database.extreme(false).FloatString(2) && smallest.Value.String() == "125.50",
		"min is %s, the stand-in database says %s", smallest.Value.String(), database.extreme(false).FloatString(2))
	largest := runOverList(t, "max(contributions)", contributions)
	require(t, largest.Value.String() == database.extreme(true).FloatString(2) && largest.Value.String() == "500.00",
		"max is %s, the stand-in database says %s", largest.Value.String(), database.extreme(true).FloatString(2))
}

func TestBR53PassingTheSumInGivesTheSameAnswersAsGivingTheList(t *testing.T) {
	contributions := []string{"250.00", "500.00", "250.00", "125.50"}
	pushed := compileF(t, formula.DefaultOptions())
	fromTotal, err := pushed.RunBatch(context.Background(), shareRows(contributions...), batch.Options{Shared: sharedTotals("400.00", "1125.50")})
	requireNoError(t, err)
	listed := compile(t, function.Definition{Inputs: []function.InputDecl{
		decimalInput("contribution", 2),
		{Name: "contributions", Type: formula.TypeListDecimal, Rules: []function.Rule{{Name: problem.RuleRequired}}},
		decimalInput("payout", 2),
	}}, "round(payout * contribution / sum(contributions), 2)", nil, formula.DefaultOptions())
	fromList, err := listed.RunBatch(context.Background(), shareRows(contributions...), batch.Options{
		Shared: function.Row{"contributions": numtext.Many(contributions...), "payout": numtext.One("400.00")},
	})
	requireNoError(t, err)
	requireShape(t, fromTotal, 4)
	requireShape(t, fromList, 4)
	for i := range fromTotal.Rows {
		requireSameRow(t, "the total passed in against the list given", fromTotal.Rows[i], fromList.Rows[i])
	}
}

func TestBR54ThousandContributionsOfOneCentSumToTenExactly(t *testing.T) {
	contributions := make([]string, 1000)
	for i := range contributions {
		contributions[i] = "0.01"
	}
	total := runOverList(t, "sum(contributions)", contributions)
	require(t, total.Value.String() == "10.00", "the sum is %s, want exactly 10.00", total.Value.String())
	compiled := compileF(t, formula.DefaultOptions())
	want := make([]answer, 1000)
	for i := range want {
		want[i] = value("0.01")
	}
	result := runAndCheck(t, compiled, shareRows(contributions...), sharedTotals("10.00", total.Value.String()), batch.Options{ChunkSize: 100}, want)
	requireSum(t, result.Rows, 0, 1000, "10.00")
}

func TestBR55AnAverageAsAPushedSumAndAPushedCount(t *testing.T) {
	definition := function.Definition{Inputs: []function.InputDecl{
		decimalInput("contribution", 2), decimalInput("total_sum", 2), integerInput("member_count"),
	}}
	compiled := compile(t, definition, "round(contribution - total_sum / member_count, 2)", nil, formula.DefaultOptions())
	shared := row("total_sum", "900.00", "member_count", "3")
	runAndCheck(t, compiled, shareRows("200.00", "300.00", "400.00"), shared, batch.Options{}, []answer{value("-100.00"), value("0.00"), value("100.00")})
}

func TestBR56TheLibraryCannotCheckThatTheTotalCoversTheSameRows(t *testing.T) {
	compiled := compileF(t, formula.DefaultOptions())
	rows := shareRows("250.00", "500.00")
	result := runAndCheck(t, compiled, rows, sharedTotals("400.00", "1000.00"), batch.Options{}, []answer{value("100.00"), value("200.00")})
	rowsSum := newStandIn(t, "250.00", "500.00").sum()
	require(t, rowsSum.FloatString(2) == "750.00", "the test's own sum of the rows is %s, want 750.00", rowsSum.FloatString(2))
	require(t, rowsSum.Cmp(new(big.Rat).SetInt64(1000)) != 0, "the rows sum to the total, so the case proves nothing")
	requireSummaryCounts(t, result, 2, 0, 0)
}

func TestBR57ASplitThatDoesNotDivideEvenlyIsCutDownToTheSmallestUnit(t *testing.T) {
	options := formula.DefaultOptions()
	options.Numbers = number.Settings{Places: 2, Mode: number.RoundDown, DivisionDigits: 20}
	compiled := compileF(t, options)
	contributions := make([]string, 7)
	want := make([]answer, 7)
	for i := range contributions {
		contributions[i] = "100.00"
		want[i] = value("14.28")
	}
	result := runAndCheck(t, compiled, shareRows(contributions...), sharedTotals("100.00", "700.00"), batch.Options{}, want)
	requireSum(t, result.Rows, 0, 7, "99.96")
	requireDecimalText(t, leftOver(t, "100.00", result.Rows), "0.04")
	sum := decimalOf(t, sumOf(t, result.Rows, 0, 7))
	require(t, sum.Cmp(decimalOf(t, "100.00")) <= 0, "the shares add to %s, more than the payout", sum.String())
}
