// Package deploymentconfig owns the fail-closed gameserver deployment
// environment. Production startup and its preflight command use this same
// decoder; secret values are returned to composition but never included in an
// error.
package deploymentconfig

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const (
	ModeProduction        = "production"
	ProductionContentRoot = "/opt/cloud-clicker/content"
	defaultListenAddress  = ":8080"
	maximumSecretBytes    = 64 << 10
)

var (
	ErrInvalid = errors.New("invalid deployment configuration")
	uuid       = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	keyID      = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)
)

type KeyPair struct {
	CurrentID  string
	Current    []byte
	PreviousID string
	Previous   []byte
}

type Config struct {
	Mode             string
	PublicOrigin     string
	TrustedProxyHops int
	ContentRoot      string
	ServerID         string
	ListenAddress    string
	DatabaseURL      string
	ActivityBracket  string
	JWT              KeyPair
	Bootstrap        KeyPair
	Cursor           *KeyPair
}

type ReadFile func(string) ([]byte, error)

func LoadEnvironment() (Config, error) {
	return Load(os.Environ(), os.ReadFile)
}

func Load(environ []string, readFile ReadFile) (Config, error) {
	values, err := environmentMap(environ)
	if err != nil || readFile == nil {
		return Config{}, ErrInvalid
	}
	if values["CLOUD_CLICKER_DEPLOYMENT_MODE"] == ModeProduction {
		return loadProduction(values, readFile)
	}
	return loadDevelopment(values)
}

func loadProduction(values map[string]string, readFile ReadFile) (Config, error) {
	if err := rejectUnknownProductionKeys(values); err != nil {
		return Config{}, err
	}
	origin := values["CLOUD_CLICKER_PUBLIC_ORIGIN"]
	if !ValidProductionOrigin(origin) {
		return Config{}, fieldError("CLOUD_CLICKER_PUBLIC_ORIGIN")
	}
	if values["CLOUD_CLICKER_TRUSTED_PROXY_HOPS"] != "1" {
		return Config{}, fieldError("CLOUD_CLICKER_TRUSTED_PROXY_HOPS")
	}
	if values["CLOUD_CLICKER_CONTENT_ROOT"] != ProductionContentRoot {
		return Config{}, fieldError("CLOUD_CLICKER_CONTENT_ROOT")
	}
	if !uuid.MatchString(values["CLOUD_CLICKER_SERVER_ID"]) {
		return Config{}, fieldError("CLOUD_CLICKER_SERVER_ID")
	}
	listen := values["LISTEN_ADDR"]
	if listen == "" {
		listen = defaultListenAddress
	}
	if !validListenAddress(listen) {
		return Config{}, fieldError("LISTEN_ADDR")
	}
	databaseURL, err := readSecretText("DATABASE_URL_FILE", values["DATABASE_URL_FILE"], readFile)
	if err != nil || !validDatabaseURL(databaseURL) {
		return Config{}, fieldError("DATABASE_URL_FILE")
	}
	jwt, err := requiredKeyPair(values, readFile, "CLOUD_CLICKER_JWT", 32, false)
	if err != nil {
		return Config{}, err
	}
	bootstrap, err := requiredKeyPair(values, readFile, "CLOUD_CLICKER_BOOTSTRAP", 32, true)
	if err != nil {
		return Config{}, err
	}
	cursor, err := optionalKeyPair(values, readFile, "CLOUD_CLICKER_CURSOR", 32, false)
	if err != nil {
		return Config{}, err
	}
	return Config{Mode: ModeProduction, PublicOrigin: origin, TrustedProxyHops: 1,
		ContentRoot: ProductionContentRoot, ServerID: values["CLOUD_CLICKER_SERVER_ID"], ListenAddress: listen,
		DatabaseURL: databaseURL, ActivityBracket: "activity.standard", JWT: jwt, Bootstrap: bootstrap, Cursor: cursor}, nil
}

