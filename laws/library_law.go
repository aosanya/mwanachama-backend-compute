package laws

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"

	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/library"
	"github.com/aosanya/mwanachama-backend-compute/number"
)

const pinnedSeedOffset = 1 << 22

var pmtSignature = formula.Signature{
	Params: []formula.Param{
		{Name: "principal", Type: formula.TypeDecimal},
		{Name: "rate", Type: formula.TypeDecimal},
		{Name: "months", Type: formula.TypeInteger},
	},
	Result: formula.TypeDecimal,
	Cost:   20,
}

var pmtNote = library.Note{
	Changed: "The payment is rounded to the nearest 0.01. It was rounded down to a whole number.",
	Matters: "Any agency that shows or collects a loan payment.",
	Example: library.Example{Input: "12000.00 at 0.01 over 24 months", Old: "564", New: "564.88"},
}

func exactPayment(args []formula.Value) float64 {
	principal, _ := strconv.ParseFloat(args[0].Decimal().String(), 64)
	rate, _ := strconv.ParseFloat(args[1].Decimal().String(), 64)
	months := float64(args[2].Integer())
	return principal * rate / (1 - math.Pow(1+rate, -months))
}

func decimalText(text string) formula.Value {
	parsed, err := number.ParseDecimal(text)
	if err != nil {
		panic(err)
	}
	return formula.DecimalValue(parsed)
}

func PmtRoundedDown(args []formula.Value) (formula.Value, *formula.Problem) {
	return decimalText(strconv.Itoa(int(math.Floor(exactPayment(args))))), nil
}

func PmtRoundedToCents(args []formula.Value) (formula.Value, *formula.Problem) {
	return decimalText(strconv.FormatFloat(math.Round(exactPayment(args)*100)/100, 'f', 2, 64)), nil
}

func FixtureReleases() []library.Release {
	return []library.Release{
		{Library: "fin", Name: "pmt", Version: 1, Signature: pmtSignature, Apply: PmtRoundedDown},
		{Library: "fin", Name: "pmt", Version: 2, Signature: pmtSignature, Apply: PmtRoundedToCents, Note: pmtNote},
	}
}

func RegistryOf(releases []library.Release) *library.Registry {
	registry := library.NewRegistry()
	for _, release := range releases {
		if err := registry.Publish(release); err != nil {
			panic(err)
		}
	}
	return registry
}

func FixtureLibrary() *library.Registry {
	return RegistryOf(FixtureReleases())
}

type Reference struct {
	Ref   library.Ref
	Args  []formula.Value
	Given string
	Want  string
}

func ParseReferences(text string, registry *library.Registry) ([]Reference, error) {
	var references []Reference
	for lineNumber, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		left, want, found := strings.Cut(line, " => ")
		if !found {
			return nil, fmt.Errorf("line %d has no ' => '", lineNumber+1)
		}
		fields := strings.Fields(left)
		ref, refused := library.ParseRef(fields[0])
		if refused != nil {
			return nil, fmt.Errorf("line %d: %s", lineNumber+1, refused.Sentence)
		}
		signature, published := registry.Function(ref.Library, ref.Name, ref.Version)
		if !published {
			return nil, fmt.Errorf("line %d: %s is not published", lineNumber+1, ref)
		}
		if len(fields)-1 != len(signature.Params) {
			return nil, fmt.Errorf("line %d: %s takes %d inputs, the line has %d", lineNumber+1, ref, len(signature.Params), len(fields)-1)
		}
		args := make([]formula.Value, len(signature.Params))
		for i, param := range signature.Params {
			value, err := valueOfText(param.Type, fields[i+1])
			if err != nil {
				return nil, fmt.Errorf("line %d: %v", lineNumber+1, err)
			}
			args[i] = value
		}
		references = append(references, Reference{Ref: ref, Args: args, Given: strings.Join(fields[1:], " "), Want: strings.TrimSpace(want)})
	}
	return references, nil
}

func valueOfText(kind formula.Type, text string) (formula.Value, error) {
	switch kind {
	case formula.TypeInteger:
		whole, err := strconv.ParseInt(text, 10, 64)
		return formula.IntegerValue(whole), err
	case formula.TypeDecimal:
		parsed, err := number.ParseDecimal(text)
		return formula.DecimalValue(parsed), err
	}
	return formula.Value{}, fmt.Errorf("a reference input of type %s is not supported", kind)
}

