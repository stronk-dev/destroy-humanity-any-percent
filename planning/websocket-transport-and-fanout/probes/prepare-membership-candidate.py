#!/usr/bin/env python3
"""Prepare only the pinned RP-767 compiler overlays; never edit live sources."""
import difflib
import hashlib
import json
from pathlib import Path
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parents[3]
BASE = "215a98b3330445a1d7cc840d4c46a5cfad7c92ac"
SOURCE = "planning/websocket-transport-and-fanout/probes/membership-shutdown-integration_test.go.txt"
TARGET = "server/gameserver/membership_shutdown_observation_test.go"


def replace(source, before, after, count=1):
    if source.count(before) != count:
        raise RuntimeError(f"candidate anchor drift: {before!r}")
    return source.replace(before, after)


changes = {}


def original(path):
    data = subprocess.check_output(["git", "show", f"{BASE}:{path}"], cwd=ROOT)
    if (ROOT / path).read_bytes() != data:
        raise RuntimeError(f"source differs from pinned base: {path}")
    return data.decode()


path = "server/transport/authz.go"
value = original(path)
value = replace(value, 'import "strings"', 'import (\n"context"\n"strings"\n)')
for method, args in (("GuildMember", "accountID, guildID"), ("CohortMember", "founderID, cohortID"), ("MatchParticipant", "founderID, matchID")):
    value = replace(value, f"{method}({args} string) bool", f"{method}(ctx context.Context, {args} string) bool")
value = replace(value, "func Authorized(identity Identity,", "func Authorized(ctx context.Context, identity Identity,")
value = replace(value, 'if identity.AccountID == ""', 'if ctx.Err() != nil || identity.AccountID == ""')
for method in ("GuildMember", "CohortMember", "MatchParticipant"):
    value = replace(value, f"memberships.{method}(identity.", f"memberships.{method}(ctx, identity.")
changes[path] = value

path = "server/transport/node.go"
value = original(path)
value = replace(value, '''		client.OnSubscribe(func(event centrifuge.SubscribeEvent, callback centrifuge.SubscribeCallback) {
			identity, ok := identityFromClient(client)
			if !ok || !Authorized(identity, event.Channel, n.memberships) {''', '''		client.OnSubscribe(func(event centrifuge.SubscribeEvent, callback centrifuge.SubscribeCallback) {
			if n.authCtx.Err() != nil {
				callback(centrifuge.SubscribeReply{}, disconnectServerDrain)
				return
			}
			lookupCtx, cancelLookup := context.WithCancel(client.Context())
			stopNodeCancellation := context.AfterFunc(n.authCtx, cancelLookup)
			defer stopNodeCancellation()
			defer cancelLookup()
			identity, ok := identityFromClient(client)
			allowed := ok && Authorized(lookupCtx, identity, event.Channel, n.memberships)
			if n.authCtx.Err() != nil {
				callback(centrifuge.SubscribeReply{}, disconnectServerDrain)
				return
			}
			if !allowed {''')
changes[path] = value

path = "server/gameserver/composition.go"
value = original(path)
for method, args in (("GuildMember", "accountID, guildID"), ("CohortMember", "founderID, cohortID")):
    value = replace(value, f"{method}({args} string) bool", f"{method}(ctx context.Context, {args} string) bool")
    value = replace(value, f".{method}({args})", f".{method}(ctx, {args})")
value = replace(value, "MatchParticipant(string, string) bool", "MatchParticipant(context.Context, string, string) bool")
changes[path] = value

path = "server/guild/intents.go"
value = original(path)
value = replace(value, "GuildMember(accountID, guildID string) bool", "GuildMember(ctx context.Context, accountID, guildID string) bool")
value = replace(value, "service.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM guild_members", "service.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM guild_members")
value = replace(value, "CohortMember(_, _ string) bool", "CohortMember(context.Context, string, string) bool")
value = replace(value, "MatchParticipant(_, _ string) bool", "MatchParticipant(context.Context, string, string) bool")
changes[path] = value

