package main

import "testing"

func TestGoDependencyInventoryCoversEveryShippedBinary(t *testing.T) {
	dependencies, err := discoverGoDependencies("../../..")
	if err != nil {
		t.Fatal(err)
	}
	foundAge := false
	for _, dependency := range dependencies {
		if dependency.Name == "filippo.io/age" {
			foundAge = true
			break
		}
	}
	if !foundAge {
		t.Fatal("helper-only filippo.io/age dependency absent from application metadata")
	}
}
