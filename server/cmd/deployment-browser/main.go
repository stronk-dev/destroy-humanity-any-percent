package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"cloud-clicker/server/deploymentbrowser"
)

func main() {
	origin := flag.String("origin", "", "canonical HTTPS Caddy origin")
	output := flag.String("output", "", "exclusive sanitized browser result path")
	manifest := flag.String("manifest-sha256", "", "exact candidate release manifest hash")
	browser := flag.String("browser", deploymentbrowser.DefaultBrowserPath, "absolute pinned Chromium path")
	allowLoopback := flag.Bool("allow-loopback-http", false, "test-only loopback HTTP origin")
	timeout := flag.Duration("timeout", 2*time.Hour, "bounded browser workflow timeout")
	flag.Parse()
	if flag.NArg() != 0 || *timeout < time.Second || *timeout > 2*time.Hour {
		fail()
	}
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	if _, err := deploymentbrowser.Run(ctx, deploymentbrowser.Config{Origin: *origin, Output: *output,
		ManifestSHA256: *manifest, BrowserPath: *browser, AllowLoopbackHTTP: *allowLoopback, Now: time.Now}); err != nil {
		fail()
	}
	fmt.Println("deployment browser workflow passed")
}

func fail() {
	fmt.Fprintln(os.Stderr, "deployment browser failed: error_class=workflow_failed")
	os.Exit(1)
}
