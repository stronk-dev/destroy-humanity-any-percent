package releasepackage

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"cloud-clicker/server/deploymentconfig"

	yaml "go.yaml.in/yaml/v2"
)

var imageReferencePattern = regexp.MustCompile(`^(?:[a-z0-9][a-z0-9./_-]*(?::[A-Za-z0-9._-]+)?@)?sha256:[0-9a-f]{64}$`)

type composeModel struct {
	Services map[string]composeService `yaml:"services"`
	Networks map[string]composeNetwork `yaml:"networks"`
	Secrets  map[string]composeSecret  `yaml:"secrets"`
}

type composeService struct {
	Image       string            `yaml:"image"`
	User        string            `yaml:"user"`
	ReadOnly    bool              `yaml:"read_only"`
	Entrypoint  []string          `yaml:"entrypoint"`
	Command     []string          `yaml:"command"`
	CapDrop     []string          `yaml:"cap_drop"`
	SecurityOpt []string          `yaml:"security_opt"`
	Tmpfs       []string          `yaml:"tmpfs"`
	Ports       []string          `yaml:"ports"`
	Expose      []string          `yaml:"expose"`
	Environment map[string]string `yaml:"environment"`
	Networks    []string          `yaml:"networks"`
	Secrets     []string          `yaml:"secrets"`
	Volumes     []string          `yaml:"volumes"`
	DependsOn   map[string]struct {
		Condition string `yaml:"condition"`
	} `yaml:"depends_on"`
}

type composeNetwork struct {
	Internal bool `yaml:"internal"`
}

type composeSecret struct {
	File string `yaml:"file"`
}

func RenderCompose(template []byte, images map[string]string) ([]byte, error) {
	if len(template) == 0 || len(images) != 3 {
		return nil, ErrInvalidContent
	}
	result := append([]byte(nil), template...)
	for _, name := range []string{"CADDY", "GAMESERVER", "POSTGRES"} {
		reference := images[strings.ToLower(name)]
		if !imageReferencePattern.MatchString(reference) {
			return nil, fmt.Errorf("%w: %s image is not immutable", ErrInvalidContent, strings.ToLower(name))
		}
		token := []byte("@@" + name + "_IMAGE@@")
		if bytes.Count(result, token) != 1 {
			return nil, fmt.Errorf("%w: image token %s", ErrInvalidContent, token)
		}
		result = bytes.Replace(result, token, []byte(reference), 1)
	}
	if bytes.Contains(result, []byte("@@")) {
		return nil, ErrInvalidContent
	}
	if err := ValidateCompose(result); err != nil {
		return nil, err
	}
	return result, nil
}

func ValidateCompose(data []byte) error {
	var model composeModel
	if len(data) == 0 || yaml.Unmarshal(data, &model) != nil || len(model.Services) != 4 {
		return ErrInvalidContent
	}
	wantServices := []string{"backup", "caddy", "gameserver", "postgres"}
	for _, name := range wantServices {
		service, ok := model.Services[name]
		if !ok || !imageReferencePattern.MatchString(service.Image) {
			return fmt.Errorf("%w: service %s", ErrInvalidContent, name)
		}
		if name != "caddy" && len(service.Ports) != 0 {
			return fmt.Errorf("%w: non-Caddy port publication", ErrInvalidContent)
		}
	}
	if strings.HasPrefix(model.Services["caddy"].Image, "sha256:") || strings.HasPrefix(model.Services["postgres"].Image, "sha256:") {
		return fmt.Errorf("%w: upstream image requires repository digest", ErrInvalidContent)
	}
	caddy := model.Services["caddy"]
	gameserver := model.Services["gameserver"]
	postgres := model.Services["postgres"]
	backup := model.Services["backup"]
	if len(caddy.Ports) != 2 || !sameStrings(caddy.Networks, []string{"application", "edge"}) ||
		!sameStrings(gameserver.Networks, []string{"application", "database"}) || !sameStrings(postgres.Networks, []string{"database"}) || !sameStrings(backup.Networks, []string{"database"}) ||
		len(gameserver.Expose) != 1 || gameserver.Expose[0] != "8080" || len(postgres.Expose) != 0 || len(backup.Expose) != 0 {
		return fmt.Errorf("%w: invalid service topology caddy_ports=%v caddy_networks=%v gameserver_expose=%v gameserver_networks=%v postgres_expose=%v postgres_networks=%v",
			ErrInvalidContent, caddy.Ports, caddy.Networks, gameserver.Expose, gameserver.Networks, postgres.Expose, postgres.Networks)
	}
	if !model.Networks["application"].Internal || !model.Networks["database"].Internal || model.Networks["edge"].Internal {
		return fmt.Errorf("%w: invalid network privacy %+v", ErrInvalidContent, model.Networks)
	}
	wantEnvironment := map[string]string{
		"CLOUD_CLICKER_DEPLOYMENT_MODE":            "production",
		"CLOUD_CLICKER_TRUSTED_PROXY_HOPS":         "1",
		"CLOUD_CLICKER_CONTENT_ROOT":               deploymentconfig.ProductionContentRoot,
		"DATABASE_URL_FILE":                        "/run/secrets/database-url",
		"CLOUD_CLICKER_JWT_CURRENT_KEY_FILE":       "/run/secrets/jwt-current",
		"CLOUD_CLICKER_BOOTSTRAP_CURRENT_KEY_FILE": "/run/secrets/bootstrap-current",
	}
	for name, want := range wantEnvironment {
		if gameserver.Environment[name] != want {
			return fmt.Errorf("%w: gameserver environment %s", ErrInvalidContent, name)
		}
	}
	if !sameStrings(gameserver.Secrets, []string{"bootstrap-current", "database-url", "jwt-current"}) ||
		!sameStrings(postgres.Secrets, []string{"postgres-password"}) || !sameStrings(backup.Secrets, []string{"database-url"}) {
		return fmt.Errorf("%w: invalid secret mounts gameserver=%v postgres=%v", ErrInvalidContent, gameserver.Secrets, postgres.Secrets)
	}
	wantBackupCommand := []string{"--age-recipient=${CLOUD_CLICKER_AGE_RECIPIENT:?set the public age X25519 recipient}", "--database-url-file=/run/secrets/database-url", "--epoch=/opt/cloud-clicker/epoch.json", "--release-manifest=/opt/cloud-clicker/release-manifest.json", "--server-id=${CLOUD_CLICKER_SERVER_ID:?set a canonical UUID}", "--target=/backups", "schedule"}
	if backup.Image != postgres.Image || backup.User != "70:70" || !backup.ReadOnly || !sameStrings(backup.CapDrop, []string{"ALL"}) || !sameStrings(backup.SecurityOpt, []string{"no-new-privileges:true"}) || len(backup.Tmpfs) != 1 ||
		!sameStrings(backup.Entrypoint, []string{"/opt/cloud-clicker/deployment-backup"}) || !sameStrings(backup.Command, wantBackupCommand) || backup.DependsOn["postgres"].Condition != "service_healthy" || len(backup.DependsOn) != 1 ||
		!sameStrings(backup.Volumes, []string{"./content/balance/epochs/phase0.json:/opt/cloud-clicker/epoch.json:ro", "./deployment-backup:/opt/cloud-clicker/deployment-backup:ro", "./release-manifest.json:/opt/cloud-clicker/release-manifest.json:ro", "${CLOUD_CLICKER_BACKUP_TARGET:?set the separately mounted backup target}:/backups"}) {
		return fmt.Errorf("%w: invalid backup worker boundary", ErrInvalidContent)
	}
	return nil
}

