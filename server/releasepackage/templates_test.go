package releasepackage

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProductionTemplatesBindPrivateTopologyAndExactRoutes(t *testing.T) {
	root := filepath.Join("..", "..")
	if err := ValidateDeploymentTemplates(root); err != nil {
		t.Fatal(err)
	}
	template, err := os.ReadFile(filepath.Join(root, "deployment", "compose.template.yml"))
	if err != nil {
		t.Fatal(err)
	}
	rendered, err := RenderCompose(template, fixtureImages())
	if err != nil || bytes.Contains(rendered, []byte("@@")) {
		t.Fatalf("render err=%v\n%s", err, rendered)
	}
}

func TestComposeRejectsMutableImagesAndPublishedPrivatePorts(t *testing.T) {
	template, err := os.ReadFile(filepath.Join("..", "..", "deployment", "compose.template.yml"))
	if err != nil {
		t.Fatal(err)
	}
	mutable := fixtureImages()
	mutable["postgres"] = "postgres:16-alpine"
	if _, err := RenderCompose(template, mutable); !errors.Is(err, ErrInvalidContent) {
		t.Fatalf("mutable image accepted: %v", err)
	}
	rendered, err := RenderCompose(template, fixtureImages())
	if err != nil {
		t.Fatal(err)
	}
	severed := bytes.Replace(rendered, []byte("    expose: [\"8080\"]"), []byte("    ports: [\"8080:8080\"]"), 1)
	if err := ValidateCompose(severed); !errors.Is(err, ErrInvalidContent) {
		t.Fatalf("published gameserver port accepted: %v", err)
	}
}

func TestComposeRejectsRootOrNonSeparateBackupWorker(t *testing.T) {
	template, err := os.ReadFile(filepath.Join("..", "..", "deployment", "compose.template.yml"))
	if err != nil {
		t.Fatal(err)
	}
	rendered, err := RenderCompose(template, fixtureImages())
	if err != nil {
		t.Fatal(err)
	}
	root := bytes.Replace(rendered, []byte("    user: \"70:70\""), []byte("    user: \"0:0\""), 1)
	if err := ValidateCompose(root); !errors.Is(err, ErrInvalidContent) {
		t.Fatalf("root backup worker accepted: %v", err)
	}
	sharedVolume := bytes.Replace(rendered,
		[]byte("${CLOUD_CLICKER_BACKUP_TARGET:?set the separately mounted backup target}:/backups"),
		[]byte("postgres_data:/backups"), 1)
	if err := ValidateCompose(sharedVolume); !errors.Is(err, ErrInvalidContent) {
		t.Fatalf("database volume accepted as backup target: %v", err)
	}
}

func TestCaddyRejectsMissingWebSocketRouteAndPublicMetrics(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "deployment", "Caddyfile"))
	if err != nil {
		t.Fatal(err)
	}
	withoutSocket := bytes.Replace(data, []byte(" /connection/websocket"), nil, 1)
	if err := ValidateCaddyfile(withoutSocket); !errors.Is(err, ErrInvalidContent) {
		t.Fatalf("missing WebSocket route accepted: %v", err)
	}
	withMetrics := append(append([]byte(nil), data...), []byte("\n# /metrics\n")...)
	if err := ValidateCaddyfile(withMetrics); !errors.Is(err, ErrInvalidContent) {
		t.Fatalf("public metrics route accepted: %v", err)
	}
}

func TestGameserverDockerfileRejectsMutableFrontendAndRootUser(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "deployment", "Dockerfile.gameserver"))
	if err != nil {
		t.Fatal(err)
	}
	firstNewline := bytes.IndexByte(data, '\n')
	mutable := append([]byte("# syntax=docker/dockerfile:1.7\n"), data[firstNewline+1:]...)
	if err := ValidateGameserverDockerfile(mutable); !errors.Is(err, ErrInvalidContent) {
		t.Fatalf("mutable Dockerfile frontend accepted: %v", err)
	}
	root := bytes.Replace(data, []byte("USER 65532:65532"), []byte("USER 0:0"), 1)
	if err := ValidateGameserverDockerfile(root); !errors.Is(err, ErrInvalidContent) {
		t.Fatalf("root gameserver image accepted: %v", err)
	}
}

func fixtureImages() map[string]string {
	return map[string]string{
		"caddy":      "caddy:2-alpine@sha256:" + strings.Repeat("a", 64),
		"gameserver": "cloud-clicker/gameserver:v0.1.0@sha256:" + strings.Repeat("b", 64),
		"postgres":   "postgres:16-alpine@sha256:" + strings.Repeat("c", 64),
	}
}
