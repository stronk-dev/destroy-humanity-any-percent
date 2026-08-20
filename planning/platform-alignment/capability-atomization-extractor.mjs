import fs from "node:fs";

const rows = fs.readFileSync("planning/platform-alignment/design-capability-ledger.tsv", "utf8").trimEnd().split("\n").slice(1).map((line) => {
  const [id, designRef, outcome, preliminary, route] = line.split("\t");
  return { id, designRef, outcome, preliminary, route };
});

const split = {
  "V-001": ["Play without real-money purchase or payout paths", "Play without advertising", "Play without functional NFT or token monetization"],
  "V-002": ["Progress through a viable idle build", "Progress through a viable active build", "Choose genuinely different idle and active verbs"],
  "V-003": ["Use a bot fallback for global milestone participation", "Use a bot fallback for guild-dependent play", "Use a bot fallback for direct PvP", "Use a bot fallback for each multiplayer minigame", "See every bot identified honestly", "Run every bot through the same validation rules as a human"],
  "V-004": ["Reach a designed terminal ending", "Choose among three terminal endings", "Submit a completed run to a leaderboard"],
  "V-005": ["Encounter a distinct mechanic grammar at each tier", "Have each new tier automate prior-tier work"],
  "V-006": ["Use attended-time mechanics", "Use wall-clock mechanics", "Use fixed-budget or swap-clock mechanics", "Use production-immune Fiscal time"],
  "V-007": ["Play without paid FOMO or absence cost", "See every parodied dark pattern disclosed", "Have gameplay telemetry minimized", "See collected gameplay telemetry disclosed"],
  "V-008": ["Have production results decided server-side", "Read published community/contribution formulas"],
  "V-009": ["Play in a browser", "Complete core play on a mobile-width viewport", "See every cap as a visible hard number", "Use in-game tools instead of a required external wiki", "Compete without bots or players receiving unearned rule exceptions"],
  "T-000": ["Perform the Tier-0 manual action", "Buy and benefit from Tier-0 generators", "Use the honest 1995 presentation", "Understand Tier-0 play through diegetic teaching"],
  "T-001": ["Progress through the Tier-1 generator cost curve", "Buy and benefit from Tier-1 upgrades", "Make a meaningful Tier-1 buy-order choice", "Automate Tier-0 manual work", "Encounter the free-shop satire"],
  "T-002": ["Allocate finite Tier-2 headcount among roles", "Observe Soul in Tier 2", "Change Soul through Tier-2 choices", "Choose a faction through a Faustian bargain", "Access the first major minigames", "Take the first elective Exit"],
  "T-003": ["Price cloud capacity against tenant demand", "Manage distinct tenant workload curves", "Meet tenant SLA constraints", "Choose an irreversible enshittification stage", "Automate headcount management", "Operate the Bakery tenant", "Use the tenant daemon lifecycle"],
  "T-004": ["Draft a region on the shared map", "Manage physical capacity constraints", "See every rack/site in the diorama", "Record addressed Externality entries", "Automate pricing"],
  "T-005": ["Allocate compute data and safety research", "Commit to timed model-training runs", "Experience recursive model collapse", "Choose a Safety or e/acc research route", "Automate datacenter operation"],
  "T-006": ["Author a founder policy script", "Execute a founder policy script", "Save a policy script", "Name a policy script", "Share a policy script", "Manage founder mortality through longevity choices", "Use billionaire-layer Personal Wealth", "Automate active lower-tier play"],
  "T-007": ["Optimize an increasingly self-playing system", "Use TAS-like route planning", "Choose a Transcendence path", "Resolve founder Soul consequences"],
  "T-008": ["Manage Planet depletion as the terminal grammar", "Choose a terminal response to depletion", "See the shared world resolve the ending"],
  "T-009": ["Reach the Ascension ending", "Reach the Long Decay ending", "Reach the Both Graphs ending", "Reach designed ending variants"],
  "T-010": ["Receive tier-flavored Exit offers", "Choose to accept or decline an Exit offer", "Carry Founder progression across runs", "Complete a multi-run arc"],
  "E-001": ["Persist Company-scoped state", "Persist Founder-scoped state", "Persist World-scoped state", "Persist Guild-scoped state"],
  "E-002": ["Apply declarative geometric generator costs", "Compute exact bulk affordability"],
  "E-003": ["Apply the production stack in declared order", "Show the contribution formula to the player", "Clamp each configured production hardcap visibly"],
  "E-004": ["Activate deterministic attended-time buff windows", "Apply the Lucky bank formula", "Validate active clicks server-side", "Use attach accrue and pop daemon mechanics"],
  "E-005": ["Accrue offline production at 90 percent", "Clamp one offline span to 24 hours", "Show the credited offline result on return"],
  "E-006": ["Apply mandatory economy sinks", "Mint currencies only through declared cap owners", "Evaluate achievement provenance proof", "Evaluate achievement possession proof", "Evaluate achievement burn proof"],
  "E-007": ["Earn and spend Reputation", "Carry designated Network slots", "Earn and spend Route Knowledge", "Earn and use Clout", "Extract and spend Personal Wealth", "Advance Founder Age"],
  "E-008": ["Receive full-terms acquihire or acquisition offers", "Wind Down a run on demand", "Complete a doctrine-gated IPO Exit", "Strip-mine a branch before Exit"],
  "E-009": ["Observe the shared Planet ratchet", "Contribute to a community milestone", "Accrue and spend guild tithe value", "Earn and spend Influence"],
  "E-010": ["Advance an Earnings Call on wall time only", "Harvest early with declared failure probability", "Receive guaranteed or automatic harvest", "Spend Investor Confidence", "Choose between hoarding and spending Confidence"],
  "E-011": ["Mint Clout from the single declared social source", "Persist Clout without adding it directly to production", "Spend lifetime Clout on reach", "Use run-local Clout through declared production effects"],
  "E-012": ["Observe ten independent Trust bars", "Observe Soul separately from Trust", "Record Externality as addressed ledger entries", "Keep p-doom as a pressure meter", "Gate routes on dated facts", "Gate endings on dated facts", "Detect Trust-Soul correlation drift"],
  "E-013": ["Drain Soul from declared choices", "Lock human content by Soul band", "Recover Soul through attended zero-output activity", "Prevent passive idle Soul refill", "Resolve the terminal ending from Soul"],
  "E-014": ["Accrue capped Compute Credits from offline time", "Spend Compute Credits on chosen acceleration", "See banked time in the primary HUD", "Opt into automatic Compute Credit spending", "Receive a visible return bonus after absence"],
  "E-015": ["Complete a daily activity bar with a done state", "Accumulate lifetime streak days without reset", "Use the published weekly theme calendar", "Compare with one bracketed weekly rival", "Spend event scrip at posted exchange-shop prices", "Auto-convert leftover event scrip", "Keep daily systems from gating tier or Fiscal progress"],
  "E-016": ["Reach the first generator within target", "Reach Tier 1 within target", "Reach the first minigame within target", "Reach the first elective Exit within target", "Reach Tier 3 within target", "Establish the first Earnings Call within target", "Reach Tier 4 within target", "Reach Tier 5 within target", "Reach Tier 6 within target", "Reach the first ending within target"],
  "E-017": ["Keep each purchasable relevant to at least one persona", "Detect trap purchases in the reference persona", "Publish per-tier contribution shares by epoch"],
  "E-018": ["Preserve Go and TypeScript big-number parity", "Use canonical decimal strings on the wire", "Refuse to persist NaN state"],
  "M-001": ["Declare the clock for every minigame", "Declare the economy hook for every minigame", "Provide an honest fallback for every multiplayer minigame", "Stagger minigame unlocks", "Persist minigame state and results", "Freeze scaling inputs at session creation", "Apply offline-quality policy"],
  "M-006": ["Play each designed board game", "Enter flat ranked board-game play", "Receive bot matchmaking for board games"],
  "M-007": ["Enter a regional parlor", "Use social parlor rituals", "Play the declared parlor game taxonomy"],
  "M-011": ["Open the Demo Disc Arcade", "Play a real arcade tenant from the Demo Disc"],
  "M-013": ["Play the lane-pusher minigame", "Receive an honest lane-pusher bot opponent"],
  "M-014": ["Play a pet battle", "Receive an honest pet-battle bot opponent"],
  "M-015": ["Complete a five-minute daily minigame session", "Stop without losing progress or rewards after the daily session"],
  "M-016": ["Freeze tier scaling inputs at minigame creation", "Apply declared fairness bounds", "Decay offline automation quality visibly"],
  "M-017": ["Create and play a Pitch session", "Reconnect to an in-progress Pitch session", "Resolve Pitch rewards", "Complete Pitch through a mounted player surface"],
  "P-001": ["Acquire a Founder companion", "Apply offline-safe pet care decay", "Perform a pet care action", "Observe pet behavior state", "Observe pet bond or trust"],
  "P-002": ["Resolve a hardcapped pet battle", "Apply care state to battle reliability", "Play pet battles against an honest bot"],
  "P-003": ["Decorate a cosmetic house", "Visit another player's house without destructive interaction"],
  "P-004": ["Open a free cosmetic lootbox", "See exact cosmetic odds and dark-pattern disclosure", "Equip a cosmetic reward"],
  "P-005": ["Breed cosmetic pet traits", "Retire an elder pet without punitive loss"],
  "P-006": ["Render a pet from CSS parts", "Respect reduced motion in creature rendering"],
  "S-001": ["Contribute to a global milestone", "Receive a personal contribution reward", "Unlock a global milestone tier", "Observe canonical milestone failure", "Receive mercy-scaled NPC contribution"],
  "S-002": ["Read a live activity feed", "Observe truthful live world counters", "See player ghosts without exposing hidden data", "See labeled NPC activity when population is low"],
  "S-003": ["Create join leave and dissolve a guild", "Exchange guild resources through declared verbs", "Participate in a weekly guild ritual", "Interact with an NPC guild network"],
  "S-004": ["Play non-destructive direct PvP", "Receive honest bot fallback for each PvP mode"],
  "S-005": ["Enter the Commons from a player surface", "Choose Commons governance actions", "Receive a stable cohort assignment", "Inspect the published Commons formula"],
  "S-006": ["Submit a verified run to a board", "Browse derived leaderboard ranks", "Choose a declared run category", "Register and execute named routes", "Inspect public run evidence"],
  "S-008": ["Retain ownership of player-authored content", "Receive attribution for player-authored content", "Preserve fork lineage for player-authored content"],
  "A-001": ["Run the authoritative Go game server", "Persist production state in Postgres", "Use the Svelte browser application", "Receive realtime updates through Centrifuge", "Serve site pages through Astro", "Proxy production traffic through Caddy", "Deploy the production stack through Compose"],
  "A-002": ["Evaluate idle production lazily from elapsed time", "Keep authoritative production on the server", "Keep client prediction numerically aligned with authoritative production"],
  "A-003": ["Rate-limit abusive requests", "Deduplicate repeated intents", "Reject invariant-violating transitions", "Validate minigame commands authoritatively", "Verify ranked runs before projection", "Reject invalid or poisoned saves"],
  "A-004": ["Match a player to a minigame opponent", "Run an honest AI opponent through human validation", "Isolate per-game engine rules"],
  "A-005": ["Start with a server-anonymous identity by default", "Continue in a labeled local-only outage mode", "Import local outage progress into the server account"],
  "RD-000": ["Ship the numeric foundation", "Ship the economy and production foundation", "Ship versioned save persistence", "Ship the balance harness foundation", "Ship the client shell foundation", "Complete a private Tier-0-to-Tier-1 vertical slice"],
  "RD-001": ["Ship public Tier-0-to-Tier-2 play", "Ship early player minigames", "Ship the player pet loop", "Ship player presence"],
  "RD-002": ["Ship Tier-3 player progression", "Ship player faction workflows", "Ship player guild workflows", "Ship global milestones", "Ship personal events", "Ship pressure-meter events"],
  "RD-003": ["Ship Tier-4 player progression", "Ship the shared world workflow", "Ship the Lane minigame", "Ship player leaderboard workflows", "Ship server-wide Situation events"],
  "RD-004": ["Ship Tier-5 player progression", "Ship the Commons player workflow", "Ship combat gameplay", "Ship challenge runs", "Ship the Founder arc"],
  "RD-005": ["Ship Tier-6 player progression", "Ship Tier-7 player progression", "Ship Tier-8 player progression", "Ship every designed terminal ending"],
  "F-001": ["Render player text in the declared voice", "Ship factual claims only with verified sourcing", "Keep player text within declared target bounds"],
  "F-002": ["Render an era-appropriate presentation for each tier", "Degrade presentation in step with enshittification"],
  "F-003": ["Ship the launch-scale generator copy bank", "Ship the launch-scale event copy bank", "Ship the launch-scale achievement copy bank", "Ship the launch-scale lore copy bank"],
  "F-004": ["Encounter the designed conspiracy narrative arc", "See conspiracy pressure respond to current world state"],
  "F-005": ["Trigger media canonization events from Clout", "Choose a response to media canonization"],
  "F-006": ["See the speedrun timer and per-tier splits", "Choose a governed speedrun category", "Discover and name a route", "Review a completed run retrospectively"],
  "F-007": ["Open the shipped honesty appendix", "Inspect every parodied dark pattern and its real-world mechanism"],
  "EV-001": ["Load declarative event definitions", "Dispatch eligible events on an action", "Apply only closed event effects", "Persist event choices and outcomes"],
  "EV-002": ["Receive a personal event", "Progress a founder story cycle", "See personal-event consequences in later choices"],
  "EV-003": ["Observe a visible pressure meter", "Forecast pressure-meter movement", "Trigger content at a meter band transition"],
  "EV-004": ["Participate in a phased Situation", "Participate in a Major Order", "Receive a personal objective floor during world events"],
  "EV-005": ["Earn event scrip", "Spend scrip at posted prices", "Receive deterministic event-end conversion"],
  "EV-006": ["Play a seasonal story arc", "Access an evergreen version after the season"],
  "EV-007": ["Operate events through a GM dashboard", "Read a public war log", "Detect stuck or invalid events through watchdogs"],
  "PS-001": ["Choose one of four rule-distinct factions", "Produce a resource another faction consumes", "Use a faction-specific gameplay verb"],
  "PS-002": ["Play an active build", "Play an idle build", "Play a check-in build", "Play a banker build", "Compete on a build-specific board"],
  "PS-003": ["Toggle a reversible ideology", "Use each ideology as a viable rule set"],
  "PS-004": ["Choose one of three doctrines at a tier transition", "Keep doctrine choice irreversible within the run", "Reach viable content and endings through each doctrine"],
  "PS-005": ["Start a rule-modified challenge run", "Complete and verify a challenge run"],
  "PS-006": ["Complete the Canon moral route", "Complete Ethical percent", "Carry moral history across runs"],
  "PS-007": ["Author an AGI policy script", "Share an AGI policy script", "Execute a policy script authoritatively"],
  "UX-001": ["Read an honest cold-open contract", "Begin the first attempt from the contract screen", "Return to a bounded post-absence summary"],
  "UX-002": ["Receive a server-anonymous account by default", "Continue locally during a server outage", "Use the Vision Slide first-session surface", "Observe the run-one presence budget"],
  "UX-003": ["Learn mechanics through diegetic prompts", "Discover the core loop within ten minutes", "Dismiss guidance without losing access to it"],
  "UX-004": ["See the run-ending cause", "See attended-time and Exit context", "See Founder progression deltas", "See the chosen route or new-route consequence", "Continue into the next run", "Experience the scripted first-failure branch"],
  "UX-005": ["Enable Advisor Mode", "Receive Advisor guidance without hidden rule changes"],
  "UX-006": ["Progress through a multi-run first-ending arc", "See run-count pacing communicate long-term commitment"],
  "UX-007": ["Ship the declared launch copy surface count", "Cover every mounted launch surface with governed copy"],
  "UX-008": ["Resolve player copy from data keys", "Reject invalid copy parameters", "Lint prohibited or unsupported claims", "Attach citations to verified claims", "Preserve a localization-ready copy shape"],
  "UX-009": ["Reach core actions within navigation depth two", "Reflow core UI at mobile width", "Reflow core UI at supported zoom", "Complete core play by keyboard", "Complete core play with assistive technology"],
  "CP-001": ["Load a versioned additive content pack", "Stage a content pack before activation", "Activate a pack without mutating historical epochs"],
  "CP-002": ["Extend economy content through data", "Extend routes through data", "Extend achievements through data", "Extend factions through data", "Extend guilds through data", "Extend minigames through their socket", "Extend events through the closed effect vocabulary", "Extend presentation copy through governed catalogs"],
  "CP-003": ["Create every minigame through one API socket", "Persist every minigame through the shared session contract"],
  "CP-004": ["Select event effects from a closed vocabulary", "Reject an event with an undeclared effect verb"],
  "CP-005": ["Research a current event before authoring", "Activate current-event content in a new epoch", "Retain prior-epoch replay behavior"],
  "CP-006": ["Load saves against their pinned pack set", "Migrate compatible pack-set saves", "Retire a pack without corrupting historical saves"],
  "W-001": ["Render a state-derived personal diorama", "Fast-forward the diorama on return", "Keep the diorama non-authoritative"],
  "W-002": ["Place adoption or regulation pressure on regions", "Watch the spatial campaign respond to player actions"],
  "W-003": ["Observe shared world state", "Observe a regional Planet ratchet", "Inspect a public immutable world artifact"],
  "W-004": ["Draft one of offered regions", "Place infrastructure under the Thermal Law", "Reveal world geography through placement"],
  "W-005": ["Interact with the Pixi or WebGPU world view", "Open a baked accessible world overview", "Use world rendering at supported scale"],
};

