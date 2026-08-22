package main

import (
	"flag"
	"fmt"
	"os"

	"cloud-clicker/server/deploymentrehearsal"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] != "validate" {
		fail("usage")
	}
	validated, err := runValidate(os.Args[2:])
	if err != nil {
		fail("invalid_evidence")
	}
	fmt.Printf("R-006 evidence valid: run=%s manifest=%s\n", validated.RunID, validated.ManifestSHA256)
}

func runValidate(args []string) (deploymentrehearsal.Evidence, error) {
	set := flag.NewFlagSet("validate", flag.ContinueOnError)
	evidence := set.String("evidence", "", "exact R-006 evidence JSON")
	if set.Parse(args) != nil || set.NArg() != 0 || *evidence == "" {
		return deploymentrehearsal.Evidence{}, deploymentrehearsal.ErrInvalid
	}
	validated, err := deploymentrehearsal.Load(*evidence)
	if err != nil {
		return deploymentrehearsal.Evidence{}, err
	}
	return validated, nil
}

func fail(class string) {
	if class != "usage" && class != "invalid_input" && class != "invalid_evidence" {
		class = "operation_failed"
	}
	fmt.Fprintf(os.Stderr, "deployment rehearsal failed: error_class=%s\n", class)
	os.Exit(1)
}
