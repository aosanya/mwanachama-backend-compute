package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
)

const rowsPerBatch = 10000

type sample struct {
	nsPerOp float64
	bytes   float64
	allocs  float64
}

type key struct {
	group   string
	engine  string
	formula string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: summarise bench-run.txt [more-bench-output.txt ...]")
		os.Exit(2)
	}
	samples := map[key][]sample{}
	for _, path := range os.Args[1:] {
		if err := read(path, samples); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	out := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for _, group := range []string{"Run", "Compile", "BatchSequential", "BatchParallel"} {
		write(out, group, samples)
	}
	out.Flush()
}

func read(path string, samples map[key][]sample) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		k, s, ok := parse(scanner.Text())
		if ok {
			samples[k] = append(samples[k], s)
		}
	}
	return scanner.Err()
}

func parse(line string) (key, sample, bool) {
	fields := strings.Fields(line)
	if len(fields) < 4 || !strings.HasPrefix(fields[0], "Benchmark") {
		return key{}, sample{}, false
	}
	name := strings.TrimPrefix(fields[0], "Benchmark")
	if cut := strings.LastIndex(name, "-"); cut > 0 {
		name = name[:cut]
	}
	parts := strings.Split(name, "/")
	if len(parts) != 3 {
		return key{}, sample{}, false
	}
	var s sample
	for i := 2; i < len(fields); i++ {
		value, err := strconv.ParseFloat(fields[i-1], 64)
		if err != nil {
			continue
		}
		switch fields[i] {
		case "ns/op":
			s.nsPerOp = value
		case "B/op":
			s.bytes = value
		case "allocs/op":
			s.allocs = value
		}
	}
	if s.nsPerOp == 0 {
		return key{}, sample{}, false
	}
	return key{parts[0], parts[1], parts[2]}, s, true
}

func write(out *tabwriter.Writer, group string, samples map[key][]sample) {
	keys := keysOf(group, samples)
	if len(keys) == 0 {
		return
	}
	batch := strings.HasPrefix(group, "Batch")
	fmt.Fprintf(out, "\n%s\n", heading(group))
	if batch {
		fmt.Fprintln(out, "engine\tformula\truns\tns/batch median\tns/row\tspread %\tB/batch\tallocs/batch")
	} else {
		fmt.Fprintln(out, "engine\tformula\truns\tns/op median\tspread %\tB/op\tallocs/op")
	}
	for _, k := range keys {
		runs := samples[k]
		ns := column(runs, func(s sample) float64 { return s.nsPerOp })
		bytes := median(column(runs, func(s sample) float64 { return s.bytes }))
		allocs := median(column(runs, func(s sample) float64 { return s.allocs }))
		mid := median(ns)
		spread := 100 * (ns[len(ns)-1] - ns[0]) / mid
		if batch {
			fmt.Fprintf(out, "%s\t%s\t%d\t%.0f\t%.1f\t%.1f\t%.0f\t%.0f\n", k.engine, k.formula, len(runs), mid, mid/rowsPerBatch, spread, bytes, allocs)
			continue
		}
		fmt.Fprintf(out, "%s\t%s\t%d\t%.0f\t%.1f\t%.0f\t%.0f\n", k.engine, k.formula, len(runs), mid, spread, bytes, allocs)
	}
}

func heading(group string) string {
	switch group {
	case "Run":
		return "Per run, one compiled program, first example inputs"
	case "Compile":
		return "Per compile, parse and check and build one program"
	case "BatchSequential":
		return fmt.Sprintf("Batch of %d rows, one goroutine", rowsPerBatch)
	}
	return fmt.Sprintf("Batch of %d rows, 8 workers splitting the rows", rowsPerBatch)
}

func keysOf(group string, samples map[key][]sample) []key {
	var keys []key
	for k := range samples {
		if k.group == group {
			keys = append(keys, k)
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].engine != keys[j].engine {
			return keys[i].engine < keys[j].engine
		}
		return keys[i].formula < keys[j].formula
	})
	return keys
}

func column(runs []sample, pick func(sample) float64) []float64 {
	values := make([]float64, len(runs))
	for i, s := range runs {
		values[i] = pick(s)
	}
	sort.Float64s(values)
	return values
}

func median(sorted []float64) float64 {
	n := len(sorted)
	if n%2 == 1 {
		return sorted[n/2]
	}
	return (sorted[n/2-1] + sorted[n/2]) / 2
}
