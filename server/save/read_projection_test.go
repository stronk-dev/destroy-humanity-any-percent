package save

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"cloud-clicker/server/economy"
)

func TestDatabaseProjectionReadRejectsInvalidInputs(t *testing.T) {
	const validID = "00000000-0000-4000-8000-000000000000"
	for _, test := range []struct {
		name   string
		store  *Store
		stream string
		scope  economy.Scope
	}{
		{"nil store", nil, validID, economy.ScopeFounder},
		{"nil database", &Store{}, validID, economy.ScopeFounder},
		{"invalid source", &Store{db: &sql.DB{}}, "invalid", economy.ScopeFounder},
		{"invalid scope", &Store{db: &sql.DB{}}, validID, economy.ScopeWorld},
	} {
		t.Run(test.name, func(t *testing.T) {
			loaded, stamp, err := test.store.LoadSiblingLatestAtDatabaseTime(context.Background(), test.stream, test.scope)
			if !errors.Is(err, ErrInvalidStream) || stamp != 0 || loaded.State != nil {
				t.Fatalf("invalid input yielded projection context: %+v stamp=%d err=%v", loaded, stamp, err)
			}
		})
	}
}
