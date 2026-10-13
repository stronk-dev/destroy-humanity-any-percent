package deploymentrelease

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRotationCommandsNeverAppendBeforeLatestLedgerTime(t *testing.T) {
	base := time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC)
	for _, family := range []KeyFamily{FamilyJWT, FamilyBootstrap, FamilyCursor} {
		foreign := FamilyJWT
		if family == FamilyJWT {
			foreign = FamilyBootstrap
		}
		for _, action := range []string{"activate", "remove"} {
			for _, offset := range []struct {
				name  string
				value time.Duration
			}{
				{"earlier", -time.Nanosecond},
				{"equal", 0},
				{"later", time.Nanosecond},
			} {
				t.Run(string(family)+"/"+action+"/"+offset.name, func(t *testing.T) {
					path := filepath.Join(t.TempDir(), "rotation-ledger.jsonl")
					latest := base.Add(time.Minute)
					if action == "remove" {
						if err := ActivateRotation(path, family, "current-2", "current-1", "operator-1", base); err != nil {
							t.Fatal(err)
						}
						overlap, err := MinimumOverlap(family)
						if err != nil {
							t.Fatal(err)
						}
						// Even the earlier arm satisfies this family's minimum.
						// Only cross-family ledger chronology can reject it.
						latest = base.Add(overlap + time.Minute)
					}
					if err := ActivateRotation(path, foreign, "foreign-2", "foreign-1", "operator-1", latest); err != nil {
						t.Fatal(err)
					}
					before, err := os.ReadFile(path)
					if err != nil {
						t.Fatal(err)
					}
					prior, err := ReadRotationLedger(path)
					if err != nil {
						t.Fatal(err)
					}
					mutate := func(at time.Time) error {
						if action == "remove" {
							return RemovePrevious(path, family, "current-2", "current-1", "operator-1", at)
						}
						return ActivateRotation(path, family, "current-2", "current-1", "operator-1", at)
					}
					at := latest.Add(offset.value)
					err = mutate(at)
					after, readErr := os.ReadFile(path)
					records, decodeErr := ReadRotationLedger(path)
					if offset.value < 0 {
						if !errors.Is(err, ErrInvalid) || readErr != nil || !bytes.Equal(before, after) || decodeErr != nil || len(records) != len(prior) {
							t.Fatalf("backdated command poisoned ledger: mutation=%v read=%v decode=%v unchanged=%t records=%d/%d",
								err, readErr, decodeErr, bytes.Equal(before, after), len(records), len(prior))
						}
						// A refused clock must leave the supported command retryable.
						at = latest.Add(time.Nanosecond)
						if err := mutate(at); err != nil {
							t.Fatalf("valid retry after chronology refusal failed: %v", err)
						}
						records, decodeErr = ReadRotationLedger(path)
					} else if err != nil || readErr != nil {
						t.Fatalf("legal ledger time refused: mutation=%v read=%v", err, readErr)
					}
					if decodeErr != nil || len(records) != len(prior)+1 {
						t.Fatalf("accepted command wrote unreadable ledger: records=%d/%d err=%v", len(records), len(prior)+1, decodeErr)
					}
					last := records[len(records)-1]
					wantAction := "activated"
					if action == "remove" {
						wantAction = "removed"
					}
					if last.Family != family || last.Action != wantAction || !last.OccurredAt.Equal(at) || last.CurrentID != "current-2" || last.PreviousID != "current-1" {
						t.Fatalf("accepted command recorded the wrong transition: %+v", last)
					}
				})
			}
		}
	}
}
