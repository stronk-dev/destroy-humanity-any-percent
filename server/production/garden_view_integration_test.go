package production

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"cloud-clicker/server/save"
)

func TestGardenReadDatabaseClockIntegration(t *testing.T) {
	for _, unsalted := range []bool{false, true} {
		for _, offset := range []time.Duration{0, -24 * time.Hour, 24 * time.Hour} {
			t.Run(offset.String()+"/unsalted="+strconvBool(unsalted), func(t *testing.T) {
				fixture := newGardenHarvestFixtureWithState(t, func(founder *save.State) {
					anchor := save.CanonicalServerTime(time.Now().Add(-750 * time.Second)).UnixMilli()
					founder.ServerGarden.TickAnchorWallMS = &anchor
					if unsalted {
						founder.ServerGarden.SaltHex, founder.ServerGarden.TickAnchorWallMS = nil, nil
					}
				})
				before := fixture.refusalSnapshot(t)
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
				if unsalted && response.Garden.TickSeq != 0 || !unsalted && response.Garden.TickSeq < 2 {
					t.Fatal("vacuous clock projection")
				}
				t.Logf("handler_offset=%s unsalted=%t database_bracket=[%d,%d] projected_server_ms=%d tick_seq=%d read_only=true", offset, unsalted, low, high, response.ServerMS, response.Garden.TickSeq)
				if response.ServerMS < low || response.ServerMS > high {
					t.Fatalf("projection used non-DB time: server_ms=%d outside [%d,%d]", response.ServerMS, low, high)
				}
			})
		}
	}
}

func strconvBool(value bool) string {
	if value {
		return "true"
	}
	return "false"
}
