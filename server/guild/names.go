package guild

import (
	"bufio"
	"bytes"
	"errors"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

var ErrInvalidNamePolicy = errors.New("invalid guild name policy")

// NormalizeGuildName is the sole canonicalization path used before policy and
// uniqueness checks. The database stores this value, never client spelling.
func NormalizeGuildName(value string) (string, bool) {
	value = strings.ToLower(norm.NFKC.String(value))
	value = strings.Join(strings.Fields(value), " ")
	if !utf8.ValidString(value) || len(value) < 3 || len(value) > 24 ||
		value[0] == ' ' || value[0] == '-' || value[0] == '_' ||
		value[len(value)-1] == ' ' || value[len(value)-1] == '-' || value[len(value)-1] == '_' {
		return "", false
	}
	hasAlphaNumeric := false
	for _, character := range value {
		if character >= 'a' && character <= 'z' || character >= '0' && character <= '9' {
			hasAlphaNumeric = true
			continue
		}
		if character != ' ' && character != '-' && character != '_' {
			return "", false
		}
	}
	return value, hasAlphaNumeric
}

// DenylistNameValidator combines the committed baseline with optional
// deployment additions. Callers cannot remove baseline entries.
type DenylistNameValidator struct {
	blocked []string
}

func NewDenylistNameValidator(baseline, additions []byte) (*DenylistNameValidator, error) {
	base, err := parseNameDenylist(baseline)
	if err != nil || len(base) == 0 {
		return nil, ErrInvalidNamePolicy
	}
	extra, err := parseNameDenylist(additions)
	if err != nil {
		return nil, ErrInvalidNamePolicy
	}
	seen := make(map[string]bool, len(base)+len(extra))
	blocked := make([]string, 0, len(base)+len(extra))
	for _, value := range append(base, extra...) {
		if !seen[value] {
			seen[value] = true
			blocked = append(blocked, value)
		}
	}
	return &DenylistNameValidator{blocked: blocked}, nil
}

func parseNameDenylist(data []byte) ([]string, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, nil
	}
	var values []string
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		normalized, ok := NormalizeGuildName(line)
		if !ok {
			return nil, ErrInvalidNamePolicy
		}
		values = append(values, normalized)
	}
	return values, scanner.Err()
}

func (validator *DenylistNameValidator) ValidateGuildName(value string) bool {
	if validator == nil {
		return false
	}
	normalized, ok := NormalizeGuildName(value)
	if !ok || normalized != value {
		return false
	}
	for _, blocked := range validator.blocked {
		if strings.Contains(normalized, blocked) {
			return false
		}
	}
	return true
}
