package laws

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/library"
)

func recordedReferences(t *testing.T) string {
	t.Helper()
	recorded, err := os.ReadFile("testdata/references.txt")
	if err != nil {
		t.Fatal(err)
	}
	return string(recorded)
}

func TestLawPublishedFunctionNeverChanges(t *testing.T) {
	if failure := ReferencesNeverChange(RealEngine(), recordedReferences(t)); failure != nil {
		t.Fatal(failure.Message)
	}
}

func TestEveryRecordedReferenceIsReadAndEveryPublishedVersionHasOne(t *testing.T) {
	registry := FixtureLibrary()
	references, err := ParseReferences(recordedReferences(t), registry)
	if err != nil {
		t.Fatal(err)
	}
	for _, release := range FixtureReleases() {
		if !hasReference(references, release.Library, release.Name, release.Version) {
			t.Fatalf("%s@%d has no recorded reference", release.Name, release.Version)
		}
	}
}

func TestLawPinnedVersionsGiveTheSameOutput(t *testing.T) {
	runLaw(t, PinnedVersionsSameOutput)
}

func TestLawPinnedVersionsGiveTheSameOutputInASecondProcess(t *testing.T) {
	if os.Getenv("LAWS_CHILD") == "1" {
		t.Skip("running as the second process")
	}
	base, count := BaseSeed(), Cases()
	here := PinnedDigest(RealEngine(), base, count)
	command := exec.Command(os.Args[0], "-test.run=^TestPrintPinnedDigest$", "-test.v")
	command.Env = append(os.Environ(), "LAWS_CHILD=1")
	printed, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("second process failed: %v\n%s", err, printed)
	}
	if !strings.Contains(string(printed), "digest "+here) {
		t.Fatalf("output moved between processes, seeds %d to %d, pinned versions, this process gave digest %s, the second printed:\n%s", base+1, base+int64(count), here, printed)
	}
}

func TestPrintPinnedDigest(t *testing.T) {
	if os.Getenv("LAWS_CHILD") != "1" {
		t.Skip("only the second process prints")
	}
	t.Logf("digest %s", PinnedDigest(RealEngine(), BaseSeed(), Cases()))
}

func TestPinnedLawTriesEveryPublishedVersion(t *testing.T) {
	engine := RealEngine()
	seen := map[int]bool{}
	for seed := int64(1); seed <= 100; seed++ {
		ref, _, rows, found := pinnedCase(engine, seed)
		if !found || len(rows) == 0 {
			t.Fatalf("seed %d gave no pinned case", seed)
		}
		seen[ref.Version] = true
	}
	if len(seen) != len(FixtureReleases()) {
		t.Fatalf("100 seeds pinned versions %v, wanted all %d", seen, len(FixtureReleases()))
	}
}

func TestBrokenReleaseThatMovesBetweenRunsFailsThePinnedLaw(t *testing.T) {
	calls := 0
	broken := RealEngine()
	broken.Library = func() *library.Registry {
		releases := FixtureReleases()
		for i := range releases {
			if releases[i].Version == 2 {
				real := releases[i].Apply
				releases[i].Apply = func(args []formula.Value) (formula.Value, *formula.Problem) {
					calls++
					value, refused := real(args)
					if calls%2 == 0 {
						return decimalText("0.01"), refused
					}
					return value, refused
				}
			}
		}
		return RegistryOf(releases)
	}
	failure := proveLawFails(t, PinnedVersionsSameOutput, broken, "output moved between runs, seed ")
	if !strings.Contains(failure.Message, ", pinned fin.pmt@2, row ") {
		t.Fatalf("message %q does not name the pinned version", failure.Message)
	}
}

func TestBrokenFunctionEditedInPlaceFailsTheReferenceLaw(t *testing.T) {
	broken := RealEngine()
	broken.Library = func() *library.Registry {
		releases := FixtureReleases()
		releases[0].Apply = PmtRoundedToCents
		return RegistryOf(releases)
	}
	failure := ReferencesNeverChange(broken, recordedReferences(t))
	if failure == nil {
		t.Fatal("a function edited in place passed the reference law")
	}
	if !strings.HasPrefix(failure.Message, "reference changed, fin.pmt@1, input ") || !strings.HasSuffix(failure.Message, "; ship the fix as a new version") {
		t.Fatalf("message %q does not read as the page's sentence", failure.Message)
	}
	if !strings.Contains(failure.Message, ", recorded 564, now 564.88;") && !strings.Contains(failure.Message, ", recorded ") {
		t.Fatalf("message %q does not say what was recorded", failure.Message)
	}
	if healthy := ReferencesNeverChange(RealEngine(), recordedReferences(t)); healthy != nil {
		t.Fatalf("the real library fails the references too: %s", healthy.Message)
	}
}

func TestBrokenPublishWithoutAReferenceFailsTheReferenceLaw(t *testing.T) {
	broken := RealEngine()
	broken.Library = func() *library.Registry {
		releases := FixtureReleases()
		releases = append(releases, library.Release{
			Library: "fin", Name: "pmt", Version: 3, Signature: pmtSignature, Apply: PmtRoundedToCents,
			Note: pmtNote,
		})
		return RegistryOf(releases)
	}
	failure := ReferencesNeverChange(broken, recordedReferences(t))
	if failure == nil {
		t.Fatal("a published version with no recorded reference passed")
	}
	if !strings.HasPrefix(failure.Message, "no reference recorded for fin.pmt@3") {
		t.Fatalf("message %q does not name the version without a reference", failure.Message)
	}
}

func TestRecordReferences(t *testing.T) {
	if os.Getenv("LAWS_RECORD") != "1" {
		t.Skip("set LAWS_RECORD=1 to print the lines to record for a newly published version")
	}
	inputs := []string{"12000.00 0.01 24", "5000.00 0.02 12", "250000.00 0.005 60", "900.50 0.03 6", "75000.00 0.0125 36"}
	registry := FixtureLibrary()
	for _, release := range FixtureReleases() {
		for _, given := range inputs {
			text := release.Library + "." + release.Name + "@" + string(rune('0'+release.Version)) + " " + given + " => 0"
			references, err := ParseReferences(text, registry)
			if err != nil {
				t.Fatal(err)
			}
			value, _ := registry.Apply(release.Library, release.Name, release.Version, references[0].Args)
			t.Logf("%s %s => %s", references[0].Ref, given, value.String())
		}
	}
}
