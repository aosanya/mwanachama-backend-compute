package examplecases

import (
	"fmt"
	"os"
	"strings"
)

const heldHeader = "| name | reason |"

func LoadHeld(path string) (map[string]string, error) {
	text, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	held := map[string]string{}
	started := false
	for index, line := range strings.Split(string(text), "\n") {
		line = strings.TrimRight(line, " \t")
		switch {
		case line == heldHeader:
			started = true
		case !started || line == "" || strings.HasPrefix(line, "|---"):
		default:
			cells, ok := splitCells(line)
			if !ok || len(cells) != 2 || cells[0] == "" || cells[1] == "" {
				return nil, fmt.Errorf("%s:%d: a held row is | name | reason |", path, index+1)
			}
			if _, twice := held[cells[0]]; twice {
				return nil, fmt.Errorf("%s:%d: duplicate name: %q is held twice", path, index+1, cells[0])
			}
			held[cells[0]] = cells[1]
		}
	}
	if !started {
		return nil, fmt.Errorf("%s:1: bad header: the table must start with %s", path, heldHeader)
	}
	return held, nil
}
