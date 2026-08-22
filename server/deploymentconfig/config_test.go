package deploymentconfig

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestProductionConfigLoadsFileSecretsAndRotationPairs(t *testing.T) {
	environment, secrets := validProductionFixture()
	config, err := Load(environment, fixtureReader(secrets))
	if err != nil {
		t.Fatal(err)
	}
	if config.Mode != ModeProduction || config.PublicOrigin != "https://play.example.test" || config.TrustedProxyHops != 1 ||
		config.ContentRoot != ProductionContentRoot || config.ListenAddress != ":8080" || config.ActivityBracket != "activity.standard" {
		t.Fatalf("unexpected production config: %+v", config)
	}
	if config.JWT.CurrentID != "jwt-current" || config.JWT.PreviousID != "jwt-previous" || len(config.JWT.Current) != 32 || len(config.JWT.Previous) != 32 {
		t.Fatalf("unexpected JWT rotation pair: %+v", config.JWT)
	}
	if config.Bootstrap.CurrentID != "bootstrap-current" || config.Bootstrap.PreviousID != "bootstrap-previous" || len(config.Bootstrap.Current) != 32 || len(config.Bootstrap.Previous) != 32 {
		t.Fatalf("unexpected bootstrap rotation pair: %+v", config.Bootstrap)
	}
	if config.Cursor == nil || config.Cursor.CurrentID != "cursor-current" || config.Cursor.PreviousID != "cursor-previous" {
		t.Fatalf("unexpected cursor rotation pair: %+v", config.Cursor)
	}
}