func loadDevelopment(values map[string]string) (Config, error) {
	databaseURL := values["DATABASE_URL"]
	jwtValue, jwtErr := decodeInlineKey(values["CLOUD_CLICKER_JWT_KEY"], 32, false)
	bootstrapValue, bootstrapErr := decodeInlineKey(values["CLOUD_CLICKER_BOOTSTRAP_KEY"], 32, true)
	serverID := values["CLOUD_CLICKER_SERVER_ID"]
	bootstrapID := values["CLOUD_CLICKER_BOOTSTRAP_KEY_ID"]
	if databaseURL == "" || !uuid.MatchString(serverID) || jwtErr != nil || bootstrapErr != nil || !keyID.MatchString(bootstrapID) {
		return Config{}, ErrInvalid
	}
	root := values["CLOUD_CLICKER_REPOSITORY_ROOT"]
	if root == "" {
		root = "."
	}
	activity := values["CLOUD_CLICKER_ACTIVITY_BRACKET"]
	if activity == "" {
		activity = "activity.standard"
	}
	listen := values["LISTEN_ADDR"]
	if listen == "" {
		listen = defaultListenAddress
	}
	if !validListenAddress(listen) {
		return Config{}, ErrInvalid
	}
	return Config{Mode: "development", ContentRoot: root, ServerID: serverID, ListenAddress: listen,
		DatabaseURL: databaseURL, ActivityBracket: activity,
		JWT:       KeyPair{CurrentID: "runtime", Current: jwtValue},
		Bootstrap: KeyPair{CurrentID: bootstrapID, Current: bootstrapValue}}, nil
}

func environmentMap(environ []string) (map[string]string, error) {
	result := make(map[string]string, len(environ))
	for _, entry := range environ {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) != 2 || parts[0] == "" {
			return nil, ErrInvalid
		}
		if _, exists := result[parts[0]]; exists {
			return nil, fieldError(parts[0])
		}
		result[parts[0]] = parts[1]
	}
	return result, nil
}

var productionKeys = map[string]bool{
	"CLOUD_CLICKER_DEPLOYMENT_MODE":             true,
	"CLOUD_CLICKER_PUBLIC_ORIGIN":               true,
	"CLOUD_CLICKER_TRUSTED_PROXY_HOPS":          true,
	"CLOUD_CLICKER_CONTENT_ROOT":                true,
	"CLOUD_CLICKER_SERVER_ID":                   true,
	"CLOUD_CLICKER_JWT_CURRENT_ID":              true,
	"CLOUD_CLICKER_JWT_CURRENT_KEY_FILE":        true,
	"CLOUD_CLICKER_JWT_PREVIOUS_ID":             true,
	"CLOUD_CLICKER_JWT_PREVIOUS_KEY_FILE":       true,
	"CLOUD_CLICKER_BOOTSTRAP_CURRENT_ID":        true,
	"CLOUD_CLICKER_BOOTSTRAP_CURRENT_KEY_FILE":  true,
	"CLOUD_CLICKER_BOOTSTRAP_PREVIOUS_ID":       true,
	"CLOUD_CLICKER_BOOTSTRAP_PREVIOUS_KEY_FILE": true,
	"CLOUD_CLICKER_CURSOR_CURRENT_ID":           true,
	"CLOUD_CLICKER_CURSOR_CURRENT_KEY_FILE":     true,
	"CLOUD_CLICKER_CURSOR_PREVIOUS_ID":          true,
	"CLOUD_CLICKER_CURSOR_PREVIOUS_KEY_FILE":    true,
	"DATABASE_URL_FILE":                         true,
	"LISTEN_ADDR":                               true,
}

func rejectUnknownProductionKeys(values map[string]string) error {
	for name := range values {
		if strings.HasPrefix(name, "CLOUD_CLICKER_") && !productionKeys[name] {
			return fieldError(name)
		}
	}
	for _, legacy := range []string{"DATABASE_URL", "CLOUD_CLICKER_JWT_KEY", "CLOUD_CLICKER_BOOTSTRAP_KEY", "CLOUD_CLICKER_BOOTSTRAP_KEY_ID"} {
		if _, exists := values[legacy]; exists {
			return fieldError(legacy)
		}
	}
	return nil
}

