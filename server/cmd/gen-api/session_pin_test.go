package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cloud-clicker/server/account"
	"cloud-clicker/server/publicapi"
	"cloud-clicker/server/publicread"
)

func TestCommittedAPIPinProtectsAccountAndSessionOperations(t *testing.T) {
	private, err := account.PrivateAPIRegistry()
	if err != nil {
		t.Fatal(err)
	}
	public, err := publicread.Registry()
	if err != nil {
		t.Fatal(err)
	}
	registry, err := publicapi.MergeRegistries(private, public)
	if err != nil {
		t.Fatal(err)
	}
	pin, err := os.ReadFile(filepath.Join("..", "..", "..", "docs", "generated", "api-compat-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := publicapi.CheckCompatibilityPin(pin, registry); err != nil {
		t.Fatal(err)
	}
	for _, row := range []struct {
		id     string
		status int
	}{
		{"create_session", 401}, {"refresh_session", 401},
		{"create_account", 500}, {"create_founder", 401}, {"get_founder", 401},
	} {
		id := row.id
		for _, fault := range []string{"operation removal", "error status removal"} {
			t.Run(id+"/"+fault, func(t *testing.T) {
				operations := registry.Operations()
				for index, operation := range operations {
					if operation.ID != id {
						continue
					}
					if fault == "operation removal" {
						operations = append(operations[:index], operations[index+1:]...)
					} else {
						for responseIndex, response := range operation.Responses {
							if response.Status == row.status {
								operations[index].Responses = append(operation.Responses[:responseIndex], operation.Responses[responseIndex+1:]...)
								break
							}
						}
					}
					break
				}
				broken, err := publicapi.NewRegistry(registry.Schemas(), operations)
				if err != nil {
					t.Fatal(err)
				}
				if err := publicapi.CheckCompatibilityPin(pin, broken); err == nil || !strings.Contains(err.Error(), id) {
					t.Fatalf("committed pin did not reject %s at %s: %v", fault, id, err)
				}
			})
		}
	}
}
