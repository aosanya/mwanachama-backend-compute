package functions

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

const platformFile = "platform-results.txt"

type PlatformResult struct {
	Key  string
	Text string
}

type Platform struct {
	Name    string
	Results map[string]PlatformResult
}

func ThisPlatform() string {
	return runtime.GOOS + "/" + runtime.GOARCH
}

func refusalCode(out Outcome) string {
	if len(out.Problems) > 0 {
		return out.Problems[0].Code
	}
	if out.Failure != nil {
		return "run:" + out.Failure.Code
	}
	return ""
}

func ResultText(out Outcome) string {
	if out.Refused() {
		return RefusedText + ":" + refusalCode(out)
	}
	return out.Value.String()
}

func Replay(versions map[string][]Case) (Platform, error) {
	platform := Platform{Name: ThisPlatform(), Results: map[string]PlatformResult{}}
	for _, key := range SortedKeys(versions) {
		candidate, found := Find(key)
		if !found {
			return Platform{}, fmt.Errorf("%s is recorded and the library does not have it", key)
		}
		callable, err := candidate.Prepare()
		if err != nil {
			return Platform{}, err
		}
		for _, c := range versions[key] {
			platform.Results[c.ID] = PlatformResult{Key: key, Text: ResultText(callable.Call(RowOf(c)))}
		}
	}
	return platform, nil
}

func (p Platform) Lines() []string {
	lines := []string{"platform\t" + p.Name}
	for _, id := range SortedKeys(p.Results) {
		result := p.Results[id]
		lines = append(lines, id+"\t"+result.Key+"\t"+result.Text)
	}
	return lines
}

func (p Platform) Write(dir string) error {
	return os.WriteFile(dir+"/"+platformFile, []byte(strings.Join(p.Lines(), "\n")+"\n"), 0o644)
}

func LoadPlatform(dir string) (Platform, error) {
	raw, err := os.ReadFile(dir + "/" + platformFile)
	if err != nil {
		return Platform{}, err
	}
	return ParsePlatform(string(raw))
}

func ParsePlatform(text string) (Platform, error) {
	platform := Platform{Results: map[string]PlatformResult{}}
	for number, line := range strings.Split(strings.TrimRight(text, "\n"), "\n") {
		fields := strings.Split(line, "\t")
		if number == 0 {
			if len(fields) != 2 || fields[0] != "platform" {
				return Platform{}, fmt.Errorf("line 1 of %s must be `platform<TAB>name`", platformFile)
			}
			platform.Name = fields[1]
			continue
		}
		if len(fields) != 3 {
			return Platform{}, fmt.Errorf("line %d of %s must be `id<TAB>function<TAB>result`", number+1, platformFile)
		}
		platform.Results[fields[0]] = PlatformResult{Key: fields[1], Text: fields[2]}
	}
	return platform, nil
}

func floatsAgree(recorded string, got string) bool {
	if recorded == got {
		return true
	}
	a, errA := strconv.ParseFloat(recorded, 64)
	b, errB := strconv.ParseFloat(got, 64)
	if errA != nil || errB != nil || a == 0 {
		return false
	}
	return math.Abs(b-a) <= 1e-9*math.Abs(a)
}

func ComparePlatforms(recorded Platform, got Platform, cases map[string][]Case) []string {
	isFloat := map[string]bool{}
	for _, list := range cases {
		for _, c := range list {
			isFloat[c.ID] = c.Expected != RefusedText && strings.HasSuffix(c.Expected, "f")
		}
	}
	var failures []string
	for _, id := range SortedKeys(got.Results) {
		now := got.Results[id]
		before, held := recorded.Results[id]
		if !held {
			failures = append(failures, id+": "+now.Key+" has no result recorded on "+recorded.Name+".")
			continue
		}
		same := before.Text == now.Text || (isFloat[id] && floatsAgree(before.Text, now.Text))
		if !same {
			failures = append(failures, id+": "+now.Key+" gave "+now.Text+" on "+got.Name+" and "+before.Text+" on "+recorded.Name+". A library function gives the same answer on every machine.")
		}
	}
	missing := []string{}
	for id := range recorded.Results {
		if _, ran := got.Results[id]; !ran {
			missing = append(missing, id)
		}
	}
	sort.Strings(missing)
	for _, id := range missing {
		failures = append(failures, id+": recorded on "+recorded.Name+" and not run on "+got.Name+".")
	}
	return failures
}