func ReferencesNeverChange(engine Engine, recorded string) *Failure {
	registry := engine.Library()
	references, err := ParseReferences(recorded, registry)
	if err != nil {
		return &Failure{Law: "published function", Message: "references could not be read: " + err.Error()}
	}
	recordedVersions := map[string]int{}
	for _, reference := range references {
		key := reference.Ref.Library + "." + reference.Ref.Name
		if reference.Ref.Version > recordedVersions[key] {
			recordedVersions[key] = reference.Ref.Version
		}
		value, refused := registry.Apply(reference.Ref.Library, reference.Ref.Name, reference.Ref.Version, reference.Args)
		now := ""
		if refused != nil {
			now = refused.Code
		} else {
			now = value.String()
		}
		if now != reference.Want {
			return &Failure{Law: "published function", Message: fmt.Sprintf("reference changed, %s, input %s, recorded %s, now %s; ship the fix as a new version", reference.Ref, reference.Given, reference.Want, now)}
		}
	}
	for key, recordedUpTo := range recordedVersions {
		lib, name, _ := strings.Cut(key, ".")
		for version := 1; version <= len(registry.Versions(lib, name)); version++ {
			if !hasReference(references, lib, name, version) {
				return &Failure{Law: "published function", Message: fmt.Sprintf("no reference recorded for %s@%d, recorded up to %s@%d; record the results it gives when it is published", key, version, key, recordedUpTo)}
			}
		}
	}
	return nil
}

func hasReference(references []Reference, lib string, name string, version int) bool {
	for _, reference := range references {
		if reference.Ref.Library == lib && reference.Ref.Name == name && reference.Ref.Version == version {
			return true
		}
	}
	return false
}

var pinnedInputs = []formula.Input{
	{Name: "principal", Type: formula.TypeDecimal},
	{Name: "rate", Type: formula.TypeDecimal},
	{Name: "months", Type: formula.TypeInteger},
}

func pinnedRows(rng *rand.Rand) []map[string]formula.Value {
	rows := make([]map[string]formula.Value, 1+rng.Intn(rowsPerCase))
	for i := range rows {
		rows[i] = map[string]formula.Value{
			"principal": formula.DecimalValue(number.NewDecimal(int64(100000+rng.Intn(9900000)), 2)),
			"rate":      formula.DecimalValue(number.NewDecimal(int64(1+rng.Intn(50)), 3)),
			"months":    formula.IntegerValue(int64(1 + rng.Intn(60))),
		}
	}
	return rows
}

func pinnedCase(engine Engine, seed int64) (library.Ref, formula.Tree, []map[string]formula.Value, bool) {
	rng := rand.New(rand.NewSource(seed ^ pinnedSeedOffset))
	versions := engine.Library().Versions("fin", "pmt")
	if len(versions) == 0 {
		return library.Ref{}, formula.Tree{}, nil, false
	}
	ref := library.Ref{Library: "fin", Name: "pmt", Version: versions[rng.Intn(len(versions))]}
	tree, _ := formula.Parse(ref.String() + "(principal, rate, months)")
	return ref, tree, pinnedRows(rng), true
}

func pinnedOutcomes(registry *library.Registry, tree formula.Tree, rows []map[string]formula.Value) []string {
	outcomes := make([]string, len(rows))
	for i, values := range rows {
		outcomes[i] = outcomeWith(registry, tree, pinnedInputs, values)
	}
	return outcomes
}

func PinnedVersionsSameOutput(engine Engine, seed int64) *Failure {
	ref, tree, rows, found := pinnedCase(engine, seed)
	if !found {
		return &Failure{Law: "same output", Seed: seed, Message: fmt.Sprintf("output moved between runs, seed %d: the library has no fin.pmt to pin", seed)}
	}
	first := pinnedOutcomes(engine.Library(), tree, rows)
	second := pinnedOutcomes(engine.Library(), tree, rows)
	for i := range first {
		if first[i] != second[i] {
			return &Failure{Law: "same output", Seed: seed, Shrunk: Show(tree), Message: fmt.Sprintf("output moved between runs, seed %d, pinned %s, row %d: run 1 gave %s, run 2 gave %s", seed, ref, i+1, first[i], second[i])}
		}
	}
	return nil
}

func PinnedDigest(engine Engine, base int64, count int) string {
	hash := sha256.New()
	for i := int64(1); i <= int64(count); i++ {
		ref, tree, rows, found := pinnedCase(engine, base+i)
		if !found {
			continue
		}
		fmt.Fprintf(hash, "%d %s\n", base+i, ref)
		for _, outcome := range pinnedOutcomes(engine.Library(), tree, rows) {
			fmt.Fprintf(hash, "%s\n", outcome)
		}
	}
	return hex.EncodeToString(hash.Sum(nil))
}
