package production

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"testing"
)

// Recorded sources describe the historical measurement, not an immutable
// checkout for a regression. Only those digests may differ; every other byte
// of the freshly executed report must still match its recorded observation.
func researchRegressionCompare(report any, sources *map[string]string, recorded []byte) ([]string, error) {
	var provenance struct {
		Sources map[string]string `json:"source_sha256"`
	}
	if err := json.Unmarshal(recorded, &provenance); err != nil {
		return nil, fmt.Errorf("invalid recorded research: %w", err)
	}
	if sources == nil || len(*sources) == 0 || len(*sources) != len(provenance.Sources) {
		return nil, fmt.Errorf("research source population changed")
	}
	var changed []string
	for path, current := range *sources {
		previous, exists := provenance.Sources[path]
		if !exists || path == "" {
			return nil, fmt.Errorf("research source population changed")
		}
		for _, digest := range []string{previous, current} {
			decoded, err := hex.DecodeString(digest)
			if err != nil || len(decoded) != 32 || hex.EncodeToString(decoded) != digest {
				return nil, fmt.Errorf("invalid research source digest: %s", path)
			}
		}
		if current != previous {
			changed = append(changed, path)
		}
	}
	sort.Strings(changed)
	current := *sources
	defer func() { *sources = current }()
	*sources = provenance.Sources
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return changed, err
	}
	if !bytes.Equal(append(encoded, '\n'), recorded) {
		return changed, fmt.Errorf("research results drifted (changed sources: %v)", changed)
	}
	return changed, nil
}

func sourceResearchRegressionArtifact(t *testing.T, path, switchName string, report any, sources *map[string]string) {
	t.Helper()
	if os.Getenv(switchName) == "1" {
		// A deliberately requested new observation still records the actual source
		// identities. A regression never writes or restamps historical evidence.
		sourceResearchArtifact(t, path, switchName, report)
		return
	}
	recorded, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	changed, err := researchRegressionCompare(report, sources, recorded)
	if err != nil {
		t.Fatal(err)
	}
	if len(changed) != 0 {
		t.Logf("current outputs match historical observation; source bytes changed: %v; not a re-observation of the recorded checkout", changed)
	}
}

func TestResearchRegressionComparison(t *testing.T) {
	type observation struct {
		Version    int               `json:"version"`
		Acceptance string            `json:"acceptance_status"`
		Sources    map[string]string `json:"source_sha256"`
		Rows       []string          `json:"rows"`
	}
	const oldDigest = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	const newDigest = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	baseline := observation{1, "NOT_PROVEN", map[string]string{"producer.go": oldDigest}, []string{"receipt", "state"}}
	encoded, err := json.MarshalIndent(baseline, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	encoded = append(encoded, '\n')
	for _, test := range []struct {
		name    string
		change  func(*observation)
		refused bool
		changed []string
	}{
		{"unchanged", func(*observation) {}, false, nil},
		{"source bytes only", func(row *observation) { row.Sources["producer.go"] = newDigest }, false, []string{"producer.go"}},
		{"wrong receipt", func(row *observation) { row.Rows[0] = "wrong" }, true, nil},
		{"source and result changed", func(row *observation) {
			row.Sources["producer.go"] = newDigest
			row.Rows[0] = "wrong"
		}, true, []string{"producer.go"}},
		{"truncated population", func(row *observation) { row.Rows = row.Rows[:1] }, true, nil},
		{"false acceptance", func(row *observation) { row.Acceptance = "PROVEN" }, true, nil},
		{"wrong version", func(row *observation) { row.Version++ }, true, nil},
		{"missing source", func(row *observation) { delete(row.Sources, "producer.go") }, true, nil},
		{"different source", func(row *observation) { delete(row.Sources, "producer.go"); row.Sources["other.go"] = oldDigest }, true, nil},
		{"invalid digest", func(row *observation) { row.Sources["producer.go"] = "not a digest" }, true, nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			actual := observation{baseline.Version, baseline.Acceptance, map[string]string{"producer.go": oldDigest}, append([]string{}, baseline.Rows...)}
			test.change(&actual)
			before, err := json.Marshal(actual)
			if err != nil {
				t.Fatal(err)
			}
			changed, err := researchRegressionCompare(&actual, &actual.Sources, encoded)
			if (err != nil) != test.refused || !reflect.DeepEqual(changed, test.changed) {
				t.Fatalf("changed=%v err=%v, want changed=%v refused=%v", changed, err, test.changed, test.refused)
			}
			after, err := json.Marshal(actual)
			if err != nil || !bytes.Equal(before, after) {
				t.Fatal("comparison mutated current observation")
			}
		})
	}
	for _, invalid := range [][]byte{
		[]byte("not JSON"),
		bytes.Replace(encoded, []byte(oldDigest), []byte("bad"), 1),
		bytes.Replace(encoded, []byte(`"version": 1,`), []byte(`"version": 1, "version": 1,`), 1),
	} {
		if _, err := researchRegressionCompare(&baseline, &baseline.Sources, invalid); err == nil {
			t.Fatal("invalid historical artifact admitted")
		}
	}
}
