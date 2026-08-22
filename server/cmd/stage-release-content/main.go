package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"cloud-clicker/server/releasepackage"
)

func main() {
	root := flag.String("root", "..", "repository root")
	output := flag.String("output", "", "empty output directory for /opt/cloud-clicker/content")
	flag.Parse()
	if *output == "" {
		fail(releasepackage.ErrInvalidContent)
	}
	closure, err := releasepackage.StageRuntimeContent(*root, *output)
	if err != nil {
		fail(err)
	}
	encoded, err := json.Marshal(closure)
	if err != nil {
		fail(err)
	}
	fmt.Println(string(encoded))
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
