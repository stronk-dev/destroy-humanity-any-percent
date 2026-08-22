package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"cloud-clicker/server/releasepackage"
)

func main() {
	templatePath := flag.String("template", "../deployment/compose.template.yml", "Compose template")
	output := flag.String("output", "", "new rendered Compose path")
	caddy := flag.String("caddy-image", "", "immutable Caddy image reference")
	gameserver := flag.String("gameserver-image", "", "immutable gameserver image reference")
	postgres := flag.String("postgres-image", "", "immutable Postgres image reference")
	flag.Parse()
	if *output == "" {
		fail(releasepackage.ErrInvalidContent)
	}
	template, err := os.ReadFile(*templatePath)
	if err != nil {
		fail(err)
	}
	rendered, err := releasepackage.RenderCompose(template, map[string]string{
		"caddy": *caddy, "gameserver": *gameserver, "postgres": *postgres,
	})
	if err != nil {
		fail(err)
	}
	if _, err := os.Stat(*output); !os.IsNotExist(err) {
		fail(releasepackage.ErrInvalidContent)
	}
	if err := os.MkdirAll(filepath.Dir(*output), 0o755); err != nil {
		fail(err)
	}
	if err := os.WriteFile(*output, rendered, 0o644); err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
