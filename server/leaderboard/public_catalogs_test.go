package leaderboard

import (
	"bytes"
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

	"cloud-clicker/server/save"
)

func TestPublicCatalogBundleExactBytesAndRefusals(t *testing.T) {
	input := []Artifact{{Name: "z", Bytes: []byte("{\n  \"value\": 1\n}\n")}, {Name: "a", Bytes: []byte(`[1,2]`)}}
	hash, err := save.ConstantsHashArtifacts(map[string][]byte{"a": input[1].Bytes, "z": input[0].Bytes})
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := publicCatalogBundle(hash, input)
	if err != nil || bundle.ConstantsHash != hash || len(bundle.Artifacts) != 2 {
		t.Fatalf("bundle=%+v err=%v", bundle, err)
	}
	for i, source := range []Artifact{input[1], input[0]} {
		artifact := bundle.Artifacts[i]
		if artifact.Name != source.Name || !bytes.Equal(artifact.Bytes, source.Bytes) || artifact.SHA256 != save.ConstantsHash(source.Bytes) {
			t.Fatalf("artifact %d lost order, exact bytes or digest: %+v", i, artifact)
		}
	}
	// Returning the source does not lend a caller ownership of input buffers.
	bundle.Artifacts[0].Bytes[0] = '!'
	if string(input[1].Bytes) != `[1,2]` {
		t.Fatal("returned buffer aliases source")
	}

	for _, test := range []struct {
		name      string
		hash      string
		artifacts []Artifact
	}{
		{"empty accepted bundle", hash, nil},
		{"missing artifact", hash, input[:1]},
		{"wrong identity", "sha256:" + strings.Repeat("0", 64), input},
		{"changed valid JSON bytes", hash, []Artifact{{Name: "a", Bytes: []byte(`[1,3]`)}, input[0]}},
		{"extra artifact", hash, append(append([]Artifact{}, input...), Artifact{Name: "extra", Bytes: []byte(`{}`)})},
		{"duplicate name", hash, []Artifact{input[0], input[0]}},
		{"invalid name", hash, []Artifact{{Name: "Invalid", Bytes: []byte(`{}`)}}},
		{"empty bytes", hash, []Artifact{{Name: "a"}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := publicCatalogBundle(test.hash, test.artifacts)
			if !errors.Is(err, ErrInvalidPublicCatalog) || !reflect.DeepEqual(got, PublicCatalogBundle{}) {
				t.Fatalf("got=%+v err=%v", got, err)
			}
		})
	}
	for _, data := range [][]byte{[]byte(`{`), []byte(`{} {}`), {'"', 0xff, '"'}} {
		hash, err := save.ConstantsHashArtifacts(map[string][]byte{"a": data})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := publicCatalogBundle(hash, []Artifact{{Name: "a", Bytes: data}}); !errors.Is(err, ErrInvalidPublicCatalog) {
			t.Fatalf("hash-matching invalid JSON admitted: %q err=%v", data, err)
		}
	}
}

// Exercise failure after successful QueryContext, including after a valid row.
// No transaction or write operation is supported by this read-only fixture.
type publicCatalogFault struct {
	data     []byte
	terminal error
}

func (fault publicCatalogFault) Connect(context.Context) (driver.Conn, error) { return fault, nil }
func (fault publicCatalogFault) Driver() driver.Driver                        { return fault }
func (fault publicCatalogFault) Open(string) (driver.Conn, error)             { return fault, nil }
func (fault publicCatalogFault) Close() error                                 { return nil }
func (fault publicCatalogFault) Begin() (driver.Tx, error) {
	return nil, errors.New("unexpected catalog write transaction")
}
func (fault publicCatalogFault) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("unexpected catalog prepared statement")
}
func (fault publicCatalogFault) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if !strings.Contains(query, "FROM epoch_hashes WHERE constants_hash=$1") || len(args) != 1 {
		return nil, errors.New("unexpected catalog query")
	}
	return &publicCatalogFaultRows{data: fault.data, terminal: fault.terminal}, nil
}

type publicCatalogFaultRows struct {
	data     []byte
	terminal error
}

func (rows *publicCatalogFaultRows) Columns() []string { return []string{"artifact_name", "bytes"} }
func (rows *publicCatalogFaultRows) Close() error      { return nil }
func (rows *publicCatalogFaultRows) Next(values []driver.Value) error {
	if rows.data != nil {
		values[0], values[1] = "a", rows.data
		rows.data = nil
		return nil
	}
	if rows.terminal != nil {
		return rows.terminal
	}
	return io.EOF
}

func TestPublicCatalogRowFailuresDoNotPublishPartialEvidence(t *testing.T) {
	data := []byte(`{}`)
	hash, err := save.ConstantsHashArtifacts(map[string][]byte{"a": data})
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name           string
		data           []byte
		terminal, want error
	}{
		{"empty read is unknown", nil, nil, ErrUnknownPublicCatalog},
		{"interrupted before row", nil, io.ErrUnexpectedEOF, io.ErrUnexpectedEOF},
		{"interrupted after valid row", data, io.ErrUnexpectedEOF, io.ErrUnexpectedEOF},
		{"cancelled after valid row", data, context.Canceled, context.Canceled},
	} {
		t.Run(test.name, func(t *testing.T) {
			db := sql.OpenDB(publicCatalogFault{data: test.data, terminal: test.terminal})
			t.Cleanup(func() { _ = db.Close() })
			got, err := (&Repository{db: db}).PublicCatalog(context.Background(), hash)
			if !errors.Is(err, test.want) || !reflect.DeepEqual(got, PublicCatalogBundle{}) {
				t.Fatalf("got=%+v err=%v; want %v", got, err, test.want)
			}
		})
	}
}

func TestPublicCatalogRejectsInvalidCompositionAndHash(t *testing.T) {
	for _, repository := range []*Repository{nil, {}} {
		if _, err := repository.PublicCatalog(context.Background(), "sha256:"+strings.Repeat("a", 64)); !errors.Is(err, ErrInvalidPublicCatalog) {
			t.Fatalf("invalid repository: %v", err)
		}
	}
	db := sql.OpenDB(publicCatalogFault{})
	t.Cleanup(func() { _ = db.Close() })
	for _, hash := range []string{"", "sha256:" + strings.Repeat("A", 64), "sha256:" + strings.Repeat("a", 63), strings.Repeat("a", 64)} {
		if _, err := (&Repository{db: db}).PublicCatalog(context.Background(), hash); !errors.Is(err, ErrUnknownPublicCatalog) {
			t.Fatalf("invalid hash %q: %v", hash, err)
		}
	}
}
