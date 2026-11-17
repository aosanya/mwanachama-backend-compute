package function

import (
	"sort"

	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/numtext"
	"github.com/aosanya/mwanachama-backend-compute/problem"
)

const noRunnerSentence = "No way to run this check was given."

type preparedCheck struct {
	decl         CheckDecl
	tree         formula.Tree
	standardText string
	reads        []string
	unreadable   string
}

type Validator struct {
	definition Definition
	checks     []preparedCheck
}

func Prepare(definition Definition) Validator {
	checks := make([]preparedCheck, len(definition.Checks))
	for i, decl := range definition.Checks {
		checks[i] = prepareCheck(decl)
	}
	return Validator{definition: definition, checks: checks}
}

func Validate(definition Definition, row Row, runner Runner) []problem.Problem {
	return Prepare(definition).Validate(row, runner)
}

func prepareCheck(decl CheckDecl) preparedCheck {
	tree, problems := formula.Parse(decl.Formula)
	if len(problems) > 0 {
		return preparedCheck{decl: decl, standardText: decl.Formula, unreadable: problems[0].Sentence}
	}
	return preparedCheck{decl: decl, tree: tree, standardText: formula.Format(tree), reads: namesRead(tree.Root)}
}

func namesRead(root formula.Node) []string {
	var names []string
	seen := map[string]bool{}
	var walk func(node formula.Node)
	walk = func(node formula.Node) {
		if name, isName := node.(formula.Name); isName {
			if !seen[name.Name] {
				seen[name.Name] = true
				names = append(names, name.Name)
			}
			return
		}
		for _, child := range node.Children() {
			walk(child)
		}
	}
	walk(root)
	return names
}

func (v Validator) Validate(row Row, runner Runner) []problem.Problem {
	problems := []problem.Problem{}
	failed := map[string]bool{}
	values := map[string]numtext.Value{}
	for _, decl := range v.definition.Inputs {
		received, given := row[decl.Name]
		outcome := validateInput(decl, received, given)
		problems = append(problems, outcome.problems...)
		failed[decl.Name] = outcome.failed()
		values[decl.Name] = outcome.value
	}
	problems = append(problems, v.unknownInputs(row)...)
	for _, prepared := range v.checks {
		problems = append(problems, v.runCheck(prepared, row, failed, values, runner)...)
	}
	return problems
}

func (v Validator) unknownInputs(row Row) []problem.Problem {
	declared := make([]string, len(v.definition.Inputs))
	known := map[string]bool{}
	for i, decl := range v.definition.Inputs {
		declared[i] = decl.Name
		known[decl.Name] = true
	}
	var names []string
	for name := range row {
		if !known[name] {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	problems := make([]problem.Problem, len(names))
	for i, name := range names {
		problems[i] = problem.UnknownInput(name, shownReceived(row[name]), declared)
	}
	return problems
}

func (v Validator) runCheck(prepared preparedCheck, row Row, failed map[string]bool, values map[string]numtext.Value, runner Runner) []problem.Problem {
	message := prepared.decl.Message
	if prepared.unreadable != "" {
		return []problem.Problem{problem.CheckError(prepared.standardText, message, prepared.unreadable)}
	}
	for _, name := range prepared.reads {
		if failed[name] {
			return nil
		}
	}
	if runner == nil {
		return []problem.Problem{problem.CheckError(prepared.standardText, message, noRunnerSentence)}
	}
	outcome := runner.Run(prepared.tree, values)
	if outcome.Failure != "" {
		return []problem.Problem{problem.CheckError(prepared.standardText, message, outcome.Failure)}
	}
	if outcome.True {
		return nil
	}
	return []problem.Problem{problem.CheckFailed(prepared.standardText, message, v.namedValues(prepared.reads, row))}
}

func (v Validator) namedValues(names []string, row Row) []problem.NamedValue {
	named := make([]problem.NamedValue, len(names))
	for i, name := range names {
		named[i] = problem.NamedValue{Name: name, Value: v.shownInput(name, row[name])}
	}
	return named
}

func (v Validator) shownInput(name string, received numtext.Received) problem.Value {
	for _, decl := range v.definition.Inputs {
		if decl.Name != name {
			continue
		}
		if info, known := infoOf(decl.Type); known && info.list {
			return problem.List(received.Items)
		}
		return shownSingle(decl.Type, received)
	}
	return shownReceived(received)
}

func shownReceived(received numtext.Received) problem.Value {
	if received.List {
		return problem.List(received.Items)
	}
	return problem.PlainText(received.Text)
}
