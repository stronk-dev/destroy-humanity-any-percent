package commonsprojection

import "testing"

func TestMergeCapacityIsFloorTriggeredWholeAndBounded(t *testing.T) {
	tests := []struct {
		name          string
		targetMembers int
		sourceMembers int
		want          bool
	}{
		{name: "exact 1.5x ceiling", targetMembers: 200, sourceMembers: 25, want: true},
		{name: "one above ceiling", targetMembers: 200, sourceMembers: 26, want: false},
		{name: "source at floor", targetMembers: 150, sourceMembers: 40, want: false},
		{name: "below floor with capacity", targetMembers: 150, sourceMembers: 39, want: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := canMergeCohort(test.targetMembers, test.sourceMembers, 40, 150); got != test.want {
				t.Fatalf("canMergeCohort(%d,%d)=%v want=%v", test.targetMembers, test.sourceMembers, got, test.want)
			}
		})
	}
}
