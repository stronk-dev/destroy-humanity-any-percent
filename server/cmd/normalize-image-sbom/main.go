package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"cloud-clicker/server/releasepackage"
)

func main() {
	input := flag.String("input", "", "raw Syft SPDX JSON")
	output := flag.String("output", "", "new normalized SPDX JSON path")
	identity := flag.String("runtime-config-sha256", "", "exact OCI runtime config digest")
	created := flag.String("created", "", "release creation time in RFC3339")
	flag.Parse()
	createdAt, err := time.Parse(time.RFC3339, *created)
	if err != nil || flag.NArg() != 0 || *input == "" || *output == "" || *identity == "" {
		fail(releasepackage.ErrInvalidContent)
	}
	if err := releasepackage.WriteNormalizedImageSPDX(*input, *output, *identity, createdAt); err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "image SBOM normalization failed: %v\n", err)
	os.Exit(1)
}
