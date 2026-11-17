package functions

import (
	"strings"
	"testing"
)

func replayed(t *testing.T) (Platform, map[string][]Case) {
	t.Helper()
	versions := loadReference(t)
	got, err := Replay(versions)
	if err != nil {
		t.Fatal(err)
	}
	return got, versions
}

func recordedPlatform(t *testing.T) Platform {
	t.Helper()
	recorded, err := LoadPlatform(referenceDir)
	if err != nil {
		t.Fatalf("the recorded platform results cannot be read: %v", err)
	}
	return recorded
}

func copyPlatform(p Platform) Platform {
	copied := Platform{Name: p.Name, Results: map[string]PlatformResult{}}
	for id, result := range p.Results {
		copied.Results[id] = result
	}
	return copied
}

func TestEveryCaseGivesTheRecordedResultOnThisPlatform(t *testing.T) {
	got, versions := replayed(t)
	for _, failure := range ComparePlatforms(recordedPlatform(t), got, versions) {
		t.Error(failure)
	}
}

func TestThePlatformFileHoldsEveryRecordedCaseAndNothingElse(t *testing.T) {
	got, _ := replayed(t)
	recorded := recordedPlatform(t)
	if len(recorded.Results) != len(got.Results) || len(recorded.Results) != 152 {
		t.Fatalf("the file holds %d results, a replay gives %d, the pages list 152", len(recorded.Results), len(got.Results))
	}
	if recorded.Name == "" {
		t.Error("the file does not name the platform it was recorded on")
	}
}

func TestADifferenceOnAnotherPlatformFailsWithTheSentence(t *testing.T) {
	got, versions := replayed(t)
	other := copyPlatform(got)
	other.Name = "linux/amd64"
	id := "REF-fin.balance_after_payment-01"
	other.Results[id] = PlatformResult{Key: "fin.balance_after_payment@1", Text: "76.25"}
	failures := ComparePlatforms(got, other, versions)
	want := id + ": fin.balance_after_payment@1 gave 76.25 on linux/amd64 and 75.25 on " + got.Name + ". A library function gives the same answer on every machine."
	if len(failures) != 1 || failures[0] != want {
		t.Fatalf("failures %v, want %q", failures, want)
	}
}

func TestAnIdenticalPlatformHasNoFailure(t *testing.T) {
	got, versions := replayed(t)
	if failures := ComparePlatforms(got, copyPlatform(got), versions); len(failures) != 0 {
		t.Fatalf("failures %v", failures)
	}
}

func TestAFloatWithinTheToleranceOfThePageAgrees(t *testing.T) {
	got, versions := replayed(t)
	other := copyPlatform(got)
	other.Name = "linux/amd64"
	id := "REF-fin.coverage_ratio-01"
	other.Results[id] = PlatformResult{Key: "fin.coverage_ratio@1", Text: "1.3630000000001"}
	if failures := ComparePlatforms(got, other, versions); len(failures) != 0 {
		t.Fatalf("a relative difference of 1e-13 failed: %v", failures)
	}
	other.Results[id] = PlatformResult{Key: "fin.coverage_ratio@1", Text: "1.364"}
	if failures := ComparePlatforms(got, other, versions); len(failures) != 1 {
		t.Fatalf("a relative difference of 7e-4 gave %v", failures)
	}
}

func TestAFloatThatIsZeroMustStayExactlyZero(t *testing.T) {
	got, versions := replayed(t)
	other := copyPlatform(got)
	other.Name = "linux/amd64"
	id := "REF-fin.coverage_ratio-03"
	if got.Results[id].Text != "0" {
		t.Fatalf("%s gave %q, the test expects 0", id, got.Results[id].Text)
	}
	other.Results[id] = PlatformResult{Key: "fin.coverage_ratio@1", Text: "0.000000000001"}
	if failures := ComparePlatforms(got, other, versions); len(failures) != 1 {
		t.Fatalf("a float noise on an expected zero gave %v", failures)
	}
}

func TestADecimalIsNeverComparedWithATolerance(t *testing.T) {
	got, versions := replayed(t)
	other := copyPlatform(got)
	other.Name = "linux/amd64"
	id := "REF-fin.equal_share-01"
	other.Results[id] = PlatformResult{Key: "fin.equal_share@1", Text: "328.610000001"}
	if failures := ComparePlatforms(got, other, versions); len(failures) != 1 {
		t.Fatalf("a decimal off by a billionth gave %v", failures)
	}
}

func TestARefusalForAnotherRuleFails(t *testing.T) {
	got, versions := replayed(t)
	other := copyPlatform(got)
	other.Name = "linux/amd64"
	id := "REF-fin.balance_after_payment-07"
	if got.Results[id].Text != "refused:no-negatives" {
		t.Fatalf("%s gave %q", id, got.Results[id].Text)
	}
	other.Results[id] = PlatformResult{Key: "fin.balance_after_payment@1", Text: "refused:too-small"}
	if failures := ComparePlatforms(got, other, versions); len(failures) != 1 || !strings.Contains(failures[0], "refused:too-small on linux/amd64 and refused:no-negatives") {
		t.Fatalf("failures %v", failures)
	}
}

func TestAResultMissingOnOneSideFails(t *testing.T) {
	got, versions := replayed(t)
	other := copyPlatform(got)
	other.Name = "linux/amd64"
	delete(other.Results, "REF-math.remainder-01")
	failures := ComparePlatforms(got, other, versions)
	if len(failures) != 1 || failures[0] != "REF-math.remainder-01: recorded on "+got.Name+" and not run on linux/amd64." {
		t.Fatalf("failures %v", failures)
	}
	failures = ComparePlatforms(other, got, versions)
	if len(failures) != 1 || failures[0] != "REF-math.remainder-01: math.remainder@1 has no result recorded on linux/amd64." {
		t.Fatalf("failures %v", failures)
	}
}

func TestAReplayGivesTheSameResultsTwice(t *testing.T) {
	first, _ := replayed(t)
	second, _ := replayed(t)
	if strings.Join(first.Lines(), "\n") != strings.Join(second.Lines(), "\n") {
		t.Error("two replays differ")
	}
}

func TestThePlatformFileIsReadBackAsItWasWritten(t *testing.T) {
	got, _ := replayed(t)
	back, err := ParsePlatform(strings.Join(got.Lines(), "\n") + "\n")
	if err != nil {
		t.Fatal(err)
	}
	if back.Name != got.Name || len(back.Results) != len(got.Results) {
		t.Fatalf("read back %q with %d results", back.Name, len(back.Results))
	}
}

func TestAMalformedPlatformFileIsRefused(t *testing.T) {
	for _, text := range []string{"REF-a-01\tfin.a@1\t5\n", "platform\tx\nREF-a-01\t5\n"} {
		if _, err := ParsePlatform(text); err == nil {
			t.Errorf("%q was accepted", text)
		}
	}
}
