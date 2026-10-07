package gameui

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"testing"
	"time"

	"cloud-clicker/server/production"
)

func fiscalSharedSnapshot(t *testing.T) json.RawMessage {
	t.Helper()
	bundle, company, founder, _ := reputationProjectionSource(t)
	const ownerID = "01986666-f200-4000-8000-000000000001"
	base := time.UnixMilli(1_800_000_000_000).UTC()
	company.EvaluatedThrough, company.ManualTokenRefilledAt, company.RunStartedAt = base, base, base
	founder.EvaluatedThrough, founder.ManualTokenRefilledAt = base, base
	founder.FiscalPeriodOpenedWallMS, founder.FiscalPeriodSequence, founder.FiscalCredit = base.UnixMilli(), 17, 4
	spawn, err := bundle.Opportunities.Spawn(ownerID, company.RunSeq, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	company.OpportunitySpawnSeq, company.NextOpportunityAttendedMS = 0, spawn.SpawnedAttendedMS
	if err := bundle.ValidateFoundationState(company); err != nil {
		t.Fatal(err)
	}
	if err := bundle.ValidateFoundationState(founder); err != nil {
		t.Fatal(err)
	}
	frozen, err := production.FrozenFounderContributions(bundle, founder)
	if err != nil {
		t.Fatal(err)
	}
	contributions, err := production.ResolveFrozenContributions(bundle.Economy, frozen)
	if err != nil {
		t.Fatal(err)
	}
	now := base.Add(time.Duration(2*bundle.Fiscal.Clock.AutoMS+bundle.Fiscal.Clock.GuaranteedMS) * time.Millisecond)
	raw, err := projectSnapshot(bundle, ownerID, 3, 7, company, founder, contributions, now, false)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestFiscalSharedSnapshot(t *testing.T) {
	raw, err := os.ReadFile("../../testdata/gameui/fiscal-snapshot-v4.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture snapshot
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SchemaVersion != 4 || fixture.Features.Fiscal == nil || fixture.Features.Fiscal.Credit != 4 ||
		fixture.Features.Fiscal.SweepPreview != (fiscalSweepPreview{CreditAfter: 10, Credited: 6, Periods: 2}) {
		t.Fatal("shared fixture lost its nonempty Fiscal preview")
	}
	var want, got any
	if err := json.Unmarshal(raw, &want); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(fiscalSharedSnapshot(t), &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatal("shared Fiscal snapshot differs from the complete Go projection")
	}
}
