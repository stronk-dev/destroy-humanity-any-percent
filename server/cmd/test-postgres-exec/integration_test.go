package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

type databaseReady struct {
	Database string
	PIDs     []int
}

func integrationDatabase(t *testing.T) (string, *pgx.Conn) {
	t.Helper()
	raw := os.Getenv(baseURLVariable)
	if raw == "" {
		raw = os.Getenv("TEST_DATABASE_URL")
	}
	if raw == "" {
		t.Skip("requires the declared Postgres test service")
	}
	if _, err := repositoryTestURL(raw); err != nil {
		t.Fatal("invalid integration fixture URL")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	c, err := pgx.Connect(ctx, raw)
	if err != nil {
		t.Fatal("declared Postgres unavailable")
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := c.Close(ctx); err != nil {
			t.Error("closing integration fixture")
		}
	})
	return raw, c
}

type runningHelper struct {
	command *exec.Cmd
	stderr  *bytes.Buffer
	done    chan error
}

func startDatabaseHelper(t *testing.T, raw, ready, release, table, value string, exit int, wrapped bool) *runningHelper {
	t.Helper()
	args := helperArgs(t, "database", ready, release, table, value, strconv.Itoa(exit), strconv.FormatBool(wrapped))
	if wrapped {
		args = helperArgs(t, "wrapper", args...)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	command := exec.CommandContext(ctx, args[0], args[1:]...)
	command.Env = helperEnvironment(raw)
	stderr := new(bytes.Buffer)
	command.Stderr = stderr
	if err := command.Start(); err != nil {
		cancel()
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- command.Wait() }()
	h := &runningHelper{command, stderr, done}
	t.Cleanup(func() {
		cancel()
		select {
		case <-done:
		case <-time.After(20 * time.Second):
			t.Error("test helper did not terminate")
		}
	})
	return h
}

func (h *runningHelper) wait(t *testing.T, want int) {
	t.Helper()
	select {
	case err := <-h.done:
		// Retain the terminal observation for cleanup without another Wait call.
		h.done <- err
		code := 0
		if err != nil {
			e, ok := err.(*exec.ExitError)
			if !ok {
				t.Fatal(err)
			}
			code = e.ExitCode()
		}
		if code != want {
			t.Fatalf("helper code=%d want=%d stderr=%s", code, want, h.stderr)
		}
	case <-time.After(25 * time.Second):
		t.Fatal("helper failed to finish")
	}
}

func readyDatabase(t *testing.T, path string) databaseReady {
	t.Helper()
	var ready databaseReady
	if err := json.Unmarshal(awaitFile(t, path), &ready); err != nil {
		t.Fatal(err)
	}
	return ready
}

func fixtureExec(t *testing.T, c *pgx.Conn, sql string, args ...any) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if _, err := c.Exec(ctx, sql, args...); err != nil {
		t.Fatal(err)
	}
}

func requireDatabaseAbsent(t *testing.T, c *pgx.Conn, name string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	var exists bool
	if err := c.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname=$1)", name).Scan(&exists); err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatalf("owned database remains: %s", name)
	}
}

func TestConcurrentDatabaseIsolationIntegration(t *testing.T) {
	raw, c := integrationDatabase(t)
	dir := t.TempDir()
	release := filepath.Join(dir, "release")
	table := fmt.Sprintf("ci_exec_%d", time.Now().UnixNano())
	quoted := pgx.Identifier{table}.Sanitize()
	fixtureExec(t, c, "CREATE TABLE "+quoted+" (value text NOT NULL)")
	t.Cleanup(func() { fixtureExec(t, c, "DROP TABLE "+quoted) })
	fixtureExec(t, c, "INSERT INTO "+quoted+" VALUES ('original fixture remains')")
	// Both children create the same relation while overlapping. Sharing a DB
	// would produce a duplicate-table error or expose the other child's row.
	a := startDatabaseHelper(t, raw, filepath.Join(dir, "a"), release, table, "alpha", 0, true)
	b := startDatabaseHelper(t, raw, filepath.Join(dir, "b"), release, table, "beta", 0, true)
	ra, rb := readyDatabase(t, filepath.Join(dir, "a")), readyDatabase(t, filepath.Join(dir, "b"))
	if ra.Database == rb.Database || ra.Database == "cloud_clicker_test" || rb.Database == "cloud_clicker_test" {
		t.Fatalf("not isolated: %+v %+v", ra, rb)
	}
	if err := os.WriteFile(release, nil, 0600); err != nil {
		t.Fatal(err)
	}
	a.wait(t, 0)
	b.wait(t, 0)
	for _, pid := range append(ra.PIDs, rb.PIDs...) {
		awaitProcessExit(t, pid)
	}
	requireDatabaseAbsent(t, c, ra.Database)
	requireDatabaseAbsent(t, c, rb.Database)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	var count int
	var value string
	if err := c.QueryRow(ctx, "SELECT count(*), min(value) FROM "+quoted).Scan(&count, &value); err != nil {
		t.Fatal(err)
	}
	if count != 1 || value != "original fixture remains" {
		t.Fatalf("source fixture altered: %d %s", count, value)
	}
}

