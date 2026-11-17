package limits

import (
	"fmt"
	"slices"
	"strings"
	"testing"
)

func (s scenario) fits(o outcome) bool {
	if s.expect.value != nil {
		return o.Status == "ok" && o.Value == fmt.Sprint(s.expect.value)
	}
	return o.Status == "refused" && slices.Contains(s.expect.limits, o.Limit)
}

func TestMatrix(t *testing.T) {
	for _, s := range scenarios() {
		for _, f := range factories {
			for _, mode := range []string{"native", "full"} {
				o := inChild(f.name, s.id, mode)
				recordOutcome(o)
				t.Logf("%-5s %-10s %-6s %-8s %-8s %-10s %s | raw: %s | alloc %d KB", s.id, f.name, mode, o.Status, o.Layer, formatDuration(o.duration()), o.Sentence+o.Value, o.Native, o.AllocBytes/1024)
				if mode == "native" {
					continue
				}
				assertScenario(t, s, f.name, o)
			}
		}
	}
}

func assertScenario(t *testing.T, s scenario, engine string, o outcome) {
	t.Helper()
	if o.Status == "hang" || o.Status == "crash" {
		t.Errorf("%s on %s with the full wrapper: %s %s", s.id, engine, o.Status, o.Native)
		return
	}
	if !s.fits(o) && slices.Contains(s.expect.strict, engine) {
		t.Errorf("%s on %s: got %s %q %q, want limit %v or value %v", s.id, engine, o.Status, o.Sentence, o.Value, s.expect.limits, s.expect.value)
	}
	if o.Status == "refused" && o.Limit != "other" && o.Limit != "parser" && !strings.HasPrefix(o.Sentence, "This formula goes over the "+o.Limit+" limit of ") {
		t.Errorf("%s on %s: sentence %q is not the limit sentence", s.id, engine, o.Sentence)
	}
	if s.maxRefusal > 0 && o.duration() > s.maxRefusal {
		t.Errorf("%s on %s: refused in %v, bound %v", s.id, engine, o.duration(), s.maxRefusal)
	}
	if s.id == "P20" || s.id == "P20b" {
		if o.Sentence != "The formula cannot be read: the formula is empty at line 1, column 1." {
			t.Errorf("%s on %s: sentence %q", s.id, engine, o.Sentence)
		}
	}
}