func TestProductionConfigRejectsEveryFailClosedFamily(t *testing.T) {
	baseEnvironment, baseSecrets := validProductionFixture()
	tests := []struct {
		name   string
		mutate func([]string, map[string][]byte) ([]string, map[string][]byte)
	}{
		{"unknown deployment key", setEnv("CLOUD_CLICKER_SURPRISE", "accepted")},
		{"duplicate environment key", appendEnv("CLOUD_CLICKER_PUBLIC_ORIGIN=https://other.example.test")},
		{"wrong deployment mode", setEnv("CLOUD_CLICKER_DEPLOYMENT_MODE", "prod")},
		{"missing public origin", removeEnv("CLOUD_CLICKER_PUBLIC_ORIGIN")},
		{"insecure origin", setEnv("CLOUD_CLICKER_PUBLIC_ORIGIN", "http://play.example.test")},
		{"origin path", setEnv("CLOUD_CLICKER_PUBLIC_ORIGIN", "https://play.example.test/game")},
		{"origin query", setEnv("CLOUD_CLICKER_PUBLIC_ORIGIN", "https://play.example.test?tenant=one")},
		{"origin fragment", setEnv("CLOUD_CLICKER_PUBLIC_ORIGIN", "https://play.example.test#game")},
		{"noncanonical origin", setEnv("CLOUD_CLICKER_PUBLIC_ORIGIN", "https://PLAY.example.test")},
		{"default TLS port", setEnv("CLOUD_CLICKER_PUBLIC_ORIGIN", "https://play.example.test:443")},
		{"wrong proxy depth", setEnv("CLOUD_CLICKER_TRUSTED_PROXY_HOPS", "0")},
		{"missing proxy depth", removeEnv("CLOUD_CLICKER_TRUSTED_PROXY_HOPS")},
		{"wrong content root", setEnv("CLOUD_CLICKER_CONTENT_ROOT", "/workspace")},
		{"missing content root", removeEnv("CLOUD_CLICKER_CONTENT_ROOT")},
		{"malformed server ID", setEnv("CLOUD_CLICKER_SERVER_ID", "server-one")},
		{"missing server ID", removeEnv("CLOUD_CLICKER_SERVER_ID")},
		{"malformed listen address", setEnv("LISTEN_ADDR", "public")},
		{"legacy database secret", appendEnv("DATABASE_URL=postgres://inline")},
		{"legacy JWT secret", appendEnv("CLOUD_CLICKER_JWT_KEY=inline")},
		{"missing database file", removeSecret("/run/secrets/database-url")},
		{"missing database file setting", removeEnv("DATABASE_URL_FILE")},
		{"relative secret path", setEnv("DATABASE_URL_FILE", "database-url")},
		{"database without password", replaceSecret("/run/secrets/database-url", []byte("postgres://cloud@clicker-db/cloud\n"))},
		{"database with empty password", replaceSecret("/run/secrets/database-url", []byte("postgres://cloud:@clicker-db/cloud\n"))},
		{"empty JWT secret", replaceSecret("/run/secrets/jwt-current", nil)},
		{"multiline JWT secret", replaceSecret("/run/secrets/jwt-current", []byte("one\ntwo\n"))},
		{"malformed JWT base64", replaceSecret("/run/secrets/jwt-current", []byte("not base64\n"))},
		{"short JWT key", replaceSecret("/run/secrets/jwt-current", encoded(bytes.Repeat([]byte{1}, 31)))},
		{"missing current JWT ID", removeEnv("CLOUD_CLICKER_JWT_CURRENT_ID")},
		{"missing current JWT file", removeEnv("CLOUD_CLICKER_JWT_CURRENT_KEY_FILE")},
		{"half previous JWT pair", removeEnv("CLOUD_CLICKER_JWT_PREVIOUS_KEY_FILE")},
		{"duplicate JWT ID", setEnv("CLOUD_CLICKER_JWT_PREVIOUS_ID", "jwt-current")},
		{"duplicate JWT value", copySecret("/run/secrets/jwt-current", "/run/secrets/jwt-previous")},
		{"wrong bootstrap key length", replaceSecret("/run/secrets/bootstrap-current", encoded(bytes.Repeat([]byte{3}, 33)))},
		{"missing current bootstrap ID", removeEnv("CLOUD_CLICKER_BOOTSTRAP_CURRENT_ID")},
		{"missing current bootstrap file", removeEnv("CLOUD_CLICKER_BOOTSTRAP_CURRENT_KEY_FILE")},
		{"half previous bootstrap pair", removeEnv("CLOUD_CLICKER_BOOTSTRAP_PREVIOUS_ID")},
		{"duplicate bootstrap ID", setEnv("CLOUD_CLICKER_BOOTSTRAP_PREVIOUS_ID", "bootstrap-current")},
		{"duplicate bootstrap value", copySecret("/run/secrets/bootstrap-current", "/run/secrets/bootstrap-previous")},
		{"half current cursor pair", removeEnv("CLOUD_CLICKER_CURSOR_CURRENT_KEY_FILE")},
		{"half previous cursor pair", removeEnv("CLOUD_CLICKER_CURSOR_PREVIOUS_ID")},
		{"duplicate cursor ID", setEnv("CLOUD_CLICKER_CURSOR_PREVIOUS_ID", "cursor-current")},
		{"duplicate cursor value", copySecret("/run/secrets/cursor-current", "/run/secrets/cursor-previous")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			environment := append([]string(nil), baseEnvironment...)
			secrets := cloneSecrets(baseSecrets)
			environment, secrets = test.mutate(environment, secrets)
			if _, err := Load(environment, fixtureReader(secrets)); !errors.Is(err, ErrInvalid) {
				t.Fatalf("invalid fixture accepted: %v", err)
			}
		})
	}
}

func TestProductionConfigAllowsCursorKeysToRemainUncomposed(t *testing.T) {
	environment, secrets := validProductionFixture()
	for _, name := range []string{"CLOUD_CLICKER_CURSOR_CURRENT_ID", "CLOUD_CLICKER_CURSOR_CURRENT_KEY_FILE", "CLOUD_CLICKER_CURSOR_PREVIOUS_ID", "CLOUD_CLICKER_CURSOR_PREVIOUS_KEY_FILE"} {
		environment, secrets = removeEnv(name)(environment, secrets)
	}
	config, err := Load(environment, fixtureReader(secrets))
	if err != nil || config.Cursor != nil {
		t.Fatalf("cursor=%+v err=%v", config.Cursor, err)
	}
}

