package save

import (
	"errors"
	"testing"
)

func TestExitVersionAxesUseIndependentPinnedFloors(t *testing.T) {
	legacyFounder := &State{WireVersion: CurrentVersion}
	legacyFinal := &State{WireVersion: CurrentVersion}
	legacyNext := &State{WireVersion: CurrentVersion}
	v16Founder := &State{WireVersion: 16}
	v16Next := &State{WireVersion: 16}
	legacyFloors := ExitVersionFloors{CurrentFounder: 14, CurrentCompany: 14, NextFounder: 14, NextCompany: 14}
	pairedV16Floors := ExitVersionFloors{CurrentFounder: 14, CurrentCompany: 14, NextFounder: 16, NextCompany: 16}

	if err := validateExitVersionTransition(Revision{Version: 1}, Revision{Version: 13}, legacyFounder, legacyFinal, legacyNext, legacyFloors); err != nil {
		t.Fatalf("legacy Exit migration rejected: %v", err)
	}
	if err := validateExitVersionTransition(Revision{Version: 14}, Revision{Version: 14}, v16Founder, legacyFinal, v16Next, pairedV16Floors); err != nil {
		t.Fatalf("atomic v16 activation rejected: %v", err)
	}
	if err := validateExitVersionTransition(Revision{Version: 16}, Revision{Version: 14}, v16Founder, legacyFinal, legacyNext,
		ExitVersionFloors{CurrentFounder: 16, CurrentCompany: 14, NextFounder: 16, NextCompany: 14}); err != nil {
		t.Fatalf("independent mixed-axis Exit rejected: %v", err)
	}
	tests := []struct {
		name                 string
		founderRevision      int
		companyRevision      int
		founder, final, next *State
		floors               ExitVersionFloors
	}{
		{name: "company misses next floor", founderRevision: 14, companyRevision: 14, founder: v16Founder, final: legacyFinal, next: legacyNext, floors: pairedV16Floors},
		{name: "founder misses next floor", founderRevision: 14, companyRevision: 14, founder: legacyFounder, final: legacyFinal, next: v16Next, floors: pairedV16Floors},
		{name: "terminal version changes", founderRevision: 14, companyRevision: 14, founder: v16Founder, final: v16Next, next: v16Next, floors: pairedV16Floors},
		{name: "current founder below floor", founderRevision: 14, companyRevision: 14, founder: v16Founder, final: legacyFinal, next: legacyNext, floors: ExitVersionFloors{CurrentFounder: 16, CurrentCompany: 14, NextFounder: 16, NextCompany: 14}},
		{name: "standalone v15 tuple", founderRevision: 15, companyRevision: 15, founder: &State{WireVersion: 15}, final: &State{WireVersion: 15}, next: &State{WireVersion: 15}, floors: ExitVersionFloors{CurrentFounder: 15, CurrentCompany: 15, NextFounder: 15, NextCompany: 15}},
		{name: "missing declared floors", founderRevision: 14, companyRevision: 14, founder: legacyFounder, final: legacyFinal, next: legacyNext},
	}
	for _, item := range tests {
		t.Run(item.name, func(t *testing.T) {
			if err := validateExitVersionTransition(Revision{Version: item.founderRevision}, Revision{Version: item.companyRevision}, item.founder, item.final, item.next, item.floors); !errors.Is(err, ErrInvalidState) {
				t.Fatalf("error=%v", err)
			}
		})
	}
}
