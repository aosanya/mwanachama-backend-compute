package laws

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strconv"

	"github.com/aosanya/mwanachama-backend-compute/formula"
)

const (
	NormalCases  = 500
	NightlyCases = 20000
)

type Engine struct {
	Parse  func(text string) (formula.Tree, []formula.Problem)
	Format func(tree formula.Tree) string
	Check  func(tree formula.Tree, inputs []formula.Input) []formula.Problem
	Hash   func(text string) string
	Run    func(tree formula.Tree, inputs []formula.Input, values map[string]formula.Value) string
}

func RealEngine() Engine {
	engine := Engine{
		Parse:  formula.Parse,
		Format: formula.Format,
		Check: func(tree formula.Tree, inputs []formula.Input) []formula.Problem {
			return formula.Check(tree, inputs, nil)
		},
		Run: runOutcome,
	}
	engine.Hash = func(text string) string {
		tree, problems := formula.Parse(text)
		if len(problems) > 0 {
			return "unreadable"
		}
		sum := sha256.Sum256([]byte(formula.Format(tree)))
		return hex.EncodeToString(sum[:])
	}
	return engine
}

func runOutcome(tree formula.Tree, inputs []formula.Input, values map[string]formula.Value) string {
	program, problems := formula.Plan(tree, inputs, nil, formula.DefaultOptions())
	if len(problems) > 0 {
		return "plan:" + problems[0].Code + ":" + problems[0].Sentence
	}
	value, problem := program.Run(values)
	if problem != nil {
		return "problem:" + problem.Code + ":" + problem.Sentence
	}
	return "value:" + string(value.Type()) + ":" + value.String()
}

func Cases() int {
	if text := os.Getenv("LAWS_CASES"); text != "" {
		if count, err := strconv.Atoi(text); err == nil && count > 0 {
			return count
		}
	}
	return NormalCases
}

func BaseSeed() int64 {
	if text := os.Getenv("LAWS_SEED"); text != "" {
		if seed, err := strconv.ParseInt(text, 10, 64); err == nil {
			return seed
		}
	}
	return 0
}
