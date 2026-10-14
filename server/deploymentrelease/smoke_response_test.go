package deploymentrelease

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

const smokeResponseFixtureLimit = 4 << 20 // existing helper ceiling, not a new budget

func smokePaddedJSON(prefix string, size int) string {
	return prefix + strings.Repeat(" ", size-len(prefix))
}

func TestSmokeResponseRejectsArtificialLimitEOF(t *testing.T) {
	for _, suffix := range []string{" ", "x", `{"value":2}`} {
		t.Run(fmt.Sprintf("suffix_%q", suffix), func(t *testing.T) {
			body := smokePaddedJSON(`{"value":1}`, smokeResponseFixtureLimit) + suffix
			var reply struct {
				Value int `json:"value"`
			}
			if err := decodeResponse(strings.NewReader(body), &reply); !errors.Is(err, ErrInvalid) {
				t.Fatalf("oversized or hidden trailing reply accepted: bytes=%d err=%v", len(body), err)
			}
		})
	}
}

func TestSmokeResponsePreservesRealEOFAtAndBelowLimit(t *testing.T) {
	for _, size := range []int{len(`{"value":1}`), smokeResponseFixtureLimit - 1, smokeResponseFixtureLimit} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			var reply struct {
				Value int `json:"value"`
			}
			if err := decodeResponse(strings.NewReader(smokePaddedJSON(`{"value":1}`, size)), &reply); err != nil || reply.Value != 1 {
				t.Fatalf("complete bounded reply refused: bytes=%d value=%d err=%v", size, reply.Value, err)
			}
		})
	}
	for _, body := range []string{`{"value":`, `{"value":1}x`, `{"value":1}{"value":2}`} {
		var reply struct {
			Value int `json:"value"`
		}
		if err := decodeResponse(strings.NewReader(body), &reply); err == nil {
			t.Fatal("malformed/trailing bounded reply accepted")
		}
	}
}

type smokeEndlessPadding struct{ bytesRead int }

func (padding *smokeEndlessPadding) Read(buffer []byte) (int, error) {
	for index := range buffer {
		buffer[index] = ' '
	}
	padding.bytesRead += len(buffer)
	return len(buffer), nil
}

func TestSmokeResponseBoundsAnEndlessTail(t *testing.T) {
	prefix := `{"value":1}`
	padding := &smokeEndlessPadding{}
	var reply struct {
		Value int `json:"value"`
	}
	if err := decodeResponse(io.MultiReader(strings.NewReader(prefix), padding), &reply); !errors.Is(err, ErrInvalid) {
		t.Fatalf("endless reply accepted: %v", err)
	}
	if padding.bytesRead != smokeResponseFixtureLimit+1-len(prefix) {
		t.Fatalf("sentinel read escaped the existing bound: bytes=%d", padding.bytesRead)
	}
}

func TestSmokeResponseRefusesBeforeAuthenticatedFollowup(t *testing.T) {
	hash := "sha256:" + strings.Repeat("a", 64)
	body := fmt.Sprintf(`{"session":{"access_token":"fixture-access"},"game_ui_snapshot":{"constants_hash":%q,"run":{"founder_id":"01986666-0000-7000-8000-000000000001"}}}`, hash)
	for _, oversized := range []bool{false, true} {
		t.Run(map[bool]string{false: "complete_control", true: "truncated_by_old_reader"}[oversized], func(t *testing.T) {
			var bootstrapCalls, stateCalls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/api/v1/bootstrap":
					bootstrapCalls.Add(1)
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusCreated)
					data := body
					if oversized {
						data = smokePaddedJSON(body, smokeResponseFixtureLimit) + "x"
					}
					_, _ = io.WriteString(w, data)
				case "/api/v1/founder/state":
					stateCalls.Add(1)
					if r.Header.Get("Authorization") != "Bearer fixture-access" {
						t.Error("control credential dispatch changed")
					}
					// The positive control deliberately stops here, before socket dial.
					w.WriteHeader(http.StatusServiceUnavailable)
				default:
					t.Error("unexpected smoke request")
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()
			session, err := OpenAuthenticatedSmoke(context.Background(), server.Client(), server.URL, hash)
			if session != nil || !errors.Is(err, ErrInvalid) || bootstrapCalls.Load() != 1 {
				t.Fatalf("invalid smoke outcome: bootstrap=%d err=%v", bootstrapCalls.Load(), err)
			}
			want := 1
			if oversized {
				want = 0
			}
			if stateCalls.Load() != int32(want) {
				t.Fatalf("authenticated state followups=%d want=%d", stateCalls.Load(), want)
			}
		})
	}
}
