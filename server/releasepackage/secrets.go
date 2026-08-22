package releasepackage

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
)

const maximumSecretScanMember = 128 << 20

type SecretFinding struct {
	Path string
	Rule string
}

var secretRules = []struct {
	name    string
	pattern *regexp.Regexp
}{
	{"private-key", regexp.MustCompile(`-----BEGIN (?:RSA |EC |OPENSSH )?PRIVATE KEY-----`)},
	{"aws-access-key", regexp.MustCompile(`AKIA[0-9A-Z]{16}`)},
	{"github-token", regexp.MustCompile(`(?:gh[pousr]_[A-Za-z0-9]{36,255}|github_pat_[A-Za-z0-9_]{40,255})`)},
	{"slack-token", regexp.MustCompile(`xox[baprs]-[0-9A-Za-z-]{20,}`)},
	{"stripe-live-key", regexp.MustCompile(`sk_live_[0-9A-Za-z]{16,}`)},
	{"google-api-key", regexp.MustCompile(`AIza[0-9A-Za-z_-]{35}`)},
	{"seeded-fixture", regexp.MustCompile(`CLOUD_CLICKER_` + `SECRET_SCAN_SENTINEL_[0-9A-Za-z]{16,}`)},
}

func ScanTrackedFiles(root string, paths []string) ([]SecretFinding, error) {
	findings := []SecretFinding{}
	for _, path := range paths {
		if !validRelativePath(path) {
			return nil, ErrInvalidContent
		}
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
		if err != nil {
			return nil, err
		}
		findings = append(findings, scanSecretBytes(path, data)...)
	}
	sortFindings(findings)
	return findings, nil
}

func ScanTree(root string) ([]SecretFinding, error) {
	findings := []SecretFinding{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil || !validRelativePath(filepath.ToSlash(relative)) || entry.Type()&os.ModeSymlink != 0 || !entry.Type().IsRegular() {
			return ErrInvalidContent
		}
		if filepath.ToSlash(relative) == "images/gameserver.docker.tar" {
			return nil
		}
		info, err := entry.Info()
		if err != nil || info.Size() > maximumSecretScanMember {
			return ErrInvalidContent
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		findings = append(findings, scanSecretBytes(filepath.ToSlash(relative), data)...)
		return nil
	})
	if err != nil {
		return nil, errors.Join(ErrInvalidContent, err)
	}
	sortFindings(findings)
	return findings, nil
}

func ScanDockerArchive(path string) ([]SecretFinding, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	reader := tar.NewReader(file)
	findings := []SecretFinding{}
	for {
		header, err := reader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil || header.Size < 0 || header.Size > maximumSecretScanMember {
			return nil, ErrInvalidContent
		}
		if header.Typeflag != tar.TypeReg && header.Typeflag != tar.TypeRegA {
			continue
		}
		data, err := io.ReadAll(io.LimitReader(reader, header.Size+1))
		if err != nil || int64(len(data)) != header.Size {
			return nil, ErrInvalidContent
		}
		findings = append(findings, scanSecretBytes("images/gameserver.docker.tar:"+header.Name, data)...)
	}
	position, seekErr := file.Seek(0, io.SeekCurrent)
	info, statErr := file.Stat()
	if seekErr != nil || statErr != nil || position != info.Size() {
		return nil, ErrInvalidContent
	}
	sortFindings(findings)
	return findings, nil
}

func RequireNoSecrets(findings []SecretFinding) error {
	if len(findings) == 0 {
		return nil
	}
	return fmt.Errorf("%w: secret material rule=%s path=%s", ErrInvalidContent, findings[0].Rule, findings[0].Path)
}

func scanSecretBytes(path string, data []byte) []SecretFinding {
	findings := []SecretFinding{}
	for _, rule := range secretRules {
		if rule.pattern.Match(data) {
			findings = append(findings, SecretFinding{Path: path, Rule: rule.name})
		}
	}
	return findings
}

func sortFindings(findings []SecretFinding) {
	sort.Slice(findings, func(left, right int) bool {
		return findings[left].Path+"\x00"+findings[left].Rule < findings[right].Path+"\x00"+findings[right].Rule
	})
}