func TestDevelopmentConfigRetainsTheDeclaredLegacyProfileOnly(t *testing.T) {
	key := base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{7}, 32))
	config, err := Load([]string{
		"DATABASE_URL=postgres://cloud:secret@localhost/cloud?sslmode=disable",
		"CLOUD_CLICKER_SERVER_ID=01986666-b001-4000-8000-000000000001",
		"CLOUD_CLICKER_JWT_KEY=" + key,
		"CLOUD_CLICKER_BOOTSTRAP_KEY_ID=dev-bootstrap",
		"CLOUD_CLICKER_BOOTSTRAP_KEY=" + key,
		"CLOUD_CLICKER_REPOSITORY_ROOT=/workspace",
		"CLOUD_CLICKER_ACTIVITY_BRACKET=activity.standard",
		"LISTEN_ADDR=127.0.0.1:18081",
	}, fixtureReader(nil))
	if err != nil || config.Mode != "development" || config.ContentRoot != "/workspace" || config.PublicOrigin != "" || config.TrustedProxyHops != 0 {
		t.Fatalf("config=%+v err=%v", config, err)
	}
}

func TestConfigErrorsNeverContainSecretValues(t *testing.T) {
	environment, secrets := validProductionFixture()
	secret := "leak-me-never"
	secrets["/run/secrets/jwt-current"] = []byte(secret)
	_, err := Load(environment, fixtureReader(secrets))
	if err == nil || strings.Contains(err.Error(), secret) {
		t.Fatalf("secret-bearing error: %v", err)
	}
}

func validProductionFixture() ([]string, map[string][]byte) {
	values := map[string]string{
		"CLOUD_CLICKER_DEPLOYMENT_MODE":             "production",
		"CLOUD_CLICKER_PUBLIC_ORIGIN":               "https://play.example.test",
		"CLOUD_CLICKER_TRUSTED_PROXY_HOPS":          "1",
		"CLOUD_CLICKER_CONTENT_ROOT":                ProductionContentRoot,
		"CLOUD_CLICKER_SERVER_ID":                   "01986666-b001-4000-8000-000000000001",
		"DATABASE_URL_FILE":                         "/run/secrets/database-url",
		"CLOUD_CLICKER_JWT_CURRENT_ID":              "jwt-current",
		"CLOUD_CLICKER_JWT_CURRENT_KEY_FILE":        "/run/secrets/jwt-current",
		"CLOUD_CLICKER_JWT_PREVIOUS_ID":             "jwt-previous",
		"CLOUD_CLICKER_JWT_PREVIOUS_KEY_FILE":       "/run/secrets/jwt-previous",
		"CLOUD_CLICKER_BOOTSTRAP_CURRENT_ID":        "bootstrap-current",
		"CLOUD_CLICKER_BOOTSTRAP_CURRENT_KEY_FILE":  "/run/secrets/bootstrap-current",
		"CLOUD_CLICKER_BOOTSTRAP_PREVIOUS_ID":       "bootstrap-previous",
		"CLOUD_CLICKER_BOOTSTRAP_PREVIOUS_KEY_FILE": "/run/secrets/bootstrap-previous",
		"CLOUD_CLICKER_CURSOR_CURRENT_ID":           "cursor-current",
		"CLOUD_CLICKER_CURSOR_CURRENT_KEY_FILE":     "/run/secrets/cursor-current",
		"CLOUD_CLICKER_CURSOR_PREVIOUS_ID":          "cursor-previous",
		"CLOUD_CLICKER_CURSOR_PREVIOUS_KEY_FILE":    "/run/secrets/cursor-previous",
	}
	environment := make([]string, 0, len(values))
	for _, name := range []string{
		"CLOUD_CLICKER_DEPLOYMENT_MODE", "CLOUD_CLICKER_PUBLIC_ORIGIN", "CLOUD_CLICKER_TRUSTED_PROXY_HOPS", "CLOUD_CLICKER_CONTENT_ROOT", "CLOUD_CLICKER_SERVER_ID", "DATABASE_URL_FILE",
		"CLOUD_CLICKER_JWT_CURRENT_ID", "CLOUD_CLICKER_JWT_CURRENT_KEY_FILE", "CLOUD_CLICKER_JWT_PREVIOUS_ID", "CLOUD_CLICKER_JWT_PREVIOUS_KEY_FILE",
		"CLOUD_CLICKER_BOOTSTRAP_CURRENT_ID", "CLOUD_CLICKER_BOOTSTRAP_CURRENT_KEY_FILE", "CLOUD_CLICKER_BOOTSTRAP_PREVIOUS_ID", "CLOUD_CLICKER_BOOTSTRAP_PREVIOUS_KEY_FILE",
		"CLOUD_CLICKER_CURSOR_CURRENT_ID", "CLOUD_CLICKER_CURSOR_CURRENT_KEY_FILE", "CLOUD_CLICKER_CURSOR_PREVIOUS_ID", "CLOUD_CLICKER_CURSOR_PREVIOUS_KEY_FILE",
	} {
		environment = append(environment, name+"="+values[name])
	}
	return environment, map[string][]byte{
		"/run/secrets/database-url":       []byte("postgres://cloud:secret@clicker-db/cloud?sslmode=disable\n"),
		"/run/secrets/jwt-current":        encoded(bytes.Repeat([]byte{1}, 32)),
		"/run/secrets/jwt-previous":       encoded(bytes.Repeat([]byte{2}, 32)),
		"/run/secrets/bootstrap-current":  encoded(bytes.Repeat([]byte{3}, 32)),
		"/run/secrets/bootstrap-previous": encoded(bytes.Repeat([]byte{4}, 32)),
		"/run/secrets/cursor-current":     encoded(bytes.Repeat([]byte{5}, 32)),
		"/run/secrets/cursor-previous":    encoded(bytes.Repeat([]byte{6}, 32)),
	}
}

