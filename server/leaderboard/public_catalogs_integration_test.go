package leaderboard

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/epochseed"
	"cloud-clicker/server/save"
)

func TestPublicCatalogIntegration(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	db, err := save.OpenPostgres(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := save.Migrate(ctx, db); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `TRUNCATE verification_projection_events,epochs,catalog_sets,accounts,save_streams RESTART IDENTITY CASCADE`); err != nil {
		t.Fatal(err)
	}
	root := filepath.Join("..", "..")
	bundle, err := epochseed.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	staged := t.TempDir()
	if err := os.MkdirAll(filepath.Join(staged, "changelog"), 0o755); err != nil {
		t.Fatal(err)
	}
	for id := int64(1); id <= bundle.Seed.CurrentEpochID+1; id++ {
		if err := os.WriteFile(filepath.Join(staged, "changelog", fmt.Sprintf("epoch-%d.md", id)), []byte("# Fixture epoch\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	repository, err := NewRepository(db, staged)
	if err != nil {
		t.Fatal(err)
	}
	started := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
	if err := repository.ReconcileSeed(ctx, bundle, started); err != nil {
		t.Fatal(err)
	}

	assertBundle := func(t *testing.T, hash string, expected map[string][]byte) {
		t.Helper()
		got, err := repository.PublicCatalog(ctx, hash)
		if err != nil || got.ConstantsHash != hash || len(got.Artifacts) != len(expected) {
			t.Fatalf("catalog=%+v err=%v; want %d artifacts", got, err, len(expected))
		}
		names := make([]string, 0, len(expected))
		for name := range expected {
			names = append(names, name)
		}
		sort.Strings(names)
		for i, name := range names {
			artifact := got.Artifacts[i]
			if artifact.Name != name || artifact.SHA256 != save.ConstantsHash(expected[name]) || !bytes.Equal(artifact.Bytes, expected[name]) {
				t.Fatalf("artifact %s lost exact stored bytes, digest or order", name)
			}
		}
		// A caller's buffer mutation must not change the next read or stored bytes.
		got.Artifacts[0].Bytes[0] = '!'
		again, err := repository.PublicCatalog(ctx, hash)
		if err != nil || !bytes.Equal(again.Artifacts[0].Bytes, expected[names[0]]) {
			t.Fatalf("caller corrupted stored evidence: %v", err)
		}
	}
	t.Run("current committed catalog bytes", func(t *testing.T) { assertBundle(t, bundle.Hash, bundle.Artifacts) })

	// A second accepted epoch uses a formula-bearing set. This is a fixture-only
	// mint through the real owner, not authorization to mint a product release.
	// The old identity must still return its old set, even though a newer set
	// now has formula bytes and its economy JSON differs in exact formatting.
	formulas, err := os.ReadFile(filepath.Join(root, "docs", "generated", "production-formulas.json"))
	if err != nil {
		t.Fatal(err)
	}
	nextArtifacts := make(map[string][]byte, len(bundle.Artifacts)+1)
	for name, data := range bundle.Artifacts {
		nextArtifacts[name] = bytes.Clone(data)
	}
	nextArtifacts["economy"] = append(nextArtifacts["economy"], '\n')
	nextArtifacts["formulas"] = formulas
	mint := make([]Artifact, 0, len(nextArtifacts))
	for name, data := range nextArtifacts {
		mint = append(mint, Artifact{Name: name, Bytes: data})
	}
	nextID := bundle.Seed.CurrentEpochID + 1
	epoch, err := repository.MintEpoch(ctx, "Public catalog fixture", started.Add(time.Hour), fmt.Sprintf("changelog/epoch-%d.md", nextID), mint)
	if err != nil || epoch.ID != nextID || len(epoch.Hashes) != 1 {
		t.Fatalf("fixture mint=%+v err=%v", epoch, err)
	}
	t.Run("historical bytes never replaced by current files", func(t *testing.T) { assertBundle(t, bundle.Hash, bundle.Artifacts) })
	t.Run("formula-bearing fixture has exact stored bytes", func(t *testing.T) { assertBundle(t, epoch.Hashes[0], nextArtifacts) })
	if _, hasFormulas := bundle.Artifacts["formulas"]; !hasFormulas {
		historical, err := repository.PublicCatalog(ctx, bundle.Hash)
		if err != nil {
			t.Fatal(err)
		}
		for _, artifact := range historical.Artifacts {
			if artifact.Name == "formulas" {
				t.Fatal("current formulas were counterfeited under historical identity")
			}
		}
	}

	// The same accepted hash can belong to multiple epochs. Joining epoch rows
	// directly must not duplicate the response or invalidate the hash.
	if _, err := db.ExecContext(ctx, `INSERT INTO epoch_hashes(epoch_id,constants_hash) VALUES($1,$2)`, nextID, bundle.Hash); err != nil {
		t.Fatal(err)
	}
	t.Run("multiple accepted epochs do not duplicate artifacts", func(t *testing.T) { assertBundle(t, bundle.Hash, bundle.Artifacts) })

	seed := func(hash string, artifacts map[string][]byte, accepted bool) {
		t.Helper()
		if _, err := db.ExecContext(ctx, `INSERT INTO catalog_sets(constants_hash) VALUES($1)`, hash); err != nil {
			t.Fatal(err)
		}
		for name, data := range artifacts {
			if _, err := db.ExecContext(ctx, `INSERT INTO catalog_artifacts(constants_hash,artifact_name,bytes) VALUES($1,$2,$3)`, hash, name, data); err != nil {
				t.Fatal(err)
			}
		}
		if accepted {
			if _, err := db.ExecContext(ctx, `INSERT INTO epoch_hashes(epoch_id,constants_hash) VALUES($1,$2)`, nextID, hash); err != nil {
				t.Fatal(err)
			}
		}
	}
	orphan := "sha256:" + strings.Repeat("1", 64)
	missing := "sha256:" + strings.Repeat("2", 64)
	corrupt := "sha256:" + strings.Repeat("3", 64)
	seed(orphan, bundle.Artifacts, false)
	seed(missing, nil, true)
	seed(corrupt, bundle.Artifacts, true)
	invalidJSON := map[string][]byte{"economy": []byte(`{"unfinished":`)}
	invalidJSONHash, err := save.ConstantsHashArtifacts(invalidJSON)
	if err != nil {
		t.Fatal(err)
	}
	seed(invalidJSONHash, invalidJSON, true)
	for _, test := range []struct {
		name, hash string
		want       error
	}{
		{"unknown hash", "sha256:" + strings.Repeat("0", 64), ErrUnknownPublicCatalog},
		{"unaccepted stored set", orphan, ErrUnknownPublicCatalog},
		{"accepted set missing bytes", missing, ErrInvalidPublicCatalog},
		{"accepted set has wrong identity", corrupt, ErrInvalidPublicCatalog},
		{"hash-matching malformed JSON", invalidJSONHash, ErrInvalidPublicCatalog},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := repository.PublicCatalog(ctx, test.hash)
			if !errors.Is(err, test.want) || !reflect.DeepEqual(got, PublicCatalogBundle{}) {
				t.Fatalf("catalog=%+v err=%v; want %v", got, err, test.want)
			}
		})
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := repository.PublicCatalog(cancelled, bundle.Hash); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled read: %v", err)
	}
}