func requiredKeyPair(values map[string]string, readFile ReadFile, prefix string, minimum int, exact bool) (KeyPair, error) {
	currentIDName, currentFileName := prefix+"_CURRENT_ID", prefix+"_CURRENT_KEY_FILE"
	currentID, currentFile := values[currentIDName], values[currentFileName]
	if !keyID.MatchString(currentID) || currentFile == "" {
		return KeyPair{}, fieldError(prefix + "_CURRENT")
	}
	current, err := readEncodedKey(currentFileName, currentFile, readFile, minimum, exact)
	if err != nil {
		return KeyPair{}, err
	}
	pair := KeyPair{CurrentID: currentID, Current: current}
	previousIDName, previousFileName := prefix+"_PREVIOUS_ID", prefix+"_PREVIOUS_KEY_FILE"
	previousID, previousFile := values[previousIDName], values[previousFileName]
	if (previousID == "") != (previousFile == "") {
		return KeyPair{}, fieldError(prefix + "_PREVIOUS")
	}
	if previousID == "" {
		return pair, nil
	}
	if !keyID.MatchString(previousID) || previousID == currentID {
		return KeyPair{}, fieldError(previousIDName)
	}
	previous, err := readEncodedKey(previousFileName, previousFile, readFile, minimum, exact)
	if err != nil {
		return KeyPair{}, err
	}
	if bytes.Equal(current, previous) {
		return KeyPair{}, fieldError(prefix + "_PREVIOUS")
	}
	pair.PreviousID, pair.Previous = previousID, previous
	return pair, nil
}

func optionalKeyPair(values map[string]string, readFile ReadFile, prefix string, minimum int, exact bool) (*KeyPair, error) {
	currentID, currentFile := values[prefix+"_CURRENT_ID"], values[prefix+"_CURRENT_KEY_FILE"]
	previousID, previousFile := values[prefix+"_PREVIOUS_ID"], values[prefix+"_PREVIOUS_KEY_FILE"]
	if currentID == "" && currentFile == "" && previousID == "" && previousFile == "" {
		return nil, nil
	}
	if currentID == "" || currentFile == "" {
		return nil, fieldError(prefix + "_CURRENT")
	}
	pair, err := requiredKeyPair(values, readFile, prefix, minimum, exact)
	if err != nil {
		return nil, err
	}
	return &pair, nil
}

func readEncodedKey(field, path string, readFile ReadFile, minimum int, exact bool) ([]byte, error) {
	text, err := readSecretText(field, path, readFile)
	if err != nil {
		return nil, err
	}
	decoded, err := base64.StdEncoding.Strict().DecodeString(text)
	if err != nil || exact && len(decoded) != minimum || !exact && len(decoded) < minimum {
		return nil, fieldError(field)
	}
	return decoded, nil
}

func decodeInlineKey(value string, minimum int, exact bool) ([]byte, error) {
	decoded, err := base64.StdEncoding.Strict().DecodeString(value)
	if err != nil || exact && len(decoded) != minimum || !exact && len(decoded) < minimum {
		return nil, ErrInvalid
	}
	return decoded, nil
}

func readSecretText(field, path string, readFile ReadFile) (string, error) {
	clean := filepath.Clean(path)
	if path == "" || !filepath.IsAbs(path) || clean != path || !strings.HasPrefix(clean, "/run/secrets/") {
		return "", fieldError(field)
	}
	data, err := readFile(path)
	if err != nil || len(data) == 0 || len(data) > maximumSecretBytes || bytes.IndexByte(data, 0) >= 0 {
		return "", fieldError(field)
	}
	if data[len(data)-1] == '\n' {
		data = data[:len(data)-1]
	}
	if len(data) == 0 || bytes.ContainsAny(data, "\r\n") || strings.TrimSpace(string(data)) != string(data) {
		return "", fieldError(field)
	}
	return string(data), nil
}

func ValidProductionOrigin(value string) bool {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil || parsed.Path != "" || parsed.RawPath != "" ||
		parsed.RawQuery != "" || parsed.Fragment != "" || parsed.Opaque != "" || parsed.String() != value {
		return false
	}
	hostname := parsed.Hostname()
	if hostname == "" || strings.ToLower(hostname) != hostname || strings.HasSuffix(hostname, ".") || parsed.Port() == "443" {
		return false
	}
	if port := parsed.Port(); port != "" {
		value, err := strconv.Atoi(port)
		if err != nil || value < 1 || value > 65535 {
			return false
		}
	}
	return true
}

func validListenAddress(value string) bool {
	_, port, err := net.SplitHostPort(value)
	if err != nil {
		return false
	}
	number, err := strconv.Atoi(port)
	return err == nil && number >= 1 && number <= 65535
}

func validDatabaseURL(value string) bool {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme != "postgres" && parsed.Scheme != "postgresql" || parsed.Host == "" || parsed.User == nil || parsed.Path == "" || parsed.Path == "/" || parsed.Fragment != "" {
		return false
	}
	password, hasPassword := parsed.User.Password()
	return parsed.User.Username() != "" && hasPassword && password != ""
}

func fieldError(field string) error {
	return fmt.Errorf("%w: %s", ErrInvalid, field)
}
