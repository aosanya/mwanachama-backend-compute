package basicruns

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/aosanya/mwanachama-backend-compute/formula"
)

type Held struct {
	Case   Case
	Reason string
}

type Resolution struct {
	Run       []Case
	Held      []Held
	Gaps      []Case
	Missing   []string
	NoRowFor  []string
	Aliased   []string
	Directory string
}

func walk(node formula.Node, visit func(formula.Node)) {
	visit(node)
	for _, child := range node.Children() {
		walk(child, visit)
	}
}

func callsOf(text string) (agencySteps bool) {
	tree, problems := formula.Parse(text)
	if len(problems) != 0 {
		return false
	}
	walk(tree.Root, func(n formula.Node) {
		call, isCall := n.(formula.Call)
		if !isCall {
			return
		}
		if call.Library == nil && strings.HasPrefix(call.Name, "step_") {
			agencySteps = true
		}
	})
	return agencySteps
}

func heldReason(c Case) string {
	steps := callsOf(c.Formula)
	asksResult := strings.HasPrefix(c.Expect, "result") || c.Expect == "accepted"
	switch {
	case strings.HasPrefix(c.Name, "rules-"):
		return "reads text into values: Validate, ISS-26"
	case steps && asksResult:
		return "Gap 9: step_a to step_e have no definition"
	}
	return ""
}

func otherPagesMention(name string) bool {
	files, err := filepath.Glob(qaFolder + "*.md")
	if err != nil {
		return false
	}
	for _, file := range files {
		if filepath.Base(file) == "basic-runs.md" || filepath.Base(file) == "formula-examples.md" {
			continue
		}
		data, err := os.ReadFile(file)
		if err == nil && strings.Contains(string(data), name) {
			return true
		}
	}
	return false
}

func resolve(page Page, examples map[string]Case) Resolution {
	var resolution Resolution
	own := map[string]Case{}
	for _, c := range page.Cases {
		if c.Kind == KindOwn {
			own[c.Name] = c
		}
	}
	place := func(c Case) {
		if reason := heldReason(c); reason != "" {
			resolution.Held = append(resolution.Held, Held{Case: c, Reason: reason})
			return
		}
		resolution.Run = append(resolution.Run, c)
	}
	seen := map[string]bool{}
	for _, c := range page.Cases {
		switch c.Kind {
		case KindNone:
			resolution.Gaps = append(resolution.Gaps, c)
		case KindHeld:
			resolution.Held = append(resolution.Held, Held{Case: c, Reason: "setting not decided: " + c.Note})
		case KindOwn:
			if !seen[c.Name] {
				seen[c.Name] = true
				place(c)
			}
		case KindAlias:
			if _, found := own[c.Name]; !found {
				resolution.Missing = append(resolution.Missing, c.Name)
			}
			resolution.Aliased = append(resolution.Aliased, c.Name)
		case KindElsewhere:
			if seen[c.Name] {
				continue
			}
			seen[c.Name] = true
			example, found := examples[c.Name]
			switch {
			case found:
				example.Section = c.Section
				example.Part = c.Part
				place(example)
			case otherPagesMention(c.Name):
				resolution.Held = append(resolution.Held, Held{Case: c, Reason: "the case is on another QA page with no row of formula-examples.md to read; the library versioning cases run in library/ (ISS-28)"})
			default:
				resolution.NoRowFor = append(resolution.NoRowFor, c.Name)
			}
		}
	}
	return resolution
}
