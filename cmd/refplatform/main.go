package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/aosanya/mwanachama-backend-compute/functions"
)

func main() {
	dir := flag.String("dir", "functions/reference", "the folder of the recorded cases and the platform results")
	record := flag.Bool("record", false, "write the results of this platform as the recorded ones")
	flag.Parse()
	versions, err := functions.LoadCases(*dir)
	if err != nil {
		fail(err)
	}
	got, err := functions.Replay(versions)
	if err != nil {
		fail(err)
	}
	if *record {
		if err := got.Write(*dir); err != nil {
			fail(err)
		}
		fmt.Printf("%d results recorded on %s\n", len(got.Results), got.Name)
		return
	}
	recorded, err := functions.LoadPlatform(*dir)
	if err != nil {
		fail(err)
	}
	failures := functions.ComparePlatforms(recorded, got, versions)
	for _, failure := range failures {
		fmt.Fprintln(os.Stderr, failure)
	}
	if len(failures) > 0 {
		os.Exit(1)
	}
	fmt.Printf("%d results on %s agree with those recorded on %s\n", len(got.Results), got.Name, recorded.Name)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
