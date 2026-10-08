package production

import (
	"fmt"
	"testing"
)

// SG5 coordinates are integer tokens, not nullable fields. In particular,
// null must not become zero and target a different plot in the live service.
func TestGardenIntentCoordinatesRequireIntegers(t *testing.T) {
	for _, kind := range []string{IntentGardenPlant, IntentGardenUproot, IntentGardenHarvest} {
		for _, field := range []string{"row", "col"} {
			for _, raw := range []string{"null", " \n null \t", "0", "5", "6", "-1", `"0"`, "0.5", "0e0", "true", "{}", "[]"} {
				t.Run(fmt.Sprintf("%s/%s/%q", kind, field, raw), func(t *testing.T) {
					row, col := "0", "0"
					if field == "row" {
						row = raw
					} else {
						col = raw
					}
					fields := fmt.Sprintf(`"row":%s,"col":%s`, row, col)
					switch kind {
					case IntentGardenPlant:
						fields += `,"species_id":"strain_a"`
					case IntentGardenHarvest:
						fields = `"plots":[{` + fields + `}]`
					}
					body := fmt.Sprintf(`{"intent_id":"01986666-6f01-7000-8000-000000000001","kind":%q,"expected_revision":1,%s}`, kind, fields)
					request, err := ParseIntent([]byte(body))
					if err != nil {
						t.Fatalf("coordinate refusal must retain an invalid intent receipt: %v", err)
					}
					valid := raw == "0" || raw == "5"
					wantDetail := field
					if kind == IntentGardenHarvest {
						wantDetail = "plots"
					}
					if valid {
						wantDetail = ""
					}
					if request.InvalidDetail != wantDetail {
						t.Fatalf("coordinate %q: invalid detail=%q want=%q; decoded row=%d col=%d plots=%+v", raw, request.InvalidDetail, wantDetail, request.GardenRow, request.GardenCol, request.GardenPlots)
					}
					if valid {
						gotRow, gotCol := request.GardenRow, request.GardenCol
						if kind == IntentGardenHarvest {
							if len(request.GardenPlots) != 1 {
								t.Fatalf("valid target lost: %+v", request.GardenPlots)
							}
							gotRow, gotCol = request.GardenPlots[0].Row, request.GardenPlots[0].Col
						}
						want := int64(0)
						if raw == "5" {
							want = 5
						}
						if field == "row" && (gotRow != want || gotCol != 0) || field == "col" && (gotCol != want || gotRow != 0) {
							t.Fatalf("valid coordinate changed: row=%d col=%d want %s=%d", gotRow, gotCol, field, want)
						}
					}
				})
			}
		}
	}
}
