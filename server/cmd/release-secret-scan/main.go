package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"cloud-clicker/server/releasepackage"
)

func main() {
	root := flag.String("root", "..", "repository root")
	archive := flag.String("gameserver-archive", "", "optional gameserver docker-save archive")
	flag.Parse()
	command := exec.Command("git", "ls-files", "-z")
	command.Dir = *root
	output, err := command.Output()
	if err != nil {
		fail(err)
	}
	parts := bytes.Split(output, []byte{0})
	paths := make([]string, 0, len(parts))
	for _, part := range parts {
		if len(part) != 0 {
			paths = append(paths, string(part))
		}
	}
	findings, err := releasepackage.ScanTrackedFiles(*root, paths)
	if err != nil {
		fail(err)
	}
	if strings.TrimSpace(*archive) != "" {
		imageFindings, err := releasepackage.ScanDockerArchive(*archive)
		if err != nil {
			fail(err)
		}
		findings = append(findings, imageFindings...)
	}
	if err := releasepackage.RequireNoSecrets(findings); err != nil {
		fail(err)
	}
	fmt.Printf("release secret scan: %d tracked files, no findings\n", len(paths))
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
