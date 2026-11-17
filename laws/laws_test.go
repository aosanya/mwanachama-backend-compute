package laws

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func runLaw(t *testing.T, law func(Engine, int64) *Failure) {
	t.Helper()
	engine := RealEngine()
	base, count := BaseSeed(), Cases()
	t.Logf("%d cases, seeds %d to %d", count, base+1, base+int64(count))
	for i := int64(1); i <= int64(count); i++ {
		if failure := law(engine, base+i); failure != nil {
			t.Fatal(failure.Message)
		}
	}
}

func TestLawRoundTrip(t *testing.T) {
	runLaw(t, RoundTrip)
}

func TestLawStandardShapeIsStable(t *testing.T) {
	runLaw(t, StandardShape)
}

func TestLawSpacingDoesNotChangeTheHash(t *testing.T) {
	runLaw(t, SpacingKeepsHash)
}

func TestLawSameInputsSameOutputInOneProcess(t *testing.T) {
	runLaw(t, SameOutput)
}

func TestLawNoRowWithProblemsIsCalculated(t *testing.T) {
	runLaw(t, NoRowWithProblemsIsCalculated)
}

func TestLawSameInputsSameOutputInASecondProcess(t *testing.T) {
	if os.Getenv("LAWS_CHILD") == "1" {
		t.Skip("running as the second process")
	}
	base, count := BaseSeed(), Cases()
	here := OutputDigest(RealEngine(), base, count)
	command := exec.Command(os.Args[0], "-test.run=^TestPrintOutputDigest$", "-test.v")
	command.Env = append(os.Environ(), "LAWS_CHILD=1")
	printed, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("second process failed: %v\n%s", err, printed)
	}
	if !strings.Contains(string(printed), "digest "+here) {
		t.Fatalf("output moved between processes, seeds %d to %d, this process gave digest %s, the second printed:\n%s", base+1, base+int64(count), here, printed)
	}
}

func TestPrintOutputDigest(t *testing.T) {
	if os.Getenv("LAWS_CHILD") != "1" {
		t.Skip("only the second process prints")
	}
	t.Logf("digest %s", OutputDigest(RealEngine(), BaseSeed(), Cases()))
}

func TestLawSameOutputOnEveryMachine(t *testing.T) {
	if os.Getenv("LAWS_PLATFORM_GOLDEN") != "1" {
		t.Skip("set LAWS_PLATFORM_GOLDEN=1 on every platform the CI offers; the digest recorded in testdata/output-digest.txt must be the same on all of them")
	}
	recorded, err := os.ReadFile("testdata/output-digest.txt")
	if err != nil {
		t.Fatal(err)
	}
	got := OutputDigest(RealEngine(), 0, 500)
	if want := strings.TrimSpace(string(recorded)); got != want {
		t.Fatalf("output moved between machines, seeds 1 to 500, recorded digest %s, this machine gave %s", want, got)
	}
}

func TestLawBatchEqualsSingleWaitsForTheBatchBuild(t *testing.T) {
	t.Skip("waiting: law 4 is switched on when the batch evaluation build (ISS-29) lands")
}

func TestLawPublishedFunctionNeverChangesWaitsForTheLibraryBuild(t *testing.T) {
	t.Skip("waiting: law 6 is switched on when the first ten library functions with recorded reference values (ISS-27) land")
}

func TestLawProblemListIsCompleteWaitsForValidate(t *testing.T) {
	t.Skip("waiting: the complete problem list half of law 7 is switched on when Validate (ISS-26) is done")
}
