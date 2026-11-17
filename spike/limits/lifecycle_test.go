package limits

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
)

type contextual interface {
	RunContext(context.Context, map[string]any) (any, error)
}

func compileRun(engine harness.Engine, source string, params []harness.Param, inputs map[string]any, result harness.Kind) (any, error) {
	program, err := engine.Compile(harness.Case{Source: source, Params: params, Result: result}, nil)
	if err != nil {
		return nil, err
	}
	return program.Run(inputs)
}

func limitOf(err error) string {
	var failure *Failure
	if errors.As(err, &failure) {
		return failure.Limit
	}
	return "none"
}

func slowLimits() Limits {
	return scenarioNamed("L06").limits(Proposed())
}

func heavyFormula() string {
	return scenarioNamed("L06").formula(Proposed())
}

func hostileRound(t *testing.T, engine harness.Engine, slow harness.Engine) {
	t.Helper()
	for _, id := range []string{"L01", "L03", "L05", "C01", "P04", "I01"} {
		s := scenarioNamed(id)
		_, err := compileRun(engine, s.formula(Proposed()), s.params, s.inputs(Proposed()), s.result)
		if err == nil {
			t.Errorf("%s was not refused", id)
		}
	}
	if _, err := compileRun(slow, heavyFormula(), nil, nil, harness.KindInteger); limitOf(err) != LimitTime {
		t.Errorf("heavy formula: %v", err)
	}
}

func TestL07GoodFormulaStillWorksAfterHostileOnes(t *testing.T) {
	for _, f := range factories {
		engine, slow := f.build(Proposed(), true), f.build(slowLimits(), true)
		hostileRound(t, engine, slow)
		got, err := compileRun(engine, "2 + 3", nil, nil, harness.KindInteger)
		row := tableRow{Case: "L07", Engine: f.name, Mode: "full", Layer: string(LayerNone), Time: "-", Sentence: fmt.Sprintf("2 + 3 returned %v after six hostile formulas and a time-limit run", got)}
		record(row)
		if err != nil || got != int64(5) {
			t.Errorf("%s: 2 + 3 gave %v, %v", f.name, got, err)
		}
	}
}

func settle(baseline int, wait time.Duration) (int, time.Duration) {
	started := time.Now()
	for {
		if now := runtime.NumGoroutine(); now <= baseline || time.Since(started) > wait {
			return runtime.NumGoroutine(), time.Since(started)
		}
		time.Sleep(2 * time.Millisecond)
	}
}

func TestL08GoroutinesAndHeapAreBackAfterCostAndTimeRefusals(t *testing.T) {
	for _, f := range factories {
		engine, slow := f.build(Proposed(), true), f.build(slowLimits(), true)
		runtime.GC()
		time.Sleep(20 * time.Millisecond)
		goroutines := runtime.NumGoroutine()
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		cost := scenarioNamed("L05")
		for range 5 {
			_, err := compileRun(engine, cost.formula(Proposed()), nil, nil, harness.KindInteger)
			if limitOf(err) != LimitCost {
				t.Errorf("%s cost run: %v", f.name, err)
			}
			_, err = compileRun(slow, heavyFormula(), nil, nil, harness.KindInteger)
			if limitOf(err) != LimitTime {
				t.Errorf("%s time run: %v", f.name, err)
			}
		}
		right := runtime.NumGoroutine()
		final, waited := settle(goroutines, 5*time.Second)
		runtime.GC()
		runtime.ReadMemStats(&after)
		growth := int64(after.HeapInuse) - int64(before.HeapInuse)
		summary := fmt.Sprintf("goroutines %d before, %d right after the refusals, %d after waiting %s; heap in use changed by %d KB", goroutines, right, final, formatDuration(waited), growth/1024)
		record(tableRow{Case: "L08", Engine: f.name, Mode: "full", Layer: string(LayerNone), Time: formatDuration(waited), Sentence: summary})
		t.Log(f.name + ": " + summary)
		if final > goroutines {
			t.Errorf("%s: goroutines did not return: %s", f.name, summary)
		}
		if growth > 64<<20 {
			t.Errorf("%s: heap grew: %s", f.name, summary)
		}
	}
}

func TestL09HostDeadlineShorterThanTheTimeLimitWins(t *testing.T) {
	limits := slowLimits()
	limits.Time = 400 * time.Millisecond
	for _, f := range factories {
		engine := f.build(limits, true)
		program, err := engine.Compile(harness.Case{Source: heavyFormula(), Result: harness.KindInteger}, nil)
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Millisecond)
		started := time.Now()
		_, err = program.(contextual).RunContext(ctx, nil)
		took := time.Since(started)
		cancel()
		record(tableRow{Case: "L09", Engine: f.name, Mode: "full", Layer: layerOf(err), Time: formatDuration(took), Sentence: fmt.Sprint(err)})
		t.Logf("%s: host deadline 15ms, time limit 400ms: returned in %s with %v", f.name, took, err)
		if limitOf(err) != LimitTime || took > 100*time.Millisecond {
			t.Errorf("%s: %v after %s", f.name, err, took)
		}
		if err.Error() != LimitSentence(LimitTime, "15ms") {
			t.Errorf("%s: sentence %q", f.name, err)
		}
	}
}

func layerOf(err error) string {
	var failure *Failure
	if errors.As(err, &failure) {
		return string(failure.Layer)
	}
	return string(LayerNone)
}

func TestL10TenHostileFormulasAtOnceWhileAGoodOneAnswers(t *testing.T) {
	for _, f := range factories {
		engine, slow := f.build(Proposed(), true), f.build(slowLimits(), true)
		type attempt struct {
			id      string
			engine  harness.Engine
			formula string
			limit   string
			params  []harness.Param
			inputs  map[string]any
			result  harness.Kind
		}
		var attempts []attempt
		for _, id := range []string{"P04", "P05", "L01", "L03", "L05", "C01", "C02", "I01", "I10"} {
			s := scenarioNamed(id)
			attempts = append(attempts, attempt{id, engine, s.formula(Proposed()), "", s.params, s.inputs(Proposed()), s.result})
		}
		attempts = append(attempts, attempt{"L06", slow, heavyFormula(), LimitTime, nil, nil, harness.KindInteger})
		var wait sync.WaitGroup
		results := make([]string, len(attempts))
		started := time.Now()
		for i, a := range attempts {
			wait.Add(1)
			go func() {
				defer wait.Done()
				_, err := compileRun(a.engine, a.formula, a.params, a.inputs, a.result)
				results[i] = fmt.Sprintf("%s=%s", a.id, limitOf(err))
				if limitOf(err) == "none" || limitOf(err) == "other" {
					t.Errorf("%s on %s: %v", a.id, f.name, err)
				}
			}()
		}
		var goodTook time.Duration
		wait.Add(1)
		go func() {
			defer wait.Done()
			time.Sleep(2 * time.Millisecond)
			begin := time.Now()
			got, err := compileRun(engine, "2 + 3", nil, nil, harness.KindInteger)
			goodTook = time.Since(begin)
			if err != nil || got != int64(5) {
				t.Errorf("%s: good formula gave %v, %v", f.name, got, err)
			}
		}()
		wait.Wait()
		total := time.Since(started)
		summary := fmt.Sprintf("ten hostile at once, good answered 5 in %s: %v", formatDuration(goodTook), results)
		record(tableRow{Case: "L10", Engine: f.name, Mode: "full", Layer: string(LayerNone), Time: formatDuration(total), Sentence: summary})
		t.Log(f.name + ": " + summary)
	}
}
