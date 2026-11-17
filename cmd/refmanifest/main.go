package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/aosanya/mwanachama-backend-compute/functions"
)

func main() {
	dir := flag.String("dir", "functions/reference", "the folder of the recorded cases and the manifest")
	flag.Parse()
	versions, err := functions.LoadCases(*dir)
	if err != nil {
		fail(err)
	}
	manifest, err := functions.LoadManifest(*dir)
	if err != nil {
		fail(err)
	}
	served := map[string][]string{}
	for _, candidate := range functions.All() {
		served[candidate.Key()] = candidate.ParamOrder()
	}
	extended, failures, added := functions.ExtendManifest(manifest, versions, served)
	for _, failure := range failures {
		fmt.Fprintln(os.Stderr, failure)
	}
	if len(failures) > 0 {
		os.Exit(1)
	}
	if err := extended.Write(*dir); err != nil {
		fail(err)
	}
	fmt.Printf("%d cases added to the manifest\n", len(added))
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
