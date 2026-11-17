package inputchecks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/function"
	"github.com/aosanya/mwanachama-backend-compute/problem"
)

const (
	emitSwitch = "INPUTCHECKS_EMIT"
	listsBegin = "BEGIN-LISTS"
	listsEnd   = "END-LISTS"
)

var entryIds = []string{"ic-c04", "ic-g03-s", "ic-h04", "ic-i04"}

func entryCases(t testing.TB) []testCase {
	t.Helper()
	byId := map[string]testCase{}
	for _, group := range allGroups() {
		for _, c := range group {
			byId[c.id] = c
		}
	}
	cases := make([]testCase, len(entryIds))
	for i, id := range entryIds {
		c, found := byId[id]
		if !found {
			t.Fatalf("the case %s is not in the suite", id)
		}
		cases[i] = c
	}
	return cases
}

func singleRun(c testCase) []problem.Problem {
	return function.Validate(c.def.body, c.row, &engine{})
}

func batchRow(c testCase, validator function.Validator) []problem.Problem {
	return validator.Validate(c.row, &engine{})
}

func toolCall(t *testing.T, c testCase) []problem.Problem {
	t.Helper()
	data, err := problem.ListJSON(singleRun(c))
	if err != nil {
		t.Fatal(err)
	}
	var decoded []problem.Problem
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("%s: the list does not read back from its JSON: %v", c.id, err)
	}
	return decoded
}

func listJSON(t testing.TB, problems []problem.Problem) string {
	t.Helper()
	data, err := problem.ListJSON(problems)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestIcM05TheThreeWaysInGiveOneProblemList(t *testing.T) {
	for _, c := range entryCases(t) {
		t.Run(c.id, func(t *testing.T) {
			single := singleRun(c)
			if !sameList(single, c.want) {
				t.Fatalf("the single run differs from the page\nwant:\n%s\ngot:\n%s", show(c.want), show(single))
			}
			validator := function.Prepare(c.def.body)
			for round := 1; round <= 3; round++ {
				if row := batchRow(c, validator); !sameList(row, single) {
					t.Errorf("batch row, use %d, differs from the single run\nsingle:\n%s\nrow:\n%s", round, show(single), show(row))
				}
			}
			if tool := toolCall(t, c); !sameList(tool, single) {
				t.Errorf("the list read back from its tool JSON differs from the single run\nsingle:\n%s\ntool:\n%s", show(single), show(tool))
			}
			if listJSON(t, batchRow(c, validator)) != listJSON(t, single) {
				t.Errorf("the batch row and the single run write different JSON")
			}
		})
	}
}

func TestIcM05TheWrappersOfABatchRowAndOfAToolAnswerAreOpen(t *testing.T) {
	t.Log("open question 6: no page gives the shape of a tool answer or of a batch row's position; only the problem list is compared")
	t.Log("the batch row here is Validator.Validate, the surface the function package gives a batch; the batch package is ISS-29")
	t.Log("the tool call here is the problem list written as JSON and read back; no tool call is built")
}

func TestIcM06TheListIsTheSameEachTimeItIsAsked(t *testing.T) {
	for _, c := range entryCases(t) {
		t.Run(c.id, func(t *testing.T) {
			first := listJSON(t, singleRun(c))
			second := listJSON(t, singleRun(c))
			if first != second {
				t.Fatalf("two calls in one process differ\nfirst:  %s\nsecond: %s", first, second)
			}
		})
	}
}

func TestEmitTheProblemListsForAnotherProcess(t *testing.T) {
	if os.Getenv(emitSwitch) == "" {
		t.Skip("this test only runs inside the second process of ic-m06")
	}
	fmt.Println(listsBegin)
	for _, c := range entryCases(t) {
		fmt.Println(c.id + " " + listJSON(t, singleRun(c)))
	}
	fmt.Println(listsEnd)
}

func TestIcM06TheListIsTheSameInASecondProcess(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command(executable, "-test.run=^TestEmitTheProblemListsForAnotherProcess$", "-test.v")
	command.Env = append(os.Environ(), emitSwitch+"=1")
	var output bytes.Buffer
	command.Stdout = &output
	command.Stderr = &output
	if err := command.Run(); err != nil {
		t.Fatalf("the second process failed: %v\n%s", err, output.String())
	}
	text := output.String()
	start := strings.Index(text, listsBegin)
	end := strings.Index(text, listsEnd)
	if start < 0 || end < start {
		t.Fatalf("the second process printed no lists:\n%s", text)
	}
	printed := map[string]string{}
	for _, entry := range strings.Split(strings.TrimSpace(text[start+len(listsBegin):end]), "\n") {
		id, list, found := strings.Cut(entry, " ")
		if found {
			printed[id] = list
		}
	}
	for _, c := range entryCases(t) {
		if printed[c.id] != listJSON(t, singleRun(c)) {
			t.Errorf("%s: the second process gave another list\nhere:  %s\nother: %s", c.id, listJSON(t, singleRun(c)), printed[c.id])
		}
	}
	if len(printed) != len(entryIds) {
		t.Errorf("the second process printed %d lists, want %d", len(printed), len(entryIds))
	}
}