func encoded(value []byte) []byte {
	return []byte(base64.StdEncoding.EncodeToString(value) + "\n")
}

func fixtureReader(secrets map[string][]byte) ReadFile {
	return func(path string) ([]byte, error) {
		value, ok := secrets[path]
		if !ok {
			return nil, fmt.Errorf("missing")
		}
		return append([]byte(nil), value...), nil
	}
}

type fixtureMutation func([]string, map[string][]byte) ([]string, map[string][]byte)

func setEnv(name, value string) fixtureMutation {
	return func(environment []string, secrets map[string][]byte) ([]string, map[string][]byte) {
		prefix := name + "="
		for index := range environment {
			if strings.HasPrefix(environment[index], prefix) {
				environment[index] = prefix + value
				return environment, secrets
			}
		}
		return append(environment, prefix+value), secrets
	}
}

func appendEnv(entry string) fixtureMutation {
	return func(environment []string, secrets map[string][]byte) ([]string, map[string][]byte) {
		return append(environment, entry), secrets
	}
}

func removeEnv(name string) fixtureMutation {
	return func(environment []string, secrets map[string][]byte) ([]string, map[string][]byte) {
		result := environment[:0]
		for _, entry := range environment {
			if !strings.HasPrefix(entry, name+"=") {
				result = append(result, entry)
			}
		}
		return result, secrets
	}
}

func replaceSecret(path string, value []byte) fixtureMutation {
	return func(environment []string, secrets map[string][]byte) ([]string, map[string][]byte) {
		secrets[path] = append([]byte(nil), value...)
		return environment, secrets
	}
}

func removeSecret(path string) fixtureMutation {
	return func(environment []string, secrets map[string][]byte) ([]string, map[string][]byte) {
		delete(secrets, path)
		return environment, secrets
	}
}

func copySecret(from, to string) fixtureMutation {
	return func(environment []string, secrets map[string][]byte) ([]string, map[string][]byte) {
		secrets[to] = append([]byte(nil), secrets[from]...)
		return environment, secrets
	}
}

func cloneSecrets(source map[string][]byte) map[string][]byte {
	result := make(map[string][]byte, len(source))
	for path, value := range source {
		result[path] = append([]byte(nil), value...)
	}
	return result
}
