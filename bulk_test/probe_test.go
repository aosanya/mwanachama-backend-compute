package bulkruns

import (
	"sync/atomic"
	"time"

	"github.com/aosanya/mwanachama-backend-compute/formula"
)

type probe struct {
	holdValue string
	waitFor   int64
	started   atomic.Int64
	finished  atomic.Int64
	released  atomic.Bool
}

func (p *probe) Function(library string, name string, version int) (formula.Signature, bool) {
	if library == "" && name == "probe" && version == 1 {
		return formula.Signature{
			Params: []formula.Param{{Name: "amount", Type: formula.TypeDecimal}},
			Result: formula.TypeDecimal,
			Cost:   1,
		}, true
	}
	return formula.Signature{}, false
}

func (p *probe) Versions(library string, name string) []int {
	if library == "" && name == "probe" {
		return []int{1}
	}
	return nil
}

func (p *probe) Apply(library string, name string, version int, args []formula.Value) (formula.Value, *formula.Problem) {
	p.started.Add(1)
	if p.holdValue != "" && args[0].Decimal().String() == p.holdValue {
		deadline := time.Now().Add(20 * time.Second)
		for p.finished.Load() < p.waitFor && time.Now().Before(deadline) {
			time.Sleep(time.Millisecond)
		}
		p.released.Store(p.finished.Load() >= p.waitFor)
	}
	p.finished.Add(1)
	return args[0], nil
}

const bodyProbed = "round(probe@1(balance) * rate, 2)"
