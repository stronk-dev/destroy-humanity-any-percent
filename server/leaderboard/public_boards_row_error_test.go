package leaderboard

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"strings"
	"testing"
)

// Inject failure at Rows.Next, after QueryContext succeeded. A query error
// alone cannot exercise the reader's missing rows.Err check.
type publicBoardRowFault struct{ terminal error }

func (fault publicBoardRowFault) Connect(context.Context) (driver.Conn, error) { return fault, nil }
func (fault publicBoardRowFault) Driver() driver.Driver                        { return fault }
func (fault publicBoardRowFault) Open(string) (driver.Conn, error)             { return fault, nil }
func (fault publicBoardRowFault) Close() error                                 { return nil }
func (fault publicBoardRowFault) Begin() (driver.Tx, error) {
	return nil, errors.New("unexpected transaction in public board read fixture")
}
func (fault publicBoardRowFault) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("unexpected prepared statement in public board read fixture")
}
func (fault publicBoardRowFault) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	switch {
	case strings.Contains(query, "SELECT EXISTS(SELECT 1 FROM epochs"):
		return &publicBoardFaultRows{column: "exists", row: []driver.Value{true}}, nil
	case strings.Contains(query, "SELECT hashes.constants_hash FROM epoch_hashes"):
		return &publicBoardFaultRows{column: "constants_hash", terminal: fault.terminal}, nil
	default:
		return nil, errors.New("unexpected query in public board read fixture")
	}
}

type publicBoardFaultRows struct {
	column   string
	row      []driver.Value
	terminal error
}

func (rows *publicBoardFaultRows) Columns() []string { return []string{rows.column} }
func (rows *publicBoardFaultRows) Close() error      { return nil }
func (rows *publicBoardFaultRows) Next(values []driver.Value) error {
	if rows.row != nil {
		copy(values, rows.row)
		rows.row = nil
		return nil
	}
	if rows.terminal != nil {
		return rows.terminal
	}
	return io.EOF
}

func TestPublicBoardRankingKindPropagatesRowErrors(t *testing.T) {
	interrupted := errors.New("injected catalog hash row failure")
	for _, test := range []struct {
		name           string
		rowError, want error
	}{
		{"completed empty read is unknown category", nil, ErrUnknownPublicCategory},
		{"row failure is not unknown category", interrupted, interrupted},
		{"cancelled row read is not unknown category", context.Canceled, context.Canceled},
	} {
		t.Run(test.name, func(t *testing.T) {
			db := sql.OpenDB(publicBoardRowFault{terminal: test.rowError})
			t.Cleanup(func() { _ = db.Close() })
			repository := &Repository{db: db}
			kind, err := repository.PublicBoardRankingKind(context.Background(), "any_percent", 8)
			if kind != "" || !errors.Is(err, test.want) {
				t.Fatalf("ranking kind=%q error=%v; want original error %v", kind, err, test.want)
			}
		})
	}
}