func ValidateCaddyfile(data []byte) error {
	text := string(data)
	for _, required := range []string{"{$CLOUD_CLICKER_PUBLIC_ORIGIN}", "/api/*", "/connection/websocket", "/healthz", "/readyz", "reverse_proxy gameserver:8080", "root * /srv", "try_files {path} /index.html"} {
		if strings.Count(text, required) != 1 {
			return fmt.Errorf("%w: Caddy route %q", ErrInvalidContent, required)
		}
	}
	for _, forbidden := range []string{"/metrics", "prometheus", "localhost"} {
		if strings.Contains(strings.ToLower(text), forbidden) {
			return fmt.Errorf("%w: public Caddy config contains %q", ErrInvalidContent, forbidden)
		}
	}
	return nil
}

func ValidateGameserverDockerfile(data []byte) error {
	text := string(data)
	if !strings.HasPrefix(text, "# syntax=docker/dockerfile:1.7@sha256:") || strings.Count(text, "ARG SOURCE_DATE_EPOCH") != 1 || strings.Count(text, "\nFROM scratch\n") != 1 ||
		strings.Count(text, "USER 65532:65532") != 1 || strings.Count(text, "STOPSIGNAL SIGTERM") != 1 ||
		strings.Count(text, "ENTRYPOINT [\"/usr/local/bin/gameserver\"]") != 1 || strings.Contains(text, "COPY . ") {
		return ErrInvalidContent
	}
	return nil
}

func ValidateDeploymentTemplates(root string) error {
	template, err := os.ReadFile(filepath.Join(root, "deployment", "compose.template.yml"))
	if err != nil {
		return err
	}
	images := map[string]string{
		"caddy":      "caddy:fixture@sha256:" + strings.Repeat("a", 64),
		"gameserver": "cloud-clicker/gameserver:fixture@sha256:" + strings.Repeat("b", 64),
		"postgres":   "postgres:fixture@sha256:" + strings.Repeat("c", 64),
	}
	if _, err := RenderCompose(template, images); err != nil {
		return err
	}
	caddyfile, err := os.ReadFile(filepath.Join(root, "deployment", "Caddyfile"))
	if err != nil {
		return err
	}
	if err := ValidateCaddyfile(caddyfile); err != nil {
		return err
	}
	dockerfile, err := os.ReadFile(filepath.Join(root, "deployment", "Dockerfile.gameserver"))
	if err != nil {
		return err
	}
	return ValidateGameserverDockerfile(dockerfile)
}

func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	leftCopy, rightCopy := append([]string(nil), left...), append([]string(nil), right...)
	sort.Strings(leftCopy)
	sort.Strings(rightCopy)
	for index := range leftCopy {
		if leftCopy[index] != rightCopy[index] {
			return false
		}
	}
	return true
}
