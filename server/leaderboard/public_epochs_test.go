package leaderboard

import (
	"reflect"
	"testing"
)

func TestPublicEpochHashesAreByteSortedAndNeverNil(t *testing.T) {
	if hashes := publicEpochHashes(""); hashes == nil || len(hashes) != 0 {
		t.Fatalf("empty aggregate = %#v", hashes)
	}
	got := publicEpochHashes("sha256:c,sha256:0,sha256:B,sha256:b")
	if want := []string{"sha256:0", "sha256:B", "sha256:b", "sha256:c"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("hashes = %v, want byte order %v", got, want)
	}
}
