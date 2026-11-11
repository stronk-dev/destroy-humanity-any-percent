package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const baseURLVariable = "CLOUD_CLICKER_TEST_DATABASE_BASE_URL"

// Contain failed fixture setup/cleanup. This is not a test, pacing or CI latency
// allowance; the ordinary Go test timeout and outer job guard remain unchanged.
const databaseOperationGuard = 15 * time.Second

type databaseAdmin interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Close(context.Context) error
}

type executor struct {
	connect func(context.Context, string) (databaseAdmin, error)
	random  io.Reader
	in      io.Reader
	out     io.Writer
	err     io.Writer
}

func newExecutor(in io.Reader, out, stderr io.Writer) executor {
	return executor{connect: func(ctx context.Context, raw string) (databaseAdmin, error) {
		return pgx.Connect(ctx, raw)
	}, random: rand.Reader, in: in, out: out, err: stderr}
}

func (e executor) run(ctx context.Context, args, environment []string) (code int) {
	refuse := func(stage string) int {
		// Connection errors can contain credentials. Emit only the stage.
		fmt.Fprintln(e.err, "test database isolation failed:", stage)
		return 1
	}
	if len(args) == 0 || !filepath.IsAbs(args[0]) {
		return refuse("test executable")
	}
	for _, entry := range environment {
		for _, name := range []string{"TEST_ADMIN_DATABASE_URL", "TEST_RESTORE_DATABASE_URL"} {
			if strings.HasPrefix(entry, name+"=") && entry != name+"=" {
				return refuse("multi-database profiles require separate isolation")
			}
		}
	}
	raw, ok := singleEnvironmentValue(environment, "TEST_DATABASE_URL")
	base, err := repositoryTestURL(raw)
	if !ok || err != nil {
		return refuse("repository test URL")
	}
	var nonce [16]byte
	if _, err := io.ReadFull(e.random, nonce[:]); err != nil {
		return refuse("database identity")
	}
	name := "cc_pkg_test_" + hex.EncodeToString(nonce[:])
	setup, cancelSetup := context.WithTimeout(ctx, databaseOperationGuard)
	admin, err := e.connect(setup, raw)
	if err != nil {
		cancelSetup()
		return refuse("database connection")
	}
	created := false
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), databaseOperationGuard)
		defer cancel()
		if created {
			_, err := admin.Exec(cleanup, "DROP DATABASE "+pgx.Identifier{name}.Sanitize()+" WITH (FORCE)")
			if err != nil {
				refuse("owned database cleanup")
				if code == 0 {
					code = 1
				}
			}
		}
		if err := admin.Close(cleanup); err != nil {
			refuse("database connection cleanup")
			if code == 0 {
				code = 1
			}
		}
	}()
	_, err = admin.Exec(setup, "CREATE DATABASE "+pgx.Identifier{name}.Sanitize())
	cancelSetup()
	if err != nil {
		// A failed/ambiguous CREATE never grants authority to drop an existing DB.
		return refuse("database creation; cleanup authority not obtained")
	}
	created = true
	base.Path = "/" + name
	command := exec.CommandContext(ctx, args[0], args[1:]...)
	command.Env = isolatedEnvironment(environment, base.String(), raw)
	command.Stdin, command.Stdout, command.Stderr = e.in, e.out, e.err
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	command.WaitDelay = databaseOperationGuard
	if err := command.Start(); err != nil {
		return refuse("test execution")
	}
	// Even a successful test must not leave subprocesses using its disposable DB.
	defer syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
	if err := command.Wait(); err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			if status, ok := exit.Sys().(syscall.WaitStatus); ok && status.Signaled() {
				return 128 + int(status.Signal())
			}
			return exit.ExitCode()
		}
		return refuse("test execution")
	}
	return 0
}

func repositoryTestURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "postgres" || u.User == nil || u.User.Username() != "cloud_clicker" ||
		u.EscapedPath() != "/cloud_clicker_test" || u.RawQuery != "sslmode=disable" || u.Fragment != "" || u.Opaque != "" {
		return nil, errors.New("not the repository test database")
	}
	switch u.Hostname() {
	case "postgres", "localhost", "127.0.0.1", "::1":
		return u, nil
	default:
		return nil, errors.New("not the repository test host")
	}
}

func singleEnvironmentValue(environment []string, name string) (string, bool) {
	var value string
	found := false
	for _, entry := range environment {
		if strings.HasPrefix(entry, name+"=") {
			if found {
				return "", false
			}
			value, found = strings.TrimPrefix(entry, name+"="), true
		}
	}
	return value, found && value != ""
}

func isolatedEnvironment(environment []string, isolatedURL, baseURL string) []string {
	result := make([]string, 0, len(environment)+2)
	for _, entry := range environment {
		if !strings.HasPrefix(entry, "TEST_DATABASE_URL=") && !strings.HasPrefix(entry, baseURLVariable+"=") {
			result = append(result, entry)
		}
	}
	return append(result, "TEST_DATABASE_URL="+isolatedURL, baseURLVariable+"="+baseURL)
}
