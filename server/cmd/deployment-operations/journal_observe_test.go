package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/operations"
)

func TestJournalObserveUsesRequestedDirectoryAndRefusesIncompleteCollection(t *testing.T) {
	for _, mode := range []string{"valid", "warning", "access-warning", "failed", "truncated", "empty"} {
		t.Run(mode, func(t *testing.T) {
			directory := t.TempDir()
			argumentsPath := filepath.Join(directory, "arguments.json")
			binary, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			// Exercise the actual CLI subprocess path, including stderr and exit
			// status, without claiming this fixture is a Linux journal.
			script := "#!/bin/sh\nexec '" + strings.ReplaceAll(binary, "'", "'\\''") + "' -test.run='^TestJournalObserveCommandProcess$' -- \"$@\"\n"
			if err := os.WriteFile(filepath.Join(directory, "journalctl"), []byte(script), 0o700); err != nil {
				t.Fatal(err)
			}
			t.Setenv("PATH", directory+string(os.PathListSeparator)+os.Getenv("PATH"))
			t.Setenv("CLOUD_CLICKER_JOURNAL_FIXTURE_MODE", mode)
			t.Setenv("CLOUD_CLICKER_JOURNAL_FIXTURE_ARGUMENTS", argumentsPath)
			output, err := os.CreateTemp(directory, "observation-")
			if err != nil {
				t.Fatal(err)
			}
			defer output.Close()
			previous := os.Stdout
			os.Stdout = output
			defer func() { os.Stdout = previous }()
			started := time.Now().UTC().Add(-time.Hour).Truncate(time.Second)
			observeErr := runJournalObserve([]string{"--population=journal_collection_fixture", "--started-at=" + started.Format(time.RFC3339),
				"--journal-max-use-bytes=1048576", "--journal-path=" + directory})
			os.Stdout = previous
			data, err := os.ReadFile(output.Name())
			if err != nil {
				t.Fatal(err)
			}
			if mode != "valid" {
				if observeErr == nil || len(data) != 0 {
					t.Fatalf("incomplete collection produced evidence: mode=%s error=%v bytes=%d", mode, observeErr, len(data))
				}
				if strings.Contains(observeErr.Error(), "private") {
					t.Fatal("private journal diagnostic escaped the bounded refusal")
				}
				return
			}
			observation, err := operations.DecodeJournalObservation(data)
			if observeErr != nil || err != nil || observation.Samples != 2 || !observation.StartedAt.Equal(started) {
				t.Fatalf("complete collection rejected: error=%v decode=%v", observeErr, err)
			}
			argumentBytes, err := os.ReadFile(argumentsPath)
			if err != nil {
				t.Fatal(err)
			}
			var arguments []string
			if err := json.Unmarshal(argumentBytes, &arguments); err != nil {
				t.Fatal(err)
			}
			for _, required := range []string{"--directory=" + directory, "--since=@" + fmt.Sprint(started.Unix()), "--output=json", "--all", "--no-pager"} {
				if !containsJournalArgument(arguments, required) {
					t.Errorf("journal collection omitted exact argument %q", required)
				}
			}
			for _, service := range []string{"caddy", "gameserver", "postgres", "backup", "prometheus", "alertmanager", "node-exporter"} {
				if !containsJournalArgument(arguments, "SYSLOG_IDENTIFIER=cloud-clicker-"+service) {
					t.Errorf("journal collection omitted service %s", service)
				}
			}
		})
	}
}

func containsJournalArgument(arguments []string, want string) bool {
	for _, argument := range arguments {
		if argument == want {
			return true
		}
	}
	return false
}

func TestJournalObserveCommandProcess(t *testing.T) {
	mode := os.Getenv("CLOUD_CLICKER_JOURNAL_FIXTURE_MODE")
	if mode == "" {
		return
	}
	arguments := []string{}
	for index, argument := range os.Args {
		if argument == "--" {
			arguments = os.Args[index+1:]
			break
		}
	}
	data, err := json.Marshal(arguments)
	if err != nil || os.WriteFile(os.Getenv("CLOUD_CLICKER_JOURNAL_FIXTURE_ARGUMENTS"), data, 0o600) != nil {
		os.Exit(2)
	}
	if mode != "empty" {
		fmt.Println(`{"SYSLOG_IDENTIFIER":"cloud-clicker-gameserver","MESSAGE":"fixture first"}`)
		fmt.Println(`{"SYSLOG_IDENTIFIER":"cloud-clicker-caddy","MESSAGE":"fixture second"}`)
	}
	// journalctl suppresses inaccessible-journal warnings with --quiet even
	// when it can still return records from other readable journal files.
	if mode == "warning" || mode == "access-warning" && !containsJournalArgument(arguments, "--quiet") {
		fmt.Fprintln(os.Stderr, "fixture incomplete collection: recovery_code=private")
	}
	if mode == "failed" {
		os.Exit(1)
	}
	if mode == "truncated" {
		fmt.Print(`{"MESSAGE":`)
	}
	os.Exit(0)
}