func TestSharedDatabaseNegativeControlIntegration(t *testing.T) {
	raw, c := integrationDatabase(t)
	dir := t.TempDir()
	release := filepath.Join(dir, "release")
	table := fmt.Sprintf("ci_exec_shared_%d", time.Now().UnixNano())
	t.Cleanup(func() { fixtureExec(t, c, "DROP TABLE IF EXISTS "+pgx.Identifier{table}.Sanitize()) })
	a := startDatabaseHelper(t, raw, filepath.Join(dir, "a"), release, table, "alpha", 0, false)
	readyDatabase(t, filepath.Join(dir, "a"))
	b := startDatabaseHelper(t, raw, filepath.Join(dir, "b"), release, table, "beta", 0, false)
	b.wait(t, 23)
	if b.stderr.String() != "database helper: create relation\n" {
		t.Fatalf("wrong shared-DB failure: %s", b.stderr)
	}
	if err := os.WriteFile(release, nil, 0600); err != nil {
		t.Fatal(err)
	}
	a.wait(t, 0)
}

func TestFailureAndSignalDatabaseCleanupIntegration(t *testing.T) {
	raw, c := integrationDatabase(t)
	for _, cancelChild := range []bool{false, true} {
		t.Run(strconv.FormatBool(cancelChild), func(t *testing.T) {
			dir := t.TempDir()
			release := filepath.Join(dir, "release")
			ready := filepath.Join(dir, "ready")
			h := startDatabaseHelper(t, raw, ready, release, "owned_relation", "owned", 7, true)
			r := readyDatabase(t, ready)
			if cancelChild {
				if err := h.command.Process.Signal(syscall.SIGTERM); err != nil {
					t.Fatal(err)
				}
				h.wait(t, 128+int(syscall.SIGKILL))
			} else {
				if err := os.WriteFile(release, nil, 0600); err != nil {
					t.Fatal(err)
				}
				h.wait(t, 7)
			}
			for _, pid := range r.PIDs {
				awaitProcessExit(t, pid)
			}
			requireDatabaseAbsent(t, c, r.Database)
		})
	}
}

func TestCreateCollisionDoesNotDropExistingDatabaseIntegration(t *testing.T) {
	raw, c := integrationDatabase(t)
	var nonce [16]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		t.Fatal(err)
	}
	name := "cc_pkg_test_" + hex.EncodeToString(nonce[:])
	quoted := pgx.Identifier{name}.Sanitize()
	fixtureExec(t, c, "CREATE DATABASE "+quoted)
	t.Cleanup(func() { fixtureExec(t, c, "DROP DATABASE "+quoted+" WITH (FORCE)") })
	e, _, stderr := fakeExecutor(nil)
	e.connect = func(ctx context.Context, raw string) (databaseAdmin, error) { return pgx.Connect(ctx, raw) }
	e.random = bytes.NewReader(nonce[:])
	if code := e.run(context.Background(), helperArgs(t, "exit", "0"), helperEnvironment(raw)); code != 1 {
		t.Fatal(code)
	}
	if stderr.String() != "test database isolation failed: database creation; cleanup authority not obtained\n" {
		t.Fatal(stderr.String())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	var exists bool
	if err := c.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname=$1)", name).Scan(&exists); err != nil || !exists {
		t.Fatal("unowned database deleted", err)
	}
}

// Runs only in a subprocess of TestExecHelper. SQL/stream failures remain real
// nonzero child results. No test runner, result rewriting or test retry is used.
func databaseHelper(args []string) {
	fail := func(stage string) { fmt.Fprintln(os.Stderr, "database helper:", stage); os.Exit(23) }
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	c, err := pgx.Connect(ctx, os.Getenv("TEST_DATABASE_URL"))
	if err != nil {
		fail("connect")
	}
	defer c.Close(context.Background())
	quoted := pgx.Identifier{args[2]}.Sanitize()
	if _, err := c.Exec(ctx, "CREATE TABLE "+quoted+" (value text NOT NULL)"); err != nil {
		fail("create relation")
	}
	if _, err := c.Exec(ctx, "INSERT INTO "+quoted+" VALUES ($1)", args[3]); err != nil {
		fail("insert")
	}
	var db string
	if err := c.QueryRow(ctx, "SELECT current_database()").Scan(&db); err != nil {
		fail("database identity")
	}
	pids := []int{os.Getpid()}
	if args[5] == "true" {
		child := exec.Command(os.Args[0], "-test.run=^TestExecHelper$", "--", "sleep")
		if err := child.Start(); err != nil {
			fail("descendant")
		}
		defer func() { _ = child.Process.Kill(); _ = child.Wait() }()
		pids = append(pids, child.Process.Pid)
	}
	data, _ := json.Marshal(databaseReady{db, pids})
	if err := os.WriteFile(args[0]+".tmp", data, 0600); err != nil {
		fail("readiness")
	}
	if err := os.Rename(args[0]+".tmp", args[0]); err != nil {
		fail("readiness commit")
	}
	for {
		if _, err := os.Stat(args[1]); err == nil {
			break
		}
		select {
		case <-ctx.Done():
			fail("release timeout")
		case <-time.After(10 * time.Millisecond):
		}
	}
	var count int
	var value string
	if err := c.QueryRow(ctx, "SELECT count(*),min(value) FROM "+quoted).Scan(&count, &value); err != nil || count != 1 || value != args[3] {
		fail("cross-package contamination")
	}
	code, _ := strconv.Atoi(args[4])
	// os.Exit deliberately bypasses the child's defers: the wrapper must clean
	// both its database and this otherwise surviving descendant on every exit.
	os.Exit(code)
}
