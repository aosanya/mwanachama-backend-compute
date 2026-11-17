package bulkruns

import "testing"

func hold(t *testing.T, ids string, question string) {
	t.Helper()
	t.Skipf("%s held: %s", ids, question)
}

func TestBR72ToBR77HeldCancelAndDeadline(t *testing.T) {
	hold(t, "BR-72 to BR-77", "whether and how a host's own cancel or deadline is honoured is open (decision 19; open questions 1 and 8 of bulk-runs.md); the cases follow the proposal and wait for the owner")
}

func TestBR80HeldRowsInWorkAtOnce(t *testing.T) {
	hold(t, "BR-80", "counting the rows a worker holds needs a test-only hook compiled into batch/ (open question 7); the hook is not reachable from outside the package")
}

func TestBR81HeldMemoryOfTenThousandBadRows(t *testing.T) {
	hold(t, "BR-81", "the 32 MiB figure waits for the engine choice figures (open question 10)")
}

func TestBR82HeldHeapAndGoroutinesAfterBR80AndBR81(t *testing.T) {
	hold(t, "BR-82", "depends on BR-80 and BR-81, and the 1 MiB drift is a proposed figure (open question 10)")
}

func TestBR83HeldOneHundredCallsOfTenThousandRows(t *testing.T) {
	hold(t, "BR-83", "the 1 MiB drift over 100 calls is a proposed figure (open question 10)")
}

func TestBR84HeldHeapAfterARefusedCall(t *testing.T) {
	hold(t, "BR-84", "the 1 MiB drift is a proposed figure (open question 10)")
}

func TestBR85HeldGoroutinesAndHeapAfterACancelledCall(t *testing.T) {
	hold(t, "BR-85", "follows BR-73, which waits for the cancel decision (open questions 1 and 8)")
}

func TestHeldABatchSizeLimitBelowTheAgencysCeiling(t *testing.T) {
	hold(t, "the smaller agency setting case of section 7", "the ceiling above which no agency may set its rows per batch is open (open question 11)")
}

func TestHeldPushedDownAggregatesDeclaredByTheProgram(t *testing.T) {
	hold(t, "Aggregates() of the program (section 6, G09 on batch-equals-single.md)", "the language has no declaration syntax for a pushed-down input, so batch/ has no Aggregates method; BR-50 to BR-57 run with the total as a shared input")
}
