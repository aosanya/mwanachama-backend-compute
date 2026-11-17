package limits

import (
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
)

func TestProposedLimitsLetAllNineteenFormulasPass(t *testing.T) {
	for _, f := range factories {
		t.Run(f.name, func(t *testing.T) {
			harness.RunTest(t, f.build(Proposed(), true))
		})
	}
}