const children = [];
for (const row of rows) {
  const outcomes = split[row.id] ?? [row.outcome];
  outcomes.forEach((outcome, index) => {
    children.push([`${row.id}.${String(index + 1).padStart(2, "0")}`, row.id, row.designRef, outcome, row.preliminary, row.route]);
  });
}

const minimumSplits = { "V-003": 6, "T-002": 6, "E-007": 6, "M-001": 7, "S-003": 4, "A-003": 6, "UX-004": 6, "CP-002": 8 };
function validate(values) {
  if (rows.length !== 121 || new Set(rows.map((row) => row.id)).size !== 121) throw new Error("parent denominator failure");
  if (new Set(values.map((row) => row[0])).size !== values.length) throw new Error("duplicate child ID");
  const byParent = new Map(rows.map((row) => [row.id, []]));
  for (const child of values) {
    if (!byParent.has(child[1])) throw new Error(`unknown parent ${child[1]}`);
    byParent.get(child[1]).push(child);
  }
  for (const [parent, childRows] of byParent) {
    if (childRows.length === 0) throw new Error(`unmapped parent ${parent}`);
    for (let index = 0; index < childRows.length; index++) if (childRows[index][0] !== `${parent}.${String(index + 1).padStart(2, "0")}`) throw new Error(`child sequence failure ${parent}`);
  }
  for (const [parent, minimum] of Object.entries(minimumSplits)) if (byParent.get(parent).length < minimum) throw new Error(`mandatory split failure ${parent}`);
  if (new Set(values.map((row) => row[3])).size !== values.length) throw new Error("duplicate child outcome");
}
validate(children);
const seededFailures = [];
function mustReject(label, operation) {
  try { operation(); } catch { seededFailures.push(label); return; }
  throw new Error(`seeded failure was accepted: ${label}`);
}
mustReject("dropped-parent-child", () => validate(children.filter((row) => row[1] !== "M-002")));
mustReject("duplicate-child-id", () => validate([...children, children[0]]));
mustReject("mandatory-no-op-split", () => validate(children.filter((row) => row[1] !== "A-003" || row[0] === "A-003.01")));
console.log(["capability_id", "parent_id", "design_ref", "user_outcome", "parent_preliminary_state", "parent_route"].join("\t"));
for (const child of children) console.log(child.join("\t"));
console.error(JSON.stringify({ parents: rows.length, children: children.length, splitParents: Object.keys(split).length, atomicParents: rows.length - Object.keys(split).length, seededFailures }));
