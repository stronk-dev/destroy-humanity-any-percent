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
	Logging struct {
		Driver  string            `yaml:"driver"`
		Options map[string]string `yaml:"options"`
	} `yaml:"logging"`
}

type composeNetwork struct {
	Internal bool `yaml:"internal"`
}

type composeSecret struct {
	File string `yaml:"file"`
}

var releaseImageNames = []string{"alertmanager", "caddy", "gameserver", "node-exporter", "postgres", "prometheus"}

func RenderCompose(template []byte, images map[string]string) ([]byte, error) {
	if len(template) == 0 || len(images) != len(releaseImageNames) {
		return nil, ErrInvalidContent
	}
	result := append([]byte(nil), template...)
	for _, imageName := range releaseImageNames {
		name := strings.ToUpper(strings.ReplaceAll(imageName, "-", "_"))
		reference := images[imageName]
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
	if len(data) == 0 || yaml.Unmarshal(data, &model) != nil || len(model.Services) != 7 {
		return ErrInvalidContent
	}
	wantServices := []string{"alertmanager", "backup", "caddy", "gameserver", "node-exporter", "postgres", "prometheus"}
	for _, name := range wantServices {
		service, ok := model.Services[name]
		if !ok || !imageReferencePattern.MatchString(service.Image) {
			return fmt.Errorf("%w: service %s", ErrInvalidContent, name)
		}
		if name != "caddy" && len(service.Ports) != 0 {
			return fmt.Errorf("%w: non-Caddy port publication", ErrInvalidContent)
		}
		if service.Logging.Driver != "journald" || service.Logging.Options["tag"] != "cloud-clicker-"+name {
			return fmt.Errorf("%w: service %s does not use its bounded journald tag", ErrInvalidContent, name)
		}
	}
	for _, name := range []string{"alertmanager", "caddy", "node-exporter", "postgres", "prometheus"} {
		if strings.HasPrefix(model.Services[name].Image, "sha256:") {
			return fmt.Errorf("%w: upstream image requires repository digest", ErrInvalidContent)
		}
	}
	caddy := model.Services["caddy"]
	gameserver := model.Services["gameserver"]
	postgres := model.Services["postgres"]
	backup := model.Services["backup"]
	prometheus := model.Services["prometheus"]
	alertmanager := model.Services["alertmanager"]
	nodeExporter := model.Services["node-exporter"]
	if len(caddy.Ports) != 2 || !sameStrings(caddy.Networks, []string{"application", "edge", "operations"}) || !sameStrings(gameserver.Networks, []string{"application", "database"}) ||
		!sameStrings(postgres.Networks, []string{"database"}) || !sameStrings(backup.Networks, []string{"database"}) ||
		!sameStrings(prometheus.Networks, []string{"application", "operations"}) || !sameStrings(alertmanager.Networks, []string{"edge", "operations"}) || !sameStrings(nodeExporter.Networks, []string{"operations"}) ||
		!sameStrings(caddy.Expose, []string{"2019"}) || !sameStrings(gameserver.Expose, []string{"8080"}) || !sameStrings(prometheus.Expose, []string{"9090"}) ||
		!sameStrings(alertmanager.Expose, []string{"9093"}) || !sameStrings(nodeExporter.Expose, []string{"9100"}) || len(postgres.Expose) != 0 || len(backup.Expose) != 0 {
		return fmt.Errorf("%w: invalid service topology caddy_ports=%v caddy_networks=%v gameserver_expose=%v gameserver_networks=%v postgres_expose=%v postgres_networks=%v",
			ErrInvalidContent, caddy.Ports, caddy.Networks, gameserver.Expose, gameserver.Networks, postgres.Expose, postgres.Networks)
	}
	if !model.Networks["application"].Internal || !model.Networks["database"].Internal || !model.Networks["operations"].Internal || model.Networks["edge"].Internal {
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
	wantBackupCommand := []string{"--age-recipient=${CLOUD_CLICKER_AGE_RECIPIENT:?set the public age X25519 recipient}", "--database-url-file=/run/secrets/database-url", "--epoch=/opt/cloud-clicker/content/balance/epochs/phase0.json", "--metrics-dir=/operations", "--release-manifest=/opt/cloud-clicker/release-manifest.json", "--server-id=${CLOUD_CLICKER_SERVER_ID:?set a canonical UUID}", "--target=/backups", "schedule"}
	if backup.Image != postgres.Image || backup.User != "70:70" || !backup.ReadOnly || !sameStrings(backup.CapDrop, []string{"ALL"}) || !sameStrings(backup.SecurityOpt, []string{"no-new-privileges:true"}) || len(backup.Tmpfs) != 1 ||
		!sameStrings(backup.Entrypoint, []string{"/opt/cloud-clicker/deployment-backup"}) || !sameStrings(backup.Command, wantBackupCommand) || backup.DependsOn["postgres"].Condition != "service_healthy" || len(backup.DependsOn) != 1 ||
		!sameStrings(backup.Volumes, []string{"./content:/opt/cloud-clicker/content:ro", "./deployment-backup:/opt/cloud-clicker/deployment-backup:ro", "./release-manifest.json:/opt/cloud-clicker/release-manifest.json:ro", "${CLOUD_CLICKER_BACKUP_TARGET:?set the separately mounted backup target}:/backups", "${CLOUD_CLICKER_OPERATIONS_METRICS:?set the node-exporter textfile directory}:/operations"}) {
		return fmt.Errorf("%w: invalid backup worker boundary", ErrInvalidContent)
	}
	if !sameStrings(alertmanager.Secrets, []string{"alertmanager-config"}) || len(prometheus.Secrets) != 0 || len(nodeExporter.Secrets) != 0 ||
		!prometheus.ReadOnly || !alertmanager.ReadOnly || !nodeExporter.ReadOnly ||
		!sameStrings(prometheus.Volumes, []string{"./deployment-operations:/opt/cloud-clicker/deployment-operations:ro", "./operations/cloud-clicker-alerts.yml:/etc/prometheus/cloud-clicker-alerts.yml:ro", "./operations/prometheus.yml:/etc/prometheus/prometheus.yml:ro", "prometheus_data:/prometheus"}) ||
		!sameStrings(alertmanager.Volumes, []string{"./deployment-operations:/opt/cloud-clicker/deployment-operations:ro", "alertmanager_data:/alertmanager"}) ||
		!sameStrings(nodeExporter.Volumes, []string{"/:/host/root:ro,rslave", "/proc:/host/proc:ro", "/sys:/host/sys:ro", "${CLOUD_CLICKER_OPERATIONS_METRICS:?set the node-exporter textfile directory}:/textfile:ro"}) {
		return fmt.Errorf("%w: invalid private operations boundary", ErrInvalidContent)
	}
	return nil
}

func ValidateCaddyfile(data []byte) error {
	text := string(data)
	for _, required := range []string{"admin :2019", "metrics", "{$CLOUD_CLICKER_PUBLIC_ORIGIN}", "/api/*", "/connection/websocket", "/healthz", "/readyz", "reverse_proxy gameserver:8080", "root * /srv", "try_files {path} /index.html"} {
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
	images := map[string]string{}
	for index, name := range releaseImageNames {
		images[name] = name + ":fixture@sha256:" + strings.Repeat(string(rune('a'+index)), 64)
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
