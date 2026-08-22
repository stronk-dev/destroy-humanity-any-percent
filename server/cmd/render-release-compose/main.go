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
	alertmanager := flag.String("alertmanager-image", "", "immutable Alertmanager image reference")
	caddy := flag.String("caddy-image", "", "immutable Caddy image reference")
	gameserver := flag.String("gameserver-image", "", "immutable gameserver image reference")
	nodeExporter := flag.String("node-exporter-image", "", "immutable node-exporter image reference")
	postgres := flag.String("postgres-image", "", "immutable Postgres image reference")
	prometheus := flag.String("prometheus-image", "", "immutable Prometheus image reference")
	flag.Parse()
	if *output == "" {
		fail(releasepackage.ErrInvalidContent)
	}
	template, err := os.ReadFile(*templatePath)
	if err != nil {
		fail(err)
	}
	rendered, err := releasepackage.RenderCompose(template, map[string]string{
		"alertmanager": *alertmanager, "caddy": *caddy, "gameserver": *gameserver, "node-exporter": *nodeExporter, "postgres": *postgres, "prometheus": *prometheus,
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
