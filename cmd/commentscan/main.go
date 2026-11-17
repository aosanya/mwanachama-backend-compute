package main

import (
	"fmt"
	"os"
)

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}
	findings, err := Walk(root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	for _, f := range findings {
		fmt.Printf("%s:%d: %s\n", f.File, f.Line, f.Text)
	}
	if len(findings) > 0 {
		fmt.Fprintf(os.Stderr, "%d comment lines found\n", len(findings))
		os.Exit(1)
	}
}