path = "server/commonsprojection/projector.go"
value = original(path)
value = replace(value, "CohortMember(founderID, cohortID string) bool", "CohortMember(ctx context.Context, founderID, cohortID string) bool")
value = replace(value, "p.db.QueryRow(`SELECT EXISTS(", "p.db.QueryRowContext(ctx, `SELECT EXISTS(")
changes[path] = value

path = "server/transport/buffer_test.go"
value = original(path)
value = replace(value, "import (", 'import (\n"context"')
for method in ("GuildMember", "CohortMember", "MatchParticipant"):
    value = replace(value, f"{method}(_, id string) bool", f"{method}(_ context.Context, _, id string) bool")
value = replace(value, "Authorized(identity,", "Authorized(context.Background(), identity,", count=4)
value = replace(value, '''		t.Fatal("channel authorization mismatch")
	}
}''', '''		t.Fatal("channel authorization mismatch")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for _, channel := range []string{"world", "feed", "player:founder", "guild:guild.allowed", "cohort:cohort.allowed", "match:match.allowed"} {
		if !Authorized(context.Background(), identity, channel, testMemberships{}) || Authorized(ctx, identity, channel, testMemberships{}) {
			t.Fatalf("membership cancellation changed channel permission: %s", channel)
		}
	}
}''')
changes[path] = value

path = "server/transport/node_test.go"
value = original(path)
for method, args in (("GuildMember", "accountID, id"), ("CohortMember", "founderID, id"), ("MatchParticipant", "founderID, id")):
    value = replace(value, f"{method}({args} string) bool", f"{method}(_ context.Context, {args} string) bool")
changes[path] = value

path = "server/gameserver/composition_integration_test.go"
changes[path] = replace(original(path), ".Commons.CohortMember(founderID, cohortID)", ".Commons.CohortMember(ctx, founderID, cohortID)", count=3)

path = "server/guild/intents_integration_test.go"
changes[path] = replace(original(path), "service.GuildMember(", "service.GuildMember(ctx, ", count=3)

guard = json.loads((ROOT / "kernel/affecting-paths.json").read_text())
guarded = sorted(path for path in changes if not path.endswith("_test.go") and any(path.startswith(entry) if entry.endswith("/") else path == entry for entry in guard["paths"]))
if guarded != ["server/guild/intents.go"]:
    raise RuntimeError(f"kernel scope drift: {guarded}")

parent = ROOT / (".cache/transport-membership-shutdown-" + BASE[:8])
parent.mkdir(parents=True, exist_ok=True)
output = Path(tempfile.mkdtemp(prefix="candidate-", dir=parent))
patch = []
bindings = {}
for path, value in changes.items():
    value = subprocess.run(["gofmt"], input=value.encode(), stdout=subprocess.PIPE, check=True, cwd=ROOT).stdout
    destination = output / path
    destination.parent.mkdir(parents=True, exist_ok=True)
    destination.write_bytes(value)
    before = (ROOT / path).read_bytes()
    bindings[path] = {"base_sha256": hashlib.sha256(before).hexdigest(), "candidate_sha256": hashlib.sha256(value).hexdigest()}
    patch.extend(difflib.unified_diff(before.decode().splitlines(keepends=True), value.decode().splitlines(keepends=True), fromfile="a/"+path, tofile="b/"+path))
(output / "candidate.patch").write_text("".join(patch))

for population in ("baseline", "candidate"):
    for environment in ("native", "docker"):
        base = str(ROOT) if environment == "native" else "/workspace"
        mapping = {base+"/"+TARGET: base+"/"+SOURCE}
        if population == "candidate":
            mapping.update({base+"/"+path: base+"/"+str((output/path).relative_to(ROOT)) for path in changes})
        (output / f"{population}-{environment}.json").write_text(json.dumps({"Replace": mapping}, indent=2)+"\n")
(output / "bindings.json").write_text(json.dumps({"base_commit": BASE, "files": bindings, "guarded_paths": guarded, "probe_sha256": hashlib.sha256((ROOT/SOURCE).read_bytes()).hexdigest()}, indent=2)+"\n")
print(json.dumps({"output": str(output.relative_to(ROOT)), "guarded_paths": guarded, "active_sources_changed": False}))
