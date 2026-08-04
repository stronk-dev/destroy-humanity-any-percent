package save

import (
	"errors"
	"testing"
)

func TestExitVersionActivationIsAtomicAndLegacyMigrates(t *testing.T) {
	legacyFounder := &State{WireVersion: CurrentVersion}
	legacyFinal := &State{WireVersion: CurrentVersion}
	legacyNext := &State{WireVersion: CurrentVersion}
	v16Founder := &State{WireVersion: 16}
	v16Next := &State{WireVersion: 16}

	if err := validateExitVersionTransition(Revision{Version: 1}, Revision{Version: 13}, legacyFounder, legacyFinal, legacyNext); err != nil {
		t.Fatalf("legacy Exit migration rejected: %v", err)
	}
	if err := validateExitVersionTransition(Revision{Version: 14}, Revision{Version: 14}, v16Founder, legacyFinal, v16Next); err != nil {
		t.Fatalf("atomic v16 activation rejected: %v", err)
	}
	tests := []struct {
		name                 string
		founderRevision      int
		companyRevision      int
		founder, final, next *State
	}{
		{name: "company activates without founder", founderRevision: 14, companyRevision: 14, founder: legacyFounder, final: legacyFinal, next: v16Next},
		{name: "founder activates without company", founderRevision: 14, companyRevision: 14, founder: v16Founder, final: legacyFinal, next: legacyNext},
		{name: "terminal version changes", founderRevision: 14, companyRevision: 14, founder: v16Founder, final: v16Next, next: v16Next},
		{name: "preexisting tuple mismatch", founderRevision: 16, companyRevision: 14, founder: v16Founder, final: legacyFinal, next: v16Next},
		{name: "standalone v15 tuple", founderRevision: 15, companyRevision: 15, founder: &State{WireVersion: 15}, final: &State{WireVersion: 15}, next: &State{WireVersion: 15}},
	}
	for _, item := range tests {
		t.Run(item.name, func(t *testing.T) {
			if err := validateExitVersionTransition(Revision{Version: item.founderRevision}, Revision{Version: item.companyRevision}, item.founder, item.final, item.next); !errors.Is(err, ErrInvalidState) {
				t.Fatalf("error=%v", err)
			}
		})
	}
}
