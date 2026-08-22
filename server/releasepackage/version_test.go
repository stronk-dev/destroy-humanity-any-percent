package releasepackage

import "testing"

func TestReleaseVersionValidationAndOrdering(t *testing.T) {
	for _, fixture := range []struct {
		left, right string
		want        int
	}{
		{"1.0.1", "1.0.0", 1},
		{"2.0.0", "10.0.0", -1},
		{"1.0.0", "1.0.0-preview.9", 1},
		{"1.0.0-preview.10", "1.0.0-preview.2", 1},
		{"1.0.0-preview.1", "1.0.0-preview.alpha", -1},
		{"1.0.0-preview.1", "1.0.0-preview.1", 0},
	} {
		got, err := CompareReleaseVersions(fixture.left, fixture.right)
		if err != nil || got != fixture.want {
			t.Fatalf("compare %s %s = %d, %v; want %d", fixture.left, fixture.right, got, err, fixture.want)
		}
	}
	for _, invalid := range []string{"1.0.0-", "01.0.0", "1.00.0", "1.0.0-preview.01", "1.0", "v1.0.0"} {
		if validReleaseVersion(invalid) {
			t.Fatalf("invalid semantic version accepted: %s", invalid)
		}
	}
}
