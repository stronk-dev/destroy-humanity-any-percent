# Deployment

Deployment Foundation is implementing. The repository does **not** yet claim a supported self-host
bundle or release-ready deployment; backup, rollback, operations and the exact-manifest clean-host
rehearsal remain unfinished.

## Runtime content closure

`server/releasepackage` derives the gameserver's repository-independent content from the live epoch
authority rather than maintaining another catalog list. The closure contains:

- `balance/epochs/phase0.json`, every artifact and epoch changelog it declares;
- `balance/transport/phase0.json` and `moderation/guild-names.txt`, the two additional files opened
  by gameserver composition; and
- `deployment/content-manifest.v1.json`, whose constants hash must equal the derived current bundle
  and whose copy hash must be canonical.

Every path is local, unique, nonempty and SHA-256 bound. Staging accepts only an empty destination,
copies exactly that sorted set with no test/source/planning trees, and then re-walks the output.
Missing, altered, extra or symlinked files fail validation. This means a later container can copy
the staged directory to `/opt/cloud-clicker/content` without depending on a repository checkout.

Run the staging boundary with:

```sh
make stage-release-content RELEASE_CONTENT_OUTPUT=/absolute/empty/directory
```

The command prints the epoch, constants/copy identities and every staged file hash as canonical
JSON. This is an input to the release manifest still being implemented; it is not itself release
evidence.

