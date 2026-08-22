package save

import (
	"bytes"
	"errors"
	"log/slog"
	"strings"
	"testing"
)

func TestSaveRejectionLogUsesBoundedErrorClass(t *testing.T) {
	var output bytes.Buffer
	store := &Store{logger: slog.New(slog.NewJSONHandler(&output, nil))}
	store.logRejection(errors.Join(ErrInvalidState, errors.New("recovery_code=private")), WriteContext{Cause: "intent", IntentID: "0198-private-intent"})
	text := output.String()
	if !strings.Contains(text, `"error_class":"invalid_state"`) {
		t.Fatalf("bounded error class missing: %s", text)
	}
	for _, forbidden := range []string{"recovery_code", "private", "intent_id"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("private save field logged: %q in %s", forbidden, text)
		}
	}
}
