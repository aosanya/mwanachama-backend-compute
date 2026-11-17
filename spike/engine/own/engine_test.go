package own

import (
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
)

func TestOwnEnginePassesEveryFormula(t *testing.T) {
	harness.RunTest(t, New())
}

func TestOwnEngineReportsAllNineteenPassing(t *testing.T) {
	report := harness.NewReport(New())
	if !report.AllPassed() || report.PassCount() != 19 || report.Engine != "own" {
		t.Errorf("report is:\n%s", report)
	}
}
