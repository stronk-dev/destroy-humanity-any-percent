package replaycatalog

import (
	"os"
	"testing"

	"cloud-clicker/server/epochseed"
	"cloud-clicker/server/save"
)

func TestCurriculumCandidateExtendsTheCompleteEpochChain(t *testing.T) {
	seed, err := epochseed.Load("../..")
	if err != nil {
		t.Fatal(err)
	}
	artifacts := make(map[string][]byte, len(seed.Artifacts)+1)
	for name, data := range seed.Artifacts {
		artifacts[name] = append([]byte(nil), data...)
	}
	artifacts["curriculum"], err = os.ReadFile("../../balance/testdata/t0-t1/curriculum-v2.json")
	if err != nil {
		t.Fatal(err)
	}
	hash, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := Load(hash, artifacts)
	if err != nil || bundle.Curriculum == nil || bundle.Curriculum.FirstFailure.Branches[1].Branch != "burnout" {
		t.Fatalf("curriculum bundle=%+v err=%v", bundle.Curriculum, err)
	}
	delete(artifacts, "relevance")
	invalidHash, _ := save.ConstantsHashArtifacts(artifacts)
	if _, err := Load(invalidHash, artifacts); err == nil {
		t.Fatal("curriculum loaded without its preceding relevance artifact")
	}
}
