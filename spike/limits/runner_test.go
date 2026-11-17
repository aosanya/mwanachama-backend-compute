package limits

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
)

const (
	childEnvironment = "LIMITS_CHILD"
	tableEnvironment = "LIMITS_TABLE"
	hardTimeout      = 20 * time.Second
)

type outcome struct {
	Case       string
	Engine     string
	Mode       string
	Status     string
	Limit      string
	Layer      string
	Sentence   string
	Value      string
	Native     string
	Nanos      int64
	AllocBytes uint64
}

func (o outcome) duration() time.Duration { return time.Duration(o.Nanos) }

func summarise(text string) string {
	text = strings.Join(strings.Fields(text), " ")
	if len(text) > 260 {
		text = text[:260] + "..."
	}
	return text
}

func execute(f engineFactory, s scenario, base Limits, precheck bool) outcome {
	limits := s.limits(base)
	mode := "native"
	if precheck {
		mode = "full"
	}
	out := outcome{Case: s.id, Engine: f.name, Mode: mode}
	formula := s.formula(limits)
	inputs := s.inputs(limits)
	done := make(chan outcome, 1)
	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	started := time.Now()
	go func() {
		result := out
		result.Status, result.Limit, result.Layer = "ok", "", string(LayerNone)
		engine := f.build(limits, precheck)
		c := harness.Case{Source: formula, Params: s.params, Result: s.result}
		program, err := engine.Compile(c, nil)
		var value any
		if err == nil {
			value, err = program.Run(inputs)
		}
		result.Nanos = time.Since(started).Nanoseconds()
		fill(&result, value, err)
		done <- result
	}()
	select {
	case out = <-done:
	case <-time.After(hardTimeout):
		out.Status, out.Nanos, out.Layer = "hang", hardTimeout.Nanoseconds(), string(LayerNone)
	}
	runtime.ReadMemStats(&after)
	out.AllocBytes = after.TotalAlloc - before.TotalAlloc
	return out
}

func fill(result *outcome, value any, err error) {
	if err == nil {
		result.Value = summarise(fmt.Sprint(value))
		if list, isList := value.([]int64); isList {
			result.Value = fmt.Sprintf("list of %d", len(list))
		}
		return
	}
	var failure *Failure
	if errors.As(err, &failure) {
		result.Status, result.Limit, result.Layer, result.Sentence = "refused", failure.Limit, string(failure.Layer), failure.Sentence
		if failure.Native != nil {
			result.Native = summarise(failure.Native.Error())
		} else {
			result.Native = "(no engine error: refused before the engine ran)"
		}
		return
	}
	result.Status, result.Sentence, result.Native = "error", summarise(err.Error()), summarise(err.Error())
}

func TestChildCase(t *testing.T) {
	spec := os.Getenv(childEnvironment)
	if spec == "" {
		t.Skip("only runs as a child process")
	}
	parts := strings.Split(spec, "|")
	out := execute(factoryNamed(parts[0]), scenarioNamed(parts[1]), Proposed(), parts[2] == "full")
	encoded, _ := json.Marshal(out)
	fmt.Printf("RESULT=%s\n", encoded)
}

func inChild(engine, id, mode string) outcome {
	command := exec.Command(os.Args[0], "-test.run=^TestChildCase$", "-test.count=1")
	command.Env = append(os.Environ(), childEnvironment+"="+engine+"|"+id+"|"+mode)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	started := time.Now()
	if err := command.Start(); err != nil {
		return outcome{Case: id, Engine: engine, Mode: mode, Status: "error", Sentence: err.Error()}
	}
	waited := make(chan error, 1)
	go func() { waited <- command.Wait() }()
	var err error
	select {
	case err = <-waited:
	case <-time.After(hardTimeout + 10*time.Second):
		_ = command.Process.Kill()
		<-waited
		return outcome{Case: id, Engine: engine, Mode: mode, Status: "hang", Layer: string(LayerNone), Nanos: time.Since(started).Nanoseconds()}
	}
	for _, line := range strings.Split(stdout.String(), "\n") {
		if text, found := strings.CutPrefix(line, "RESULT="); found {
			var out outcome
			if json.Unmarshal([]byte(text), &out) == nil {
				return out
			}
		}
	}
	reason := "the process ended without a result"
	if strings.Contains(stderr.String(), "stack overflow") || strings.Contains(stdout.String(), "stack overflow") {
		reason = "fatal error: stack overflow (the whole process died)"
	} else if strings.Contains(stderr.String()+stdout.String(), "out of memory") {
		reason = "fatal error: out of memory (the whole process died)"
	} else if err != nil {
		reason = "process ended: " + err.Error()
	}
	return outcome{Case: id, Engine: engine, Mode: mode, Status: "crash", Layer: string(LayerNone), Native: reason, Nanos: time.Since(started).Nanoseconds()}
}

type tableRow struct {
	Case, Engine, Mode, Layer, Time, Sentence string
}

var (
	tableMutex sync.Mutex
	tableRows  []tableRow
)

func record(row tableRow) {
	tableMutex.Lock()
	defer tableMutex.Unlock()
	tableRows = append(tableRows, row)
}

func recordOutcome(o outcome) {
	sentence := o.Sentence
	switch o.Status {
	case "ok":
		sentence = "(no refusal) result " + o.Value
	case "hang":
		sentence = "HANG: no answer within the hard timeout"
	case "crash":
		sentence = "CRASH: " + o.Native
	}
	record(tableRow{Case: o.Case, Engine: o.Engine, Mode: o.Mode, Layer: o.Layer, Time: formatDuration(o.duration()), Sentence: sentence})
}

func formatDuration(d time.Duration) string {
	switch {
	case d >= time.Second:
		return fmt.Sprintf("%.2f s", d.Seconds())
	case d >= time.Millisecond:
		return fmt.Sprintf("%.2f ms", float64(d)/1e6)
	}
	return fmt.Sprintf("%d us", d.Microseconds())
}

func TestMain(m *testing.M) {
	code := m.Run()
	if path := os.Getenv(tableEnvironment); path != "" && os.Getenv(childEnvironment) == "" {
		writeTable(path)
	}
	os.Exit(code)
}

func writeTable(path string) {
	tableMutex.Lock()
	defer tableMutex.Unlock()
	sort.SliceStable(tableRows, func(i, j int) bool {
		a, b := tableRows[i], tableRows[j]
		if a.Case != b.Case {
			return a.Case < b.Case
		}
		if a.Engine != b.Engine {
			return a.Engine < b.Engine
		}
		return a.Mode < b.Mode
	})
	var text strings.Builder
	fmt.Fprintf(&text, "%-6s %-10s %-7s %-9s %-10s %s\n", "case", "engine", "mode", "layer", "time", "sentence")
	for _, row := range tableRows {
		fmt.Fprintf(&text, "%-6s %-10s %-7s %-9s %-10s %s\n", row.Case, row.Engine, row.Mode, row.Layer, row.Time, row.Sentence)
	}
	_ = os.WriteFile(path, []byte(text.String()), 0o644)
}
