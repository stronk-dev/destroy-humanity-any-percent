package publicread

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"

	"cloud-clicker/server/epochseed"
	"cloud-clicker/server/leaderboard"
	"cloud-clicker/server/meters"
	"cloud-clicker/server/publicapi"
	"cloud-clicker/server/routes"
)

// The real manifest supplies the entire byte population, independently of the
// still-incomplete HTTP descriptor set. This is source verification, not an
// endpoint, epoch acceptance check or public-client journey.
func TestVerifyPublicCatalogBundlePreservesCurrentManifestSources(t *testing.T) {
	source, err := epochseed.Load("../..")
	if err != nil {
		t.Fatal(err)
	}
	input := leaderboard.PublicCatalogBundle{ConstantsHash: source.Hash}
	for _, artifact := range source.Seed.Artifacts {
		data := source.Artifacts[artifact.Name]
		input.Artifacts = append(input.Artifacts, leaderboard.PublicCatalogArtifact{
			Name: artifact.Name, Bytes: data, SHA256: fmt.Sprintf("sha256:%x", sha256.Sum256(data)),
		})
	}
	verified, err := leaderboard.VerifyPublicCatalogBundle(input)
	if err != nil || len(verified.Artifacts) != len(source.Seed.Artifacts) {
		t.Fatal("complete current source verification", err)
	}
	previous := ""
	compacted := 0
	for index, artifact := range verified.Artifacts {
		if artifact.Name <= previous || !bytes.Equal(artifact.Bytes, source.Artifacts[artifact.Name]) {
			t.Fatal("lost source ordering or original bytes", artifact.Name)
		}
		previous = artifact.Name
		var compact bytes.Buffer
		if err := json.Compact(&compact, artifact.Bytes); err != nil {
			t.Fatal(err)
		}
		if bytes.Equal(compact.Bytes(), artifact.Bytes) {
			continue
		}
		compacted++
		candidate := leaderboard.PublicCatalogBundle{ConstantsHash: verified.ConstantsHash,
			Artifacts: append([]leaderboard.PublicCatalogArtifact(nil), verified.Artifacts...)}
		candidate.Artifacts[index].Bytes = compact.Bytes()
		if _, err := leaderboard.VerifyPublicCatalogBundle(candidate); !errors.Is(err, leaderboard.ErrInvalidPublicCatalog) {
			t.Fatal("compacted source retained its original identity", artifact.Name, err)
		}
		// Replacing the child digest cannot repair the framed bundle's identity.
		candidate.Artifacts[index].SHA256 = fmt.Sprintf("sha256:%x", sha256.Sum256(compact.Bytes()))
		if _, err := leaderboard.VerifyPublicCatalogBundle(candidate); !errors.Is(err, leaderboard.ErrInvalidPublicCatalog) {
			t.Fatal("relabelled child bypassed bundle identity", artifact.Name, err)
		}
	}
	if compacted == 0 {
		t.Fatal("source population no longer exercises whitespace loss")
	}
	t.Logf("verified %d original sources; refused compaction and child relabelling for %d", len(verified.Artifacts), compacted)
}

// Construction diagnostic: C18's owner grammar must not be silently narrowed by
// the enclosing HTTP duplicate-key rule. This uses a candidate, not a claim that
// a stored or deployed catalog already contains duplicate members.
func TestCatalogOwnerMemberPolicyConflictsWithHTTPAdmission(t *testing.T) {
	data, err := os.ReadFile("../../balance/categories/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	routeBytes, err := os.ReadFile("../../balance/routes/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	routeCatalog, err := routes.LoadCatalog(routeBytes)
	if err != nil {
		t.Fatal(err)
	}
	gateIDs := []string{}
	for _, gate := range routeCatalog.Gates() {
		gateIDs = append(gateIDs, gate.ID)
	}
	if _, err := leaderboard.LoadCategoryCatalog(data, gateIDs); err != nil {
		t.Fatal("original control", err)
	}
	candidate := bytes.Replace(data, []byte(`"schema_version": 1`), []byte(`"schema_version": 1,"schema_version": 1`), 1)
	if bytes.Equal(candidate, data) {
		t.Fatal("candidate did not change original bytes")
	}
	if _, err := leaderboard.LoadCategoryCatalog(candidate, gateIDs); err != nil {
		t.Fatal("candidate is not owner-admitted", err)
	}
	rows, err := ComposeCatalogSchemas([]string{"categories", "meters"}, []CatalogSchemaOwner{
		{"categories", leaderboard.CategoryCatalogSchemaName, leaderboard.CategorySchemas()},
		{"meters", meters.MeterCatalogSchemaName, meters.MeterSchemas()},
	})
	if err != nil {
		t.Fatal(err)
	}
	definitions, err := publicapi.ValidateSchemaDefinitions(rows)
	if err != nil {
		t.Fatal(err)
	}
	registry, err := publicapi.NewRegistry(rows, []publicapi.Operation{{ID: "catalog_lexical_fixture", Method: "GET", Path: "/api/public/v1/catalog-lexical-fixture", Surface: publicapi.SurfacePublicV1, Auth: publicapi.AuthNone, Public: true,
		Responses: []publicapi.Response{{Kind: publicapi.ResponseSchema, Status: 200, ContentType: publicapi.ContentJSON, SchemaRef: CatalogArtifactSchemaName}},
	}})
	if err != nil {
		t.Fatal(err)
	}
	for name, source := range map[string][]byte{"original": data, "owner-admitted duplicate": candidate} {
		body, err := json.Marshal(map[string]any{"name": "categories", "sha256": fmt.Sprintf("sha256:%x", sha256.Sum256(source)), "json": json.RawMessage(source)})
		if err != nil {
			t.Fatal(err)
		}
		if err := publicapi.ValidateJSON(CatalogArtifactSchemaName, body, definitions); err != nil {
			t.Fatal(name, "structural owner arm", err)
		}
		err = registry.ValidateResponse("catalog_lexical_fixture", 200, body)
		if (err == nil) != (name == "original") {
			t.Fatal(name, "did not expose the owner/wire boundary", err)
		}
	}
}
