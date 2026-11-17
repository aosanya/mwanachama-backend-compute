package fit

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"

	celast "cel.dev/cel-go/common/ast"
	"github.com/expr-lang/expr/ast"

	"github.com/aosanya/mwanachama-backend-compute/spike/engine/cel"
	"github.com/aosanya/mwanachama-backend-compute/spike/engine/exprlang"
	"github.com/aosanya/mwanachama-backend-compute/spike/engine/own"
	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
)

type subject struct {
	name    string
	engine  harness.TreeEngine
	convert func(any) (Node, error)
}

func subjects(t testing.TB) []subject {
	t.Helper()
	all := []struct {
		engine  harness.Engine
		convert func(any) (Node, error)
	}{
		{cel.New(), FromCEL},
		{exprlang.New(), FromExprLang},
		{own.New(), FromOwn},
	}
	var found []subject
	for _, entry := range all {
		tree, ok := entry.engine.(harness.TreeEngine)
		if !ok {
			t.Fatalf("%s does not implement harness.TreeEngine", entry.engine.Name())
		}
		found = append(found, subject{entry.engine.Name(), tree, entry.convert})
	}
	return found
}

func (s subject) neutral(source string) (Node, error) {
	raw, err := s.engine.Tree(harness.Case{Source: source})
	if err != nil {
		return Node{}, err
	}
	return s.convert(raw)
}

func (s subject) printed(source string) (string, error) {
	return s.engine.Print(harness.Case{Source: source})
}

func sameTree(a, b Node) bool {
	return reflect.DeepEqual(a, b)
}

var (
	sectionLock sync.Mutex
	sections    = map[string][]string{}
)

func record(section, format string, args ...any) {
	sectionLock.Lock()
	defer sectionLock.Unlock()
	sections[section] = append(sections[section], fmt.Sprintf(format, args...))
}

func yes(ok bool) string {
	if ok {
		return "yes"
	}
	return "no"
}

func rawLabel(raw any) string {
	switch v := raw.(type) {
	case celast.Expr:
		return celLabel(v)
	case ast.Node:
		return strings.TrimPrefix(fmt.Sprintf("%T", v), "*ast.")
	case own.Node:
		return strings.TrimPrefix(fmt.Sprintf("%T", v), "own.")
	}
	return fmt.Sprintf("%T", raw)
}

func celLabel(e celast.Expr) string {
	switch e.Kind() {
	case celast.CallKind:
		return "Call " + e.AsCall().FunctionName()
	case celast.LiteralKind:
		return "Literal " + e.AsLiteral().Type().TypeName()
	case celast.IdentKind:
		return "Ident"
	case celast.ListKind:
		return "List"
	}
	return fmt.Sprintf("Kind%d", e.Kind())
}

func rawKinds(raw any) []string {
	seen := map[string]bool{}
	switch v := raw.(type) {
	case celast.Expr:
		celast.PreOrderVisit(v, celast.NewExprVisitor(func(e celast.Expr) { seen[celLabel(e)] = true }))
	case ast.Node:
		node := v
		ast.Walk(&node, kindVisitor(seen))
	case own.Node:
		collectOwn(v, seen)
	}
	kinds := make([]string, 0, len(seen))
	for kind := range seen {
		kinds = append(kinds, kind)
	}
	sort.Strings(kinds)
	return kinds
}

type kindVisitor map[string]bool

func (k kindVisitor) Visit(node *ast.Node) {
	k[rawLabel(*node)] = true
}

func collectOwn(n own.Node, seen map[string]bool) {
	seen[rawLabel(n)] = true
	for _, child := range own.Children(n) {
		collectOwn(child, seen)
	}
}

func rawCount(raw any) int {
	total := 0
	switch v := raw.(type) {
	case celast.Expr:
		celast.PreOrderVisit(v, celast.NewExprVisitor(func(celast.Expr) { total++ }))
	case ast.Node:
		node := v
		ast.Walk(&node, counter{&total})
	case own.Node:
		total = own.Count(v)
	}
	return total
}

type counter struct{ total *int }

func (c counter) Visit(*ast.Node) { *c.total++ }
