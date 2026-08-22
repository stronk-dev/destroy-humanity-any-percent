package production

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestProductionInvariantLogContainsOnlyBoundedKind(t *testing.T) {
	var output bytes.Buffer
	service := &Service{logger: slog.New(slog.NewJSONHandler(&output, nil))}
	service.recordInvariant(InvariantReport{IntentID: "0198-private-intent", Kind: InvariantResidualClamp, Detail: "player-controlled-generator"})
	text := output.String()
	if !strings.Contains(text, string(InvariantResidualClamp)) {
		t.Fatalf("bounded invariant absent: %s", text)
	}
	for _, forbidden := range []string{"private-intent", "player-controlled-generator", "intent_id", "detail"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("private invariant field logged: %q in %s", forbidden, text)
		}
	}
}
