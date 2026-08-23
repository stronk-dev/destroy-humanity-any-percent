package main

import "testing"

func TestGoDependencyInventoryCoversEveryShippedBinary(t *testing.T) {
	dependencies, err := discoverGoDependencies("../../..")
	if err != nil {
		t.Fatal(err)
	}
	foundAge := false
	foundChromedp := false
	for _, dependency := range dependencies {
		if dependency.Name == "filippo.io/age" {
			foundAge = true
		}
		if dependency.Name == "github.com/chromedp/chromedp" {
			foundChromedp = true
		}
	}
	if !foundAge {
		t.Fatal("helper-only filippo.io/age dependency absent from application metadata")
	}
	if !foundChromedp {
		t.Fatal("browser-driver-only chromedp dependency absent from application metadata")
	}
}
