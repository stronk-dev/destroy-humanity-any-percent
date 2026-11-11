package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

const fixtureURL = "postgres://cloud_clicker:cloud_clicker_test@postgres:5432/cloud_clicker_test?sslmode=disable"

type fakeAdmin struct {
	queries                      []string
	createErr, dropErr, closeErr error
	closed                       bool
	cleanupCanceled              bool
}

func (a *fakeAdmin) Exec(ctx context.Context, sql string, _ ...any) (pgconn.CommandTag, error) {
	a.queries = append(a.queries, sql)
	if strings.HasPrefix(sql, "CREATE DATABASE ") {
		return pgconn.CommandTag{}, a.createErr
	}
	a.cleanupCanceled = ctx.Err() != nil
	return pgconn.CommandTag{}, a.dropErr
}

func (a *fakeAdmin) Close(ctx context.Context) error {
	a.closed = true
	a.cleanupCanceled = a.cleanupCanceled || ctx.Err() != nil
	return a.closeErr
}

func fakeExecutor(a *fakeAdmin) (executor, *bytes.Buffer, *bytes.Buffer) {
	out, stderr := new(bytes.Buffer), new(bytes.Buffer)
	e := newExecutor(strings.NewReader("test input"), out, stderr)
	e.random = bytes.NewReader(make([]byte, 16))
	e.connect = func(context.Context, string) (databaseAdmin, error) { return a, nil }
	return e, out, stderr
}

func helperArgs(t *testing.T, mode string, args ...string) []string {
	t.Helper()
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	return append([]string{self, "-test.run=^TestExecHelper$", "--", mode}, args...)
}

func helperEnvironment(raw string) []string {
	return append(isolatedEnvironment(os.Environ(), raw, "replaced stale base"), "CC_EXEC_TEST_HELPER=1")
}

func TestRefusesUnsafeInvocationBeforeConnecting(t *testing.T) {
	cases := []struct {
		name      string
		args, env []string
	}{
		{"no executable", nil, []string{"TEST_DATABASE_URL=" + fixtureURL}},
		{"relative executable", []string{"test"}, []string{"TEST_DATABASE_URL=" + fixtureURL}},
		{"missing URL", []string{"/missing"}, nil},
		{"duplicate URL", []string{"/missing"}, []string{"TEST_DATABASE_URL=" + fixtureURL, "TEST_DATABASE_URL=" + fixtureURL}},
		{"manual restore profile", []string{"/missing"}, []string{"TEST_DATABASE_URL=" + fixtureURL, "TEST_ADMIN_DATABASE_URL=" + fixtureURL}},
		{"manual restore destination", []string{"/missing"}, []string{"TEST_DATABASE_URL=" + fixtureURL, "TEST_RESTORE_DATABASE_URL=" + fixtureURL}},
	}
	for _, raw := range []string{"", "postgres://cloud_clicker:secret@remote.example/cloud_clicker_test?sslmode=disable", strings.Replace(fixtureURL, "cloud_clicker_test?", "live?", 1), strings.Replace(fixtureURL, "cloud_clicker:", "other:", 1), fixtureURL + "&options=unsafe", fixtureURL + "#fragment", strings.Replace(fixtureURL, "postgres://", "postgresql://", 1), strings.Replace(fixtureURL, "/cloud_clicker_test", "/%63loud_clicker_test", 1)} {
		cases = append(cases, struct {
			name      string
			args, env []string
		}{raw, []string{"/missing"}, []string{"TEST_DATABASE_URL=" + raw}})
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e, _, stderr := fakeExecutor(&fakeAdmin{})
			e.connect = func(context.Context, string) (databaseAdmin, error) {
				t.Fatal("connected on invalid invocation")
				return nil, nil
			}
			if e.run(context.Background(), tc.args, tc.env) != 1 {
				t.Fatal("invalid invocation accepted")
			}
			if strings.Contains(stderr.String(), "secret") {
				t.Fatal("credentials disclosed")
			}
		})
	}
}

func TestForwardsExecutableArgumentsStreamsAndEnvironment(t *testing.T) {
	a := &fakeAdmin{}
	e, out, stderr := fakeExecutor(a)
	args := []string{"literal space", "$(never execute)", "--test.flag=value"}
	env := append(helperEnvironment(fixtureURL), "CC_EXEC_TEST_KEEP=preserved")
	if code := e.run(context.Background(), helperArgs(t, "echo", args...), env); code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	var got struct {
		Args                   []string
		Input, URL, Base, Keep string
	}
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err, out.String())
	}
	wantURL := strings.Replace(fixtureURL, "/cloud_clicker_test?", "/cc_pkg_test_"+strings.Repeat("0", 32)+"?", 1)
	if !reflect.DeepEqual(got.Args, args) || got.Input != "test input" || got.URL != wantURL || got.Base != fixtureURL || got.Keep != "preserved" {
		t.Fatalf("forwarding mismatch: %+v", got)
	}
	if stderr.String() != "child stderr\n" {
		t.Fatal(stderr.String())
	}
	if len(a.queries) != 2 || a.queries[0] != "CREATE DATABASE \"cc_pkg_test_"+strings.Repeat("0", 32)+"\"" || a.queries[1] != "DROP DATABASE \"cc_pkg_test_"+strings.Repeat("0", 32)+"\" WITH (FORCE)" || !a.closed || a.cleanupCanceled {
		t.Fatalf("lifecycle: %+v", a)
	}
}

