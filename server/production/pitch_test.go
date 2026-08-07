package production

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"cloud-clicker/server/copykeys"
	"cloud-clicker/server/fiscal"
	"cloud-clicker/server/minigame"
	"cloud-clicker/server/pitch"
	"cloud-clicker/server/save"
)

func pitchFeatureBundle(t *testing.T) CatalogBundle {
	t.Helper()
	bundle := recoverySoulBundle(t)
	bundle.Artifacts = cloneArtifactMap(bundle.Artifacts)
	minigameBytes, err := os.ReadFile("../../testdata/minigame/pitch-v3.json")
	if err != nil {
		t.Fatal(err)
	}
	minigameCatalog, err := minigame.LoadCatalog(minigameBytes)
	if err != nil {
		t.Fatal(err)
	}
	var fiscalRoot map[string]any
	if err := json.Unmarshal(bundle.Artifacts["fiscal"], &fiscalRoot); err != nil {
		t.Fatal(err)
	}
	rows, ok := fiscalRoot["unlock_rows"].([]any)
	if !ok {
		t.Fatal("Fiscal fixture has no unlock_rows")
	}
	fiscalRoot["unlock_rows"] = append([]any{map[string]any{"unlock_id": "minigame.pitch", "cost": float64(3)}}, rows...)
	fiscalBytes, err := json.Marshal(fiscalRoot)
	if err != nil {
		t.Fatal(err)
	}
	fiscalCatalog, err := fiscal.LoadCatalog(fiscalBytes, bundle.Economy)
	if err != nil {
		t.Fatal(err)
	}
	pitchBytes, err := os.ReadFile("../../balance/testdata/pitch-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	keys := make(map[string]struct{})
	for _, key := range copykeys.All() {
		keys[key] = struct{}{}
	}
	pitchCatalog, err := pitch.LoadCatalog(pitchBytes, pitch.Declarations{CopyKeys: keys})
	if err != nil {
		t.Fatal(err)
	}
	bundle.Artifacts["minigames"], bundle.Artifacts["fiscal"], bundle.Artifacts["pitch"] = minigameBytes, fiscalBytes, pitchBytes
	bundle.ConstantsHash, err = save.ConstantsHashArtifacts(bundle.Artifacts)
	if err != nil {
		t.Fatal(err)
	}
	bundle.Minigames, bundle.Fiscal, bundle.Pitch = minigameCatalog, fiscalCatalog, pitchCatalog
	if !bundle.valid(bundle.ConstantsHash) {
		t.Fatal("Pitch fixture bundle is not internally valid")
	}
	return bundle
}

func TestPitchBundleResolvesImmutableTenantContent(t *testing.T) {
	bundle := pitchFeatureBundle(t)
	set := ReplayCatalogSet{bundle.ConstantsHash: bundle}
	content, ok := set.ResolveTenantContent(bundle.ConstantsHash, pitch.EngineRef, pitch.EngineVersion)
	if !ok || content.Hash != pitch.ContentHash(bundle.Artifacts["pitch"]) || content.SchemaVersion != pitch.SchemaVersion ||
		!bytes.Equal(content.Bytes, bundle.Artifacts["pitch"]) {
		t.Fatalf("content=%+v ok=%v", content, ok)
	}
	content.Bytes[0] ^= 1
	again, ok := set.ResolveTenantContent(bundle.ConstantsHash, pitch.EngineRef, pitch.EngineVersion)
	if !ok || !bytes.Equal(again.Bytes, bundle.Artifacts["pitch"]) {
		t.Fatal("tenant content resolver leaked mutable artifact bytes")
	}
	for _, mismatch := range [][3]string{{"sha256:bad", pitch.EngineRef, pitch.EngineVersion}, {bundle.ConstantsHash, "other", pitch.EngineVersion}, {bundle.ConstantsHash, pitch.EngineRef, "2.0.0"}} {
		if _, ok := set.ResolveTenantContent(mismatch[0], mismatch[1], mismatch[2]); ok {
			t.Fatalf("resolved mismatched identity %v", mismatch)
		}
	}
	definition, found := bundle.Minigames.Definition("pitch")
	if !found || definition.Unlock.Kind != "fiscal_unlock" || definition.Unlock.UnlockID != "minigame.pitch" {
		t.Fatalf("definition=%+v found=%v", definition, found)
	}
}
