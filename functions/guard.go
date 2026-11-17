package functions

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	StateDraft     = "draft"
	StatePublished = "published"
)

var (
	pageHeading = regexp.MustCompile(`^### ([a-z]+\.[a-z_]+@\d+)\s*$`)
	pageCase    = regexp.MustCompile(`^\| (REF-[^ |]+) \|`)
	sourceRow   = regexp.MustCompile(`^\| (S-\d+) \|`)
)

type ManifestVersion struct {
	State string            `json:"state"`
	Cases map[string]string `json:"cases"`
}

type Manifest map[string]ManifestVersion

func LoadManifest(dir string) (Manifest, error) {
	raw, err := os.ReadFile(filepath.Join(dir, manifestFile))
	if os.IsNotExist(err) {
		return Manifest{}, nil
	}
	if err != nil {
		return nil, err
	}
	manifest := Manifest{}
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return nil, err
	}
	return manifest, nil
}

func (m Manifest) Write(dir string) error {
	raw, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, manifestFile), append(raw, '\n'), 0o644)
}

func PageCaseIDs(markdown string) map[string][]string {
	ids := map[string][]string{}
	current := ""
	for _, line := range strings.Split(markdown, "\n") {
		if heading := pageHeading.FindStringSubmatch(line); heading != nil {
			current = heading[1]
			continue
		}
		if strings.HasPrefix(line, "## ") {
			current = ""
		}
		if row := pageCase.FindStringSubmatch(line); row != nil && current != "" {
			ids[current] = append(ids[current], row[1])
		}
	}
	return ids
}

func SourceIDs(markdown string) map[string]bool {
	ids := map[string]bool{}
	for _, line := range strings.Split(markdown, "\n") {
		if row := sourceRow.FindStringSubmatch(line); row != nil {
			ids[row[1]] = true
		}
	}
	return ids
}

type Guard struct {
	Dir      string
	Cases    map[string][]Case
	Manifest Manifest
	Served   map[string][]string
	Page     map[string][]string
	Sources  map[string]bool
}

type Report struct {
	Failures []string
	Notices  []string
}

func (g Guard) dataFile(key string) string {
	return g.Dir + "/" + key + ".json"
}

func (g Guard) orderOf(key string, cases []Case) []string {
	if order, served := g.Served[key]; served {
		return order
	}
	names := map[string]bool{}
	for _, c := range cases {
		for name := range c.Inputs {
			names[name] = true
		}
	}
	return SortedKeys(names)
}

func (g Guard) Check() Report {
	var report Report
	for _, key := range SortedKeys(g.Manifest) {
		g.checkRecorded(&report, key)
	}
	for _, key := range SortedKeys(g.Cases) {
		g.checkSources(&report, key)
		g.checkNew(&report, key)
	}
	g.checkPage(&report)
	return report
}

func (g Guard) checkRecorded(report *Report, key string) {
	version := g.Manifest[key]
	if version.State != StatePublished {
		return
	}
	if _, served := g.Served[key]; !served {
		report.Failures = append(report.Failures, key+" is published but the library no longer has it. A published version is never removed.")
	}
	cases, hasFile := g.Cases[key]
	if !hasFile {
		report.Failures = append(report.Failures, key+" is published but "+g.dataFile(key)+" does not exist.")
		return
	}
	byID := map[string]Case{}
	for _, c := range cases {
		byID[c.ID] = c
	}
	order := g.orderOf(key, cases)
	for _, id := range SortedKeys(version.Cases) {
		current, present := byID[id]
		if !present {
			report.Failures = append(report.Failures, id+": this case of published "+key+" is missing. A case of a published version is never removed.")
			continue
		}
		if now := Hash(current, order); now != version.Cases[id] {
			report.Failures = append(report.Failures, id+": this case of published "+key+" has changed (recorded "+Short(version.Cases[id])+", now "+Short(now)+"). A published version is never edited. Ship the fix as the next version of "+nameOf(key)+".")
		}
	}
}

func nameOf(key string) string {
	if at := strings.Index(key, "@"); at >= 0 {
		return key[:at]
	}
	return key
}

func (g Guard) checkSources(report *Report, key string) {
	for _, c := range g.Cases[key] {
		if c.Source == "" {
			report.Failures = append(report.Failures, c.ID+": this case has no source id. A case with no source is not accepted.")
			continue
		}
		if !g.Sources[c.Source] {
			report.Failures = append(report.Failures, c.ID+": source "+c.Source+" has no row in documentation/1. requirements/sources.md.")
		}
	}
}

func (g Guard) checkNew(report *Report, key string) {
	version, recorded := g.Manifest[key]
	if !recorded || version.State != StatePublished {
		return
	}
	for _, c := range g.Cases[key] {
		if _, known := version.Cases[c.ID]; !known {
			report.Notices = append(report.Notices, c.ID+": new case of published "+key+" is not in the manifest yet.")
		}
	}
}

func (g Guard) checkPage(report *Report) {
	keys := map[string]bool{}
	for key := range g.Page {
		keys[key] = true
	}
	for key := range g.Cases {
		keys[key] = true
	}
	for _, key := range SortedKeys(keys) {
		inFile := idSet(g.Cases[key])
		inPage := map[string]bool{}
		for _, id := range g.Page[key] {
			inPage[id] = true
		}
		for _, id := range SortedKeys(inPage) {
			if !inFile[id] {
				report.Failures = append(report.Failures, id+" is on the page and not in "+key+".json.")
			}
		}
		for _, id := range SortedKeys(inFile) {
			if !inPage[id] {
				report.Failures = append(report.Failures, id+" is in "+key+".json and not on the page.")
			}
		}
	}
}

func idSet(cases []Case) map[string]bool {
	set := map[string]bool{}
	for _, c := range cases {
		set[c.ID] = true
	}
	return set
}

func ExtendManifest(manifest Manifest, versions map[string][]Case, served map[string][]string) (Manifest, []string, []string) {
	extended := Manifest{}
	for key, version := range manifest {
		copied := ManifestVersion{State: version.State, Cases: map[string]string{}}
		for id, hash := range version.Cases {
			copied.Cases[id] = hash
		}
		extended[key] = copied
	}
	var failures, added []string
	guard := Guard{Served: served, Cases: versions}
	for _, key := range SortedKeys(versions) {
		order := guard.orderOf(key, versions[key])
		version, known := extended[key]
		if !known {
			version = ManifestVersion{State: StateDraft, Cases: map[string]string{}}
		}
		for _, c := range versions[key] {
			now := Hash(c, order)
			recorded, has := version.Cases[c.ID]
			switch {
			case !has:
				version.Cases[c.ID] = now
				added = append(added, c.ID)
			case recorded != now:
				failures = append(failures, c.ID+": this case of "+key+" has changed (recorded "+Short(recorded)+", now "+Short(now)+"). The manifest never changes an entry.")
			}
		}
		extended[key] = version
	}
	sort.Strings(added)
	return extended, failures, added
}