func TestSetupAndCleanupFailuresStayRed(t *testing.T) {
	for _, tc := range []struct {
		name                 string
		create, drop, close  bool
		child, want, queries int
	}{
		{"create denied", true, false, false, 0, 1, 1},
		{"child fails", false, false, false, 7, 7, 2},
		{"drop fails", false, true, false, 0, 1, 2},
		{"close fails", false, false, true, 0, 1, 2},
		{"child and cleanup fail", false, true, true, 7, 7, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := &fakeAdmin{}
			secret := errors.New("password=do-not-disclose")
			if tc.create {
				a.createErr = secret
			}
			if tc.drop {
				a.dropErr = secret
			}
			if tc.close {
				a.closeErr = secret
			}
			e, _, stderr := fakeExecutor(a)
			if code := e.run(context.Background(), helperArgs(t, "exit", strconv.Itoa(tc.child)), helperEnvironment(fixtureURL)); code != tc.want {
				t.Fatalf("code=%d want=%d", code, tc.want)
			}
			if len(a.queries) != tc.queries || !a.closed || strings.Contains(stderr.String(), "do-not-disclose") {
				t.Fatalf("lifecycle=%+v stderr=%s", a, stderr)
			}
		})
	}
	for _, stage := range []string{"random", "connect", "start"} {
		t.Run(stage, func(t *testing.T) {
			a := &fakeAdmin{}
			e, _, stderr := fakeExecutor(a)
			args := helperArgs(t, "exit", "0")
			switch stage {
			case "random":
				e.random = strings.NewReader("")
			case "connect":
				e.connect = func(context.Context, string) (databaseAdmin, error) {
					return nil, errors.New("password=do-not-disclose")
				}
			case "start":
				args = []string{"/nonexistent/cloud-clicker-test"}
			}
			if code := e.run(context.Background(), args, helperEnvironment(fixtureURL)); code != 1 {
				t.Fatal(code)
			}
			if strings.Contains(stderr.String(), "do-not-disclose") {
				t.Fatal(stderr.String())
			}
			if stage == "start" && (len(a.queries) != 2 || !a.closed) {
				t.Fatalf("owned DB leaked: %+v", a)
			}
			if stage != "start" && len(a.queries) != 0 {
				t.Fatal("setup continued")
			}
		})
	}
}

func TestCancellationKillsChildGroupAndCleansDatabase(t *testing.T) {
	a := &fakeAdmin{}
	e, _, stderr := fakeExecutor(a)
	ready := filepath.Join(t.TempDir(), "ready")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan int, 1)
	args := helperArgs(t, "group", ready)
	go func() { done <- e.run(ctx, args, helperEnvironment(fixtureURL)) }()
	data := awaitFile(t, ready)
	var pids []int
	if err := json.Unmarshal(data, &pids); err != nil {
		t.Fatal(err)
	}
	cancel()
	select {
	case code := <-done:
		if code != 128+int(syscall.SIGKILL) {
			t.Fatalf("code=%d stderr=%s", code, stderr)
		}
	case <-time.After(20 * time.Second):
		t.Fatal("cancellation failed")
	}
	for _, pid := range pids {
		awaitProcessExit(t, pid)
	}
	if len(a.queries) != 2 || !a.closed || a.cleanupCanceled {
		t.Fatalf("cancellation skipped independent cleanup: %+v", a)
	}
}

func awaitFile(t *testing.T, path string) []byte {
	t.Helper()
	until := time.Now().Add(20 * time.Second)
	for time.Now().Before(until) {
		data, err := os.ReadFile(path)
		if err == nil {
			return data
		}
		if !os.IsNotExist(err) {
			t.Fatal(err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("child readiness missing: %s", path)
	return nil
}

func awaitProcessExit(t *testing.T, pid int) {
	t.Helper()
	until := time.Now().Add(5 * time.Second)
	for time.Now().Before(until) {
		if errors.Is(syscall.Kill(pid, 0), syscall.ESRCH) {
			return
		}
		// A killed grandchild can briefly await init reaping in Linux containers.
		if b, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid)); err == nil {
			if i := strings.LastIndex(string(b), ") "); i >= 0 && len(b) > i+2 && b[i+2] == 'Z' {
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("child process %d still running", pid)
}

func TestExecHelper(t *testing.T) {
	if os.Getenv("CC_EXEC_TEST_HELPER") != "1" {
		return
	}
	i := 0
	for i < len(os.Args) && os.Args[i] != "--" {
		i++
	}
	args := os.Args[i+1:]
	switch args[0] {
	case "echo":
		input := new(bytes.Buffer)
		_, _ = input.ReadFrom(os.Stdin)
		_ = json.NewEncoder(os.Stdout).Encode(struct {
			Args                   []string
			Input, URL, Base, Keep string
		}{args[1:], input.String(), os.Getenv("TEST_DATABASE_URL"), os.Getenv(baseURLVariable), os.Getenv("CC_EXEC_TEST_KEEP")})
		fmt.Fprintln(os.Stderr, "child stderr")
		os.Exit(0)
	case "exit":
		code, _ := strconv.Atoi(args[1])
		os.Exit(code)
	case "sleep":
		for {
			time.Sleep(time.Hour)
		}
	case "group":
		child := exec.Command(os.Args[0], "-test.run=^TestExecHelper$", "--", "sleep")
		if err := child.Start(); err != nil {
			panic(err)
		}
		b, _ := json.Marshal([]int{os.Getpid(), child.Process.Pid})
		if err := os.WriteFile(args[1]+".tmp", b, 0600); err != nil {
			panic(err)
		}
		if err := os.Rename(args[1]+".tmp", args[1]); err != nil {
			panic(err)
		}
		for {
			time.Sleep(time.Hour)
		}
	case "wrapper":
		os.Args = append([]string{os.Args[0]}, args[1:]...)
		main()
	case "database":
		databaseHelper(args[1:])
	default:
		panic("unknown test helper")
	}
}
