package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"cloud-clicker/server/releasepackage"
)

func main() {
	root := flag.String("root", "..", "repository root")
	output := flag.String("output", "", "empty release bundle directory")
	serverBinary := flag.String("server-binary", "", "Linux/amd64 gameserver binary")
	backupBinary := flag.String("backup-binary", "", "Linux/amd64 deployment backup binary")
	releaseBinary := flag.String("release-binary", "", "Linux/amd64 deployment release binary")
	operationsBinary := flag.String("operations-binary", "", "Linux/amd64 deployment operations binary")
	rehearsalBinary := flag.String("rehearsal-binary", "", "Linux/amd64 deployment rehearsal binary")
	gameserverArchive := flag.String("gameserver-archive", "", "docker save archive for the gameserver image")
	clientDist := flag.String("client-dist", "", "built client directory")
	metadata := flag.String("metadata", "", "generated release metadata directory")
	version := flag.String("version", "", "release version")
	commit := flag.String("commit", "", "full source commit")
	dockerVersion := flag.String("docker-version", "", "tested Docker Engine version")
	composeVersion := flag.String("compose-version", "", "tested Docker Compose version")
	imageValues := map[string]*string{}
	configValues := map[string]*string{}
	sbomValues := map[string]*string{}
	for _, name := range []string{"alertmanager", "caddy", "gameserver", "node-exporter", "postgres", "prometheus"} {
		imageValues[name] = flag.String(name+"-image", "", "immutable "+name+" image reference")
		configValues[name] = flag.String(name+"-config-id", "", "linux/amd64 "+name+" runtime config digest")
		sbomValues[name] = flag.String(name+"-sbom", "", name+" image SPDX SBOM")
	}
	flag.Parse()
	images, configIDs, sboms := map[string]string{}, map[string]string{}, map[string]string{}
	for _, name := range []string{"alertmanager", "caddy", "gameserver", "node-exporter", "postgres", "prometheus"} {
		images[name], sboms[name] = strings.TrimSpace(*imageValues[name]), *sbomValues[name]
		configIDs[name] = strings.TrimSpace(*configValues[name])
	}
	manifest, err := releasepackage.AssembleBundle(releasepackage.BundleInput{RepositoryRoot: *root, Output: *output,
		ServerBinary: *serverBinary, BackupBinary: *backupBinary, ReleaseBinary: *releaseBinary, OperationsBinary: *operationsBinary, RehearsalBinary: *rehearsalBinary, GameserverImageArchive: *gameserverArchive, ClientDist: *clientDist, MetadataDirectory: *metadata,
		ReleaseVersion: *version, SourceCommit: *commit, DockerEngineVersion: *dockerVersion,
		DockerComposeVersion: *composeVersion, Images: images, ImageConfigIDs: configIDs, ImageSBOMs: sboms})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("release bundle %s: %d artifacts\n", manifest.ReleaseVersion, len(manifest.Artifacts))
}
