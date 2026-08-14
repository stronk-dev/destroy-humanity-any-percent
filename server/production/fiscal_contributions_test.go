package production

import (
	"testing"
	"time"

	"cloud-clicker/server/save"
)

type fixedReplayBundleResolver struct{ bundle CatalogBundle }

func (resolver fixedReplayBundleResolver) ResolveReplayCatalogs(hash string) (CatalogBundle, bool) {
	return resolver.bundle, hash == resolver.bundle.ConstantsHash
}

func TestFrozenFiscalContributionsAndNewFounderActivation(t *testing.T) {
	legacy, foundations := foundationTestBundles(t)
	_, pets := founderFeatureBundles(t, foundations)
	bundle := fiscalFeatureBundle(t, pets)
	now := time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)
	founder := replayFounderFixtureState(t, legacy, now)
	founder.WireVersion = save.CurrentVersion
	company := replayFixtureState(t, legacy.Economy, now)
	initializer := FounderInitializer{Catalogs: fixedReplayBundleResolver{bundle: bundle}}
	values, err := initializer.InitializeNewFounder(bundle.ConstantsHash, "01986666-f101-7000-8000-000000000001", now, founder, company)
	if err != nil {
		t.Fatal(err)
	}
	if save.VersionForState(founder) != 19 || save.VersionForState(company) != 16 ||
		founder.FiscalPeriodOpenedWallMS != now.UnixMilli() || len(values) != len(bundle.Fiscal.GeneratorLevelRows())+1 {
		t.Fatalf("founder version=%d company=%d opened=%d values=%+v", save.VersionForState(founder),
			save.VersionForState(company), founder.FiscalPeriodOpenedWallMS, values)
	}
	for _, value := range values {
		if value.Factor != "1e0" {
			t.Fatalf("new Founder contribution was not neutral: %+v", value)
		}
	}
	founder.FiscalCredit = 10
	founder.FiscalGeneratorLevels["generator.beige_tower"] = 2
	values, err = FrozenFiscalContributions(bundle.Fiscal, founder)
	if err != nil || len(values) != 2 || values[0].SourceID >= values[1].SourceID ||
		values[0].Factor == "1e0" || values[1].Factor == "1e0" {
		t.Fatalf("retuned contributions=%+v err=%v", values, err)
	}
}
