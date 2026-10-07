package gameui

import (
	"errors"
	"reflect"
	"testing"
)

// GS3-A1 producer control. This exercises the existing real projector and
// pinned catalog, not a database or browser. The client admission proof is
// separate; both use the catalog's IDs rather than a second literal list.
func TestMetersProjectionUsesCompleteCatalogIDs(t *testing.T) {
	bundle := pinnedBundle(t)
	company, _ := featureStates(bundle)
	arm, err := projectMeters(bundle.Meters, company)
	if err != nil {
		t.Fatal(err)
	}
	ids := make([]string, 0, len(arm.Meters))
	for _, row := range arm.Meters {
		ids = append(ids, row.MeterID)
	}
	if !reflect.DeepEqual(ids, bundle.Meters.MeterIDs()) {
		t.Fatalf("projected IDs %v differ from the pinned catalog %v", ids, bundle.Meters.MeterIDs())
	}
	for _, id := range bundle.Meters.MeterIDs() {
		t.Run("missing_"+id, func(t *testing.T) {
			company, _ := featureStates(bundle)
			delete(company.MeterValues, id)
			if projected, err := projectMeters(bundle.Meters, company); !errors.Is(err, ErrInvalidProjection) || projected != nil {
				t.Fatalf("missing saved value %s yielded arm=%+v err=%v", id, projected, err)
			}
		})
	}
}
