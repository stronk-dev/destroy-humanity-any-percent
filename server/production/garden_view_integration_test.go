package production

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"testing"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/save"
)

func TestGardenReadDatabaseClockIntegration(t *testing.T) {
	for _, unsalted := range []bool{false, true} {
		for _, offset := range []time.Duration{0, -24 * time.Hour, 24 * time.Hour} {
			t.Run(offset.String()+"/unsalted="+strconv.FormatBool(unsalted), func(t *testing.T) {
				fixture := newGardenHarvestFixtureWithState(t, func(founder *save.State) {
					anchor := save.CanonicalServerTime(time.Now().Add(-750 * time.Second)).UnixMilli()
					founder.ServerGarden.TickAnchorWallMS = &anchor
					if unsalted {
						founder.ServerGarden.SaltHex, founder.ServerGarden.TickAnchorWallMS = nil, nil
					}
				})
				before := fixture.refusalSnapshot(t)
				loaded, err := fixture.store.LoadLatest(fixture.ctx, fixture.founderStreamID)
				if err != nil {
					t.Fatal(err)
				}
				headLow := fixture.databaseMS(t)
				head, stamp, err := fixture.store.LoadSiblingLatestAtDatabaseTime(fixture.ctx, fixture.companyStreamID, economy.ScopeFounder)
				headHigh := fixture.databaseMS(t)
				if err != nil || !reflect.DeepEqual(head, loaded) || stamp < headLow || stamp > headHigh {
					t.Fatalf("database-stamped read changed the loaded head or clock: head=%+v expected=%+v stamp=%d bracket=[%d,%d] err=%v", head, loaded, stamp, headLow, headHigh, err)
				}
				low := fixture.databaseMS(t)
				encoded, err := fixture.service.GardenView(fixture.ctx, fixture.companyStreamID, time.Now().Add(offset))
				high := fixture.databaseMS(t)
				if err != nil {
					t.Fatal(err)
				}
				var response struct {
					Kind            string     `json:"kind"`
					ServerMS        int64      `json:"server_ms"`
					FounderRevision int64      `json:"founder_revision"`
					Garden          gardenView `json:"garden"`
				}
				if err := json.Unmarshal(encoded, &response); err != nil || response.Kind != "active" {
					t.Fatalf("invalid read: %s err=%v", encoded, err)
				}
				after := fixture.refusalSnapshot(t)
				if !reflect.DeepEqual(before, after) {
					t.Fatalf("read changed persistence: before=%+v after=%+v", before, after)
				}
				if response.FounderRevision != before.State.FounderRevision {
					t.Fatal("projection revision is not saved head")
				}
				t.Logf("handler_offset=%s unsalted=%t database_bracket=[%d,%d] projected_server_ms=%d tick_seq=%d read_only=true", offset, unsalted, low, high, response.ServerMS, response.Garden.TickSeq)
				if response.ServerMS < low || response.ServerMS > high {
					t.Fatalf("projection used non-DB time: server_ms=%d outside [%d,%d]", response.ServerMS, low, high)
				}
				if unsalted && response.Garden.TickSeq != 0 || !unsalted && response.Garden.TickSeq < 2 {
					t.Fatal("vacuous clock projection")
				}
			})
		}
	}
}

func TestGardenReadDatabaseFailureIntegration(t *testing.T) {
	fixture := newGardenHarvestFixture(t)
	before := fixture.refusalSnapshot(t)
	cancelled, cancel := context.WithCancel(fixture.ctx)
	cancel()
	if encoded, err := fixture.service.GardenView(cancelled, fixture.companyStreamID, time.Now()); !errors.Is(err, context.Canceled) || len(encoded) != 0 {
		t.Fatalf("cancelled DB query fell back to handler time: response=%s err=%v", encoded, err)
	}
	for _, streamID := range []string{"not-a-stream", "00000000-0000-4000-8000-000000000000"} {
		encoded, err := fixture.service.GardenView(fixture.ctx, streamID, time.Now())
		if len(encoded) != 0 || !errors.Is(err, save.ErrInvalidStream) && !errors.Is(err, save.ErrNotFound) {
			t.Fatalf("invalid/missing source produced a view: stream=%q response=%s err=%v", streamID, encoded, err)
		}
	}
	if after := fixture.refusalSnapshot(t); !reflect.DeepEqual(before, after) {
		t.Fatal("failed reads changed persistence")
	}
}
