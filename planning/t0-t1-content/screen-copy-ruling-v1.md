# The Phase-A screen-copy ruling (GU-C12 / T01-C4 / FCE-C7 lane) — v1, 2026-08-10

**Authored by:** Claude (owner lane), from the commissioned draft, reviewed string-by-string.
**Status:** RULED — these texts are the owner-authored literals. Codex assembles the copy
documents byte-exactly from this file (orphan-first per FCE-C8), files assembly SHAs for Marco's
ratification. Any text change after ratification is a copy retune through the normal pipeline.

## Rulings on the draft's gaps and flags

1. **permits cap row** — ADOPTED as authored (new row).
2. **cash cap replacement** — ADOPTED (re-authoring the landed placeholder row is a legitimate
   copy retune; the new text states the same law better).
3. **`event.exit_offer_resolved`** (the GU-C11 event) — AUTHORED HERE, closing the draft's gap 3:
   `event.exit_offer_resolved` — params: (none) — tone: corporate —
   Text: "Offer accepted. The rest is paperwork."
   The event-copy candidate gains this row in its revision (a ninth… eighth-kind addition to the
   closed set, with the T01-C9 unknown-kind law unchanged).
4. **Gate titles** — the manifest round gains a `gate.*.title` family;
   `gate.t0_to_t1.title` = "Garage" ADOPTED (design/01's split name). Later gates authored as
   their tiers draft.
5. **Exit-type display names** — the manifest round gains an `exit_type.*.title` family; Codex
   enumerates the shipped exit-type IDs and the texts are authored in the follow-up ruling (no
   raw IDs may render meanwhile — the frame falls back to the curriculum/standard titles).
6. **Ladder-rung copy** — DEFERRED, named gap (no ruled display surface exists; when the ladder
   toast/tooltip surface is designed, the copy follows).
7. **Company naming** — the fallback string stands; the naming flow is a named successor
   (post-Phase-A content), not a copy gap.

Flag dispositions: "Elbow Grease" ADOPTED; the Vision Slide slide texts ADOPTED; the
mechanical-ID substitution contract is ROUTED to Game-UI implementation (the feed renderer MUST
substitute presentation titles — mechanical IDs never render; this is a GU acceptance-test
requirement, recorded here); legal_dept never renders in Phase A (its window opens at t2_to_t3)
— non-issue, noted; the offline-merge non-promise is correct and stays (spec gap named for the
bootstrap successor); the tone/era linter tag is at Codex's discretion; the two over-length
strings are approved by slot.

---

# Phase-A screen-copy set — DRAFT for refinement + owner ruling

Status: **draft — NOT committed, NOT ratified.** Authored per GU-C12 (FCE-C7 orphan-first lane).
Destination on ratification: Copy catalog rows (phase0 grammar: `key / text / params /
era_variants / provenance / tone`), consumed byte-exactly by the T0–T1 screen-copy manifest.

Grounding: `design/08-satire-flavor.md §1` (voice, binding), `design/11-ux-writing.md` (+ §1b
adoptions), `design/01-tiers.md` T0/T1, `design/research/era-1995-satire.md` (esp. §3, §5, §6),
`design/research/culture-genx.md`, `balance/testdata/t0-t1/presentation-v1.json` /
`event-copy-v1.json` / `curriculum-v1.json` / `economy-v4.json` (mechanical truth),
`rfc/game-ui-screens.md` U1 + GU-C1–C12 rulings, `copy/catalog/phase0.json` (grammar).

Register rules applied throughout:
- **T0 = earnest webmaster/shareware author** (era-1995 dossier §5): first person, warm,
  sincere `!!` legal, guests not users. Banned in T0 rows: *engagement, users, metrics,
  growth, content, monetize* (copy-linter era boundary).
- **T1 = 2000 web-1.0 / dot-com**, first cracks of corporate; no double exclamation marks.
- **Curtain rule** (law 10 / 08 §1.2): every parodied pattern states what it is, in the copy.
- **No real-world statistics** anywhere in this set. The only real-world *claims* are
  period-generic and dossier-grounded (noted inline). The 90% / 24 h numbers are game canon
  (`docs/production-engine.md`), not research claims.
- Tones use the existing catalog vocabulary (`corporate` / `diegetic`). See Flags for a
  possible `webmaster` tone tag.
- Format below: key, params, tone; `Text:`; `era_2000:` variant only where the register
  genuinely shifts; `Why:` only where the joke needs defending.

---

## 1. Vision Slide (surface `vision_slide`) — 12 strings

The ten-second pitch-deck cold open (future Tier-7 dashboard), then the smash-cut to 1995 and
the contract screen. Silent anonymous bootstrap — NO signup form. First run shows `PB: —` and
NO world-record line (binding, 11 §1b / GU ruling).

- `screen.vision_slide.slide_heading` — tone: corporate
  Text: "OUR VISION"
- `screen.vision_slide.slide_body` — tone: corporate
  Text: "One dashboard. One number. A world that runs itself — optimized, forever."
  Why: the Tier-7 referent stated as a pitch-deck promise; gives the timer chrome its meaning
  before the smash-cut strips it away.
- `screen.vision_slide.slide_footnote` — tone: corporate
  Text: "Founder deck — slide 1 of 1."
  Why: a one-slide deck is the era-5 gag (the vision IS the product); deadpan, no wink.
- `screen.vision_slide.skip` — tone: corporate
  Text: "Skip"
- `screen.vision_slide.contract_title` — tone: corporate
  Text: "HUMANITY"
- `screen.vision_slide.contract_category` — tone: corporate
  Text: "Category: Destroy Any%"
- `screen.vision_slide.timer_frame` — params: rta, pb — tone: corporate
  Text: "Timer: {rta} — PB: {pb}"
  (First run renders `{pb}` from `chrome.run_title.pb_empty`. No WR fragment exists in this
  set, deliberately — the no-fake-WR rule; the WR fragment is authored only when validated
  boards exist.)
- `screen.vision_slide.begin_attempt` — tone: corporate
  Text: "BEGIN ATTEMPT"
- `screen.vision_slide.small_print` — tone: diegetic
  Text: "Free forever. No purchases. No ads. The only thing this game harvests is the fictional planet."
  (Verbatim from design/11 §1 — the honesty voice; do not edit without touching design/11.)
- `screen.vision_slide.connecting` — tone: diegetic
  Text: "Dialing…"
  Why: bootstrap wait state worn as 1995 chrome; the smash-cut has already happened when this
  shows.
- `screen.vision_slide.offline_fallback` — tone: diegetic
  Text: "OFFLINE — the server did not pick up. You can play on this machine only for now."
  (Labeled OFFLINE fallback per 11 §1b; makes no promise about later merging — that contract
  is not specified anywhere I could find. See Flags.)
- `screen.vision_slide.retry` — tone: corporate
  Text: "Try Again"

## 2. Navigation / chrome — 15 strings

Surfaces per GU-C2 ruling: `vision_slide`, `desk`, `run_end`, `offer_sheet`, `settings`.
Components: run-title bar, visitor counter (persistent chrome), splits panel (desk child).

- `surface.desk.title` — tone: diegetic
  Text: "The Desk"
- `surface.run_end.title` — tone: corporate
  Text: "End of Run"
- `surface.offer_sheet.title` — tone: corporate
  Text: "The Offer Sheet"
- `surface.settings.title` — tone: corporate
  Text: "Settings"
  era_1995: "Options"
  Why: 1995 software had Options; "Settings" arrives with the millennium. Cheap, correct.
- `chrome.run_title.rta_label` — tone: corporate
  Text: "RTA"
- `chrome.run_title.pb_label` — tone: corporate
  Text: "PB"
- `chrome.run_title.pb_empty` — tone: corporate
  Text: "—"
  (The bar states, it does not judge, until a PB exists — 08 §6 timer semantics.)
- `chrome.run_title.tier_frame` — params: tier — tone: corporate
  Text: "Tier {tier}"
- `chrome.run_title.company_fallback` — tone: corporate
  Text: "UNTITLED COMPANY"
  (Fallback for the bar's COMPANY fragment; no company-naming flow exists in Phase A — see
  Gap list.)
- `chrome.splits.label` — tone: corporate
  Text: "Splits"
- `chrome.splits.first_attempt_note` — tone: diegetic
  Text: "First attempt. The comparison column stays empty until there is a personal best to compare."
  Why: the non-judgment rule rendered as plain fact — states, never consoles.
- `chrome.visitor_counter.frame` — params: count — tone: diegetic
  Text: "You are visitor #{count}"
- `chrome.visitor_counter.tooltip` — tone: diegetic
  Text: "This counter is real. It counts everyone here right now, one at a time. Each one is a person."
  Why: the run-1 presence signal (11 §1b) with its curtain pulled — period furniture admitting
  it is the live player count; "each one is a person" is the dossier's hit-counter thesis and
  pays off later tiers by contrast.
- `desk.generators_label` — tone: diegetic
  Text: "Equipment"
  era_2000: "Infrastructure"
  Why: the T0→T1 rename is a deliberate micro-loss (dossier §5: word shifts are beats).
- `desk.upgrades_label` — tone: corporate
  Text: "Upgrades"

## 3. The Desk — 10 strings

Manual action = `manual.click` (token bucket rendered as in-fiction energy). Buy 1 / buy max
over the afford fast path. Cap explanations render `reason_key` copy — never a frozen number.

- `desk.buy_one` — tone: corporate
  Text: "Buy 1"
- `desk.buy_max` — tone: corporate
  Text: "Buy Max"
- `desk.owned_frame` — params: count — tone: corporate
  Text: "Owned: {count}"
- `desk.rate_frame` — params: rate — tone: corporate
  Text: "{rate}/sec"
- `desk.manual.meter_label` — tone: diegetic
  Text: "Elbow Grease"
  Why: the token bucket needs an in-fiction name; nothing in design names it (see Flags).
  "Elbow Grease" is period-idiomatic, honest about being labor, and not a resource you buy.
- `desk.manual.meter_frame` — params: current, cap — tone: corporate
  Text: "{current}/{cap}"
- `desk.manual.meter_tooltip` — tone: diegetic
  Text: "You have two hands. The meter refills on its own, free, always. It is a speed limit, not a store."
  Why: energy bars are the genre's original sin; ours exists (rate cap) so the curtain states
  exactly what it is and is not. "Not a store" is the whole anti-monetization law in five words.
- `desk.capped_label` — tone: corporate
  Text: "AT CAP"
- `resource.company_permits.cap.phase0` — tone: diegetic
  Text: "The permit drawer is full. Caps here are hard, printed, and not for sale."
  (NEW row — economy-v4 references this reason_key; no catalog row exists. Gap list item 1.)
- `resource.company_cash.cap.phase0` — tone: diegetic — **PROPOSED REPLACEMENT, flagged**
  Text: "Cash is capped. The cap is a number, the number is visible, and nothing will ever sell you the difference."
  (Row EXISTS in phase0.json as "Company cash is capped." — re-authoring a landed row is an
  owner decision; both texts satisfy the law. Gap list item 2.)

## 4. Generators — 9 titles + 9 descriptions + 1 cap reason = 19 strings

IDs and key names from presentation-v1.json; mechanical truth from economy-v4.json (tiers,
roles). T0 rows: beige_tower, dot_matrix_queue, answering_machine, nephew_intern. T1 rows:
garage_rack, crt_wall, first_hire, beige_tower_v2. legal_dept is tier 3 in economy-v4 — see
Flags. Flavor never states balance numbers; role flavor stays qualitative.

- `generator.beige_tower.title` — tone: diegetic — Text: "Beige Tower"
- `generator.beige_tower.description` — tone: diegetic
  Text: "The computer. Beige, sturdy, warm on top. It fixes the small jobs by itself while you handle the big ones."
- `generator.beige_tower.provisioned_cap` — tone: diegetic
  Text: "Provisioned towers stop at the printed limit. Every cap in this program is a number you can read."
  (Reason_key for the provisioned hardcap; states the hardcap law in the webmaster voice.)
- `generator.dot_matrix_queue.title` — tone: diegetic — Text: "Dot-Matrix Queue"
- `generator.dot_matrix_queue.description` — tone: diegetic
  Text: "Invoices print all night on tractor-feed paper. The rip along the perforation is the sound of getting paid!!"
- `generator.answering_machine.title` — tone: diegetic — Text: "Answering Machine"
- `generator.answering_machine.description` — tone: diegetic
  Text: "Takes jobs while you're out!! Every blink of the red light is a neighbor who didn't hang up."
  Why: guests-not-users rule — customers are neighbors in T0; the machine's mechanical role
  (clicks worth more) reads as a fuller job list without stating math.
- `generator.nephew_intern.title` — tone: diegetic — Text: "The Nephew"
- `generator.nephew_intern.description` — tone: diegetic
  Text: "Your sister's kid. Great with computers. Paid in pizza and experience. Keeps a stack of finished jobs ready for Monday."
  (The stack line is the stock_rate role in fiction; the pizza line is period-true unpaid-labor
  satire kept warm — T0 never mocks its own people.)
- `generator.garage_rack.title` — tone: diegetic — Text: "Garage Rack"
- `generator.garage_rack.description` — tone: diegetic
  Text: "A real server rack in the garage, between the paint cans and the bikes. The plaque out front calls it a data center."
  Why: design/01 T1's garage mythology — the plaque that produces nothing but investor
  interest, compressed to one sentence.
- `generator.crt_wall.title` — tone: diegetic — Text: "CRT Wall"
- `generator.crt_wall.description` — tone: diegetic
  Text: "Six monitors on one desk. Nothing requires six monitors. Investors require you to have six monitors."
  Why: the 2000 register's first crack — spending shaped by observers, not work. Rule-of-three
  deadpan, no wink.
- `generator.first_hire.title` — tone: diegetic — Text: "First Hire"
- `generator.first_hire.description` — tone: diegetic
  Text: "Employee #2, paid partly in options. The handshake felt historic to both of you. Everything goes further with two of you."
  (Options-in-lieu-of-salary is the Gen-X dossier's T1 seed; "felt historic" carries the
  dramatic irony without the copy knowing it.)
- `generator.beige_tower_v2.title` — tone: diegetic — Text: "Beige Tower v2"
- `generator.beige_tower_v2.description` — tone: diegetic
  Text: "Procurement now orders towers by itself, on a schedule. You signed something. New Beige Towers just arrive."
  Why: the provision role in fiction; "you signed something" is the first appearance of the
  game's central move — automation you authorized without reading.
- `generator.legal_dept.title` — tone: corporate — Text: "Legal Department"
- `generator.legal_dept.description` — tone: corporate
  Text: "Files permits, slowly, correctly. Also accumulates institutional knowledge, most of which is the word no."
  (Tier-3 row, corporate register by design; era-skin question flagged.)

## 5. Upgrades — 10 titles + 10 descriptions = 20 strings

Every upgrade is mechanically a 2x factor; copy renders "doubling" as fiction, never as a
stated multiplier. T0 window: reply_all_macro, beige_tower_cache, continuous_feed_paper,
nephew_business_cards, hold_music_license. T1 window: the rest.

- `upgrade.reply_all_macro.title` — tone: diegetic — Text: "Reply-All Macro"
- `upgrade.reply_all_macro.description` — tone: diegetic
  Text: "One keystroke answers every message at once, mostly with 'Received!!'. Twice the correspondence, twice the business."
- `upgrade.beige_tower_cache.title` — tone: diegetic — Text: "Beige Tower Cache"
- `upgrade.beige_tower_cache.description` — tone: diegetic
  Text: "Spare RAM in a shoebox under the desk. The tower finds everything twice as fast now."
- `upgrade.continuous_feed_paper.title` — tone: diegetic — Text: "Continuous-Feed Paper"
- `upgrade.continuous_feed_paper.description` — tone: diegetic
  Text: "Five thousand sheets, all attached to each other. The printer never stops to reload. Neither, it turns out, do the invoices."
- `upgrade.nephew_business_cards.title` — tone: diegetic — Text: "Nephew's Business Cards"
- `upgrade.nephew_business_cards.description` — tone: diegetic
  Text: "Five hundred cards, raised lettering. He hands them to everyone he meets. It works. Nobody can explain why it works."
- `upgrade.hold_music_license.title` — tone: diegetic — Text: "Hold Music License"
- `upgrade.hold_music_license.description` — tone: diegetic
  Text: "Licensed fair and square from a man with a keyboard. Callers wait twice as long, contentedly. It's a waltz."
  Why: T0's one licensing joke is *paying* honestly for something small — the era's dignity of
  the transaction (dossier §3), inverted against everything the game becomes.
- `upgrade.rack_rail_standardization.title` — tone: diegetic — Text: "Rack-Rail Standardization"
- `upgrade.rack_rail_standardization.description` — tone: diegetic
  Text: "Every server now slides into every rack. One meeting decided this. It was the company's best meeting, and its last short one."
- `upgrade.crt_degauss_button.title` — tone: diegetic — Text: "Degauss Button"
- `upgrade.crt_degauss_button.description` — tone: diegetic
  Text: "THUNK. The screen wobbles, settles, perfect. Officially it is maintenance. Unofficially everyone takes a turn."
  Why: the dossier's "deeply satisfying" degauss seed, adapted so it may honestly carry a
  production effect (morale is real here); the free-toy version of the joke stays available
  for a later cosmetic.
- `upgrade.employee_handbook_v0.title` — tone: diegetic — Text: "Employee Handbook v0"
- `upgrade.employee_handbook_v0.description` — tone: diegetic
  Text: "Forty pages. Page one says 'we're like a family here.' The other thirty-nine are about liability."
  Why: the We're-A-Family seed lands at first-hire scale, where it is still almost true —
  which is what makes its T2 return land.
- `upgrade.refurbished_sticker.title` — tone: diegetic — Text: "'Refurbished' Sticker"
- `upgrade.refurbished_sticker.description` — tone: diegetic
  Text: "The same machine, plus a sticker. It sells for twice as much now. Write this trick down; the industry will."
  Why: the first honest-margin trick — the sticker is the ancestor of every rebrand upgrade in
  the later tiers; "the industry will" is the dramatic-irony register (§4), no wink.
- `upgrade.institutional_memory.title` — tone: diegetic — Text: "Institutional Memory"
- `upgrade.institutional_memory.description` — tone: diegetic
  Text: "Where everything is, finally written down — in one binder, by the one person who knew. She is no longer load-bearing."

## 6. Manual action + Horse Armor — 5 strings

- `manual.click.title` — tone: diegetic
  Text: "Fix Computer"
  (Binding: design/01 T0 — dollars, one at a time, by clicking Fix Computer.)
- `manual.click.description` — tone: diegetic
  Text: "One repair, one dollar, one happy neighbor. This is the whole business right now. It is honest work."
  Why: T0's thesis sentence — the baseline every later tier is measured against.
- `cosmetic.horse_armor_free.title` — tone: diegetic
  Text: "Horse Armor"
- `cosmetic.horse_armor_free.description` — tone: diegetic
  Text: "Decorative armor for a horse. You do not have a horse. Price: $0.00, forever."
- `cosmetic.horse_armor_free.disclosure` — tone: diegetic
  Text: "In 2006, armor like this cost real money and changed the industry. Ours is free, does nothing, and always will. That is the whole joke."
  Why: law 10's curtain at maximum disclosure; the historical claim is period-generic (design/01
  T1's 2006 first-DLC anchor), no brand, no figure.

## 7. Settings / system — 11 strings

Save status; drain notice (`server_restarting` → diegetic scheduled maintenance); the
`resync_required` full sync as a story beat, not an error (shell D2); the offline-progress
line for the return-sequence header dock.

- `settings.save_status.saved_frame` — params: ago — tone: diegetic
  Text: "Saved to the server {ago} ago."
- `settings.save_status.saving` — tone: corporate
  Text: "Saving…"
- `settings.save_status.offline` — tone: diegetic
  Text: "OFFLINE — progress is parked on this machine until the server picks up again."
- `settings.account_note` — tone: diegetic
  Text: "This account is anonymous and complete. An email, if you ever add one, is only a spare key."
  (Register-is-recovery-only, stated benefit, never a wall — 11 §1b account model.)
- `system.drain_notice.title` — tone: diegetic
  Text: "SCHEDULED MAINTENANCE"
- `system.drain_notice.body` — tone: diegetic
  Text: "The server needs a minute. Your equipment keeps working while it's away, and nothing is lost. Back soon!!"
  era_2000: "Scheduled maintenance in progress. Production continues server-side. Uptime is our passion."
  Why: the same fact in both registers — the T1 variant's "Uptime is our passion" is the first
  time maintenance apologizes like a brand.
- `system.resync.title` — tone: diegetic
  Text: "RECOUNTING THE BOOKS"
- `system.resync.body` — tone: diegetic
  Text: "The two copies of the books disagreed, so everything was recounted. The server's copy wins. It always does. Nothing was lost."
  Why: server-authority stated diegetically and flatly — the story beat is that this is
  bookkeeping, not catastrophe.
- `system.resync.continue` — tone: corporate
  Text: "Back to the desk"
- `system.offline_progress.frame` — params: amount — tone: diegetic
  Text: "While you were away: +{amount}"
- `system.offline_progress.tooltip` — tone: diegetic
  Text: "While you were away the machine kept working at 90%, up to 24 hours. BBS door games did this first, because everyone shared one phone line. They called it fairness. So do we."
  (Over 140 chars — tooltip slot demands it; text adapted near-verbatim from the era-1995
  dossier §6b; numbers are game canon per docs/production-engine.md.)

## 8. Offer sheet (surface `offer_sheet`) — 7 strings

Terms come from the offer event's terms object; accept/decline through existing intents;
countdown from `expires_at`.

- `screen.offer_sheet.heading` — tone: corporate
  Text: "LETTER OF INTENT"
- `screen.offer_sheet.preamble` — tone: diegetic
  Text: "Someone wants to buy the company. The full terms are below — all of them. Nothing on this sheet is hidden."
- `screen.offer_sheet.terms_label` — tone: corporate
  Text: "Terms"
- `screen.offer_sheet.accept` — tone: corporate
  Text: "Sign"
- `screen.offer_sheet.decline` — tone: corporate
  Text: "Decline"
- `screen.offer_sheet.countdown_frame` — params: remaining — tone: corporate
  Text: "On the table for {remaining}"
- `screen.offer_sheet.countdown_tooltip` — tone: diegetic
  Text: "The deadline is real; the terms will not improve or worsen while you decide. This timer is information, not pressure."
  Why: a countdown is the dark-pattern shape, so the curtain must distinguish ours: real
  deadline, immutable terms, no manufactured urgency. Deliberately does NOT promise another
  offer — that guarantee is not in any spec I could find.

## 9. Run-end (surface `run_end`) — 9 strings (incl. the 3 curriculum slots)

Renders exclusively from the `run_ended` payload. Standard variant + the scripted-first
variant (curriculum-v1's three bound keys) + the "run 2 opens with" delta frame (D6 assembly
facts).

- `screen.run_end.exit_frame` — params: exit_type, tier — tone: corporate
  Text: "{exit_type} — Tier {tier}"
  (exit_type display names are a gap — see Gap list item 5.)
- `screen.run_end.attended_frame` — params: attended — tone: diegetic
  Text: "Attended time: {attended}. The clock only counted while you were actually here."
  Why: the attended-time honesty rule (08 §6's RTA/IGT retarget) stated where the player first
  meets the number.
- `screen.run_end.founder_note` — tone: diegetic
  Text: "The company is over. The Founder is not. Everything below carries forward."
- `screen.run_end.delta_heading` — params: run_seq — tone: corporate
  Text: "Run {run_seq} opens with:"
- `screen.run_end.new_route` — tone: corporate
  Text: "NEW ROUTE"
- `screen.run_end.standard.title` — tone: corporate
  Text: "The Company Has Exited"
- `curriculum.scripted_first_failure.title` — tone: diegetic
  Text: "Your First Company Failed"
- `curriculum.scripted_first_failure.body` — tone: diegetic
  Text: "Your first company failed. Statistically, this is the most realistic thing in this game. Everything it taught you survives it — see below."
  (First sentence pair is the canonical line from design/11 §3, verbatim; the close points at
  the Founder card so the screen teaches what it exists to teach.)
- `curriculum.scripted_first_failure.next_run` — tone: diegetic
  Text: "The second company gets founded by someone who has already failed once. Historically, that is the good kind of founder."
  Why: the survivorship-bias joke run in reverse — failure reframed as the credential, which
  is both the Hades rule (story advances regardless) and true to the era's own mythology.

## 10. Event copy — 7 strings (the closed seven-kind set, event-copy-v1)

Feed/log register: flat, factual, deadpan. Parameter names are exactly the registered ones.
NOTE: `generator_id` / `upgrade_id` / `gate_id` are mechanical IDs — the renderer must
substitute presentation titles or the feed leaks mechanical names (Flags item 3); `_ms`
params assume renderer-side formatting.

- `event.exit_offer_spawned` — params: exit_type, expires_at_ms — tone: corporate
  Text: "An offer is on the table: {exit_type}. It stands until {expires_at_ms}."
- `event.exit_offer_declined` — params: run_seq — tone: corporate
  Text: "Offer declined. Run {run_seq} continues."
- `event.exit_offer_expired` — params: (none) — tone: corporate
  Text: "The offer lapsed, unanswered. The company is still yours."
- `event.gate_crossed` — params: gate_id — tone: corporate
  Text: "Split: {gate_id}. Time recorded."
- `event.generator_purchased` — params: generator_id, count, cost — tone: corporate
  Text: "Purchased {count} × {generator_id} for {cost}."
- `event.upgrade_purchased` — params: upgrade_id, cost — tone: corporate
  Text: "Installed {upgrade_id} for {cost}."
- `event.run_ended` — params: exit_type, attended_ms, tier — tone: corporate
  Text: "Run over — {exit_type}, Tier {tier}, {attended_ms} attended."

## 11. T0 satire elements — 13 strings

The ruled §1b composition: ORDER NOW $0.00 (the beat), the UNREGISTERED count-up nag
(standing chrome), README.TXT (the codex skin). All curtains at maximum disclosure.

### ORDER NOW: $0.00 (9 strings)

- `satire.order_form.window_title` — tone: diegetic — Text: "ORDER.FRM"
- `satire.order_form.heading` — tone: diegetic — Text: "ORDER NOW"
- `satire.order_form.item_full_version` — tone: diegetic
  Text: "Full version ............ $0.00"
- `satire.order_form.item_site_license` — tone: diegetic
  Text: "Site license ............ $0.00"
- `satire.order_form.item_shipping` — tone: diegetic
  Text: "Shipping & handling ..... $0.00"
- `satire.order_form.total` — tone: diegetic
  Text: "TOTAL ................... $0.00"
- `satire.order_form.place_order` — tone: diegetic — Text: "PLACE ORDER"
- `satire.order_form.confirmation` — params: founder — tone: diegetic
  Text: "Thank you, {founder}!! Your $0.00 has been processed. Nothing will ship. You already had everything."
  Why: the thank-you-by-name is the shareware register's warmest move; "you already had
  everything" is the entire economic thesis of the game, delivered as a receipt.
- `satire.order_form.small_print` — tone: diegetic
  Text: "There is nothing to buy. There will never be anything to buy. This form is the whole store."
  (Verbatim the dossier beat's curtain — §6a candidate 1; the honesty small print restated
  diegetically.)

### UNREGISTERED count-up nag (2 strings)

- `satire.unregistered.titlebar_frame` — params: day — tone: diegetic
  Text: "UNREGISTERED — evaluation day {day} of 40"
  ({day} is founder-persistent and counts up forever; the 40 is the era's generic eternal-trial
  shape, deliberately unbranded per the dossier's legal matrix.)
- `satire.unregistered.tooltip` — tone: diegetic
  Text: "The trial never ends and registers nothing. In 1995 this screen asked for $25 and a stamp. We ask for neither. We just liked the screen."
  Why: curtain at full disclosure; "$25 and a stamp" is the dossier's own period-generic figure
  (mail-a-check culture, §3), a texture detail, not a statistic.

### README.TXT (2 strings)

- `satire.readme.window_title` — tone: diegetic — Text: "README.TXT"
- `satire.readme.body` — tone: diegetic — (long-form slot; 78-column ASCII; the codex skin
  and the honesty appendix's T0 rendition)

  Text:
  ```
  ==========================================================================
    CLOUD CLICKER 1.0                                            README.TXT
  ==========================================================================

    Thank you for downloading this program!! It means a lot. Really.

    WHAT THIS IS
    ------------
    A small business simulator. You fix computers for your neighbors,
    one dollar at a time, and you see where that goes. (It goes
    somewhere. The title bar knows more than it is saying.)

    HOW TO PLAY
    -----------
    Click FIX COMPUTER. Buy equipment when you can afford it. Your
    equipment keeps working while you are away -- at 90%, for up to 24
    hours, the way the door games on the BBS shared one phone line.
    They called that fairness. So do we.

    REGISTRATION
    ------------
    This program is free. Not free-to-play. Not free-with-purchases.
    Free. There is no registered version, because there is nothing to
    register against. If you would like the full experience of paying,
    print ORDER.FRM and mail us $0.00. Postage neither included nor
    required.

    WISHLIST (things we are saving up for)
    --------------------------------------
    * an office with a door
    * a second phone line
    * whatever comes after all this  (see: OUR VISION, slide 1 of 1)

    KNOWN ISSUES
    ------------
    * The timer in the title bar counts something that has not been
      explained to you yet. This is intentional. It will be.
    * The visitor counter is real. We cannot fix that. We would not.

    GREETZ
    ------
    To every sysop, every beta tester, everyone who mailed a bug
    report on paper, and whoever invented the perforated edge.

                                              -- the authors, 1995
  ==========================================================================
  ```
  Why: the dossier's README voice spec (§3: plain ASCII, personal thank-you, wishlist, greetz,
  a read inbox) carrying the honesty appendix in-era; the KNOWN ISSUES bullet seeds the Vision
  Slide's payoff without breaking the "1995 surface never knows" rule — the program admits a
  mystery, it does not explain 2026.

---

## Summary

### Counts per family (128 strings total)

| Family | Strings |
|---|---|
| Vision Slide | 12 |
| Navigation / chrome | 15 |
| Desk (incl. 2 cap reasons) | 10 |
| Generators (titles + descriptions + provisioned-cap reason) | 19 |
| Upgrades (titles + descriptions) | 20 |
| Manual action + Horse Armor | 5 |
| Settings / system | 11 |
| Offer sheet | 7 |
| Run-end (incl. 3 curriculum slots) | 9 |
| Event copy (closed 7-kind set) | 7 |
| T0 satire elements (order form 9, nag 2, README 2) | 13 |

No real-world statistic appears anywhere in the set. The three real-world *claims* — the 2006
Horse Armor anchor, "$25 and a stamp," and the shareware-era texture in README.TXT — are
period-generic, unbranded, and grounded in design/01 T1 and era-1995-satire §3/§6; none is a
figure a Sources screen would need to defend.

### Gap list (mechanical IDs / referenced keys with no copy slot, and closed-set collisions)

1. **`resource.company_permits.cap.phase0`** — referenced by economy-v4's permits hardcap;
   absent from the Copy catalog. Authored here (Desk family).
2. **`resource.company_cash.cap.phase0`** — exists in phase0.json with placeholder-grade text
   ("Company cash is capped."). Replacement proposed here; whether a landed catalog row gets
   re-authored is an owner call.
3. **`exit_offer_resolved`** (GU-C11's ruled additive accepted-offer event) — NOT in
   event-copy-v1's closed seven-kind set (`unknown_event: reject`). The accepted-offer feed
   line has no legal binding; the candidate needs a revision row before that event can render.
4. **Gate/split display names** — the splits panel and `event.gate_crossed` have only
   `gate_id`; no `gate.*.title` key family exists anywhere (presentation-v1 binds generators/
   upgrades/manual/cosmetic only). Proposal (unbound, needs a manifest slot):
   `gate.t0_to_t1.title` = "Garage" (design/01: T1's split name; crossing t0→t1 completes the
   Sole Proprietor split into Garage). Flag, not invented into the manifest.
5. **Exit-type display names** — `run_ended` / `exit_offer_spawned` carry `exit_type` with no
   display-name key family. `scripted_first` is covered by its curriculum title at screen
   level, but the event lines and `screen.run_end.exit_frame` will render raw IDs. Needs a
   `exit_type.*.title` family in the manifest round.
6. **Ladder rungs** (purchased_at 25/50/100, 2x multipliers, per generator) — no copy slots
   and no ruled surface (milestone toast? tooltip line?). Not invented; flagged.
7. **Company display name** — the run-title bar's `COMPANY` fragment has no naming flow in
   Phase A. Fallback string authored (`chrome.run_title.company_fallback`); the naming flow
   itself is out of copy's scope.

### Flags (slots I could not fully ground — authored as candidates, not policy)

1. **"Elbow Grease"** (manual meter name): no design source names the token bucket's
   in-fiction energy. The name is my candidate; the tooltip's curtain content IS grounded
   (hardcap law + the anti-energy-system stance). Owner may rename freely.
2. **Vision Slide slide text**: design/11 §1b specifies intent (Tier-7 dashboard as pitch
   deck) but no literal text; the three slide strings are authorial.
3. **Event-copy mechanical-ID leak**: the registered params ARE mechanical IDs; if the
   renderer does not substitute presentation titles, the feed violates the
   mechanical-names-stay-in-code law. Needs an explicit renderer contract, not copy.
4. **`legal_dept` era**: tier 3 in economy-v4, but Phase A's era mapping is closed at
   {0: era_1995, 1: era_2000} and later tiers fail closed. Copy authored in corporate
   register; whether the row can render in Phase A at all is a content-composition question.
5. **Offline fallback merge semantics**: `screen.vision_slide.offline_fallback` deliberately
   makes no claim about reconciling local play with the server later — no spec found.
6. **Tone vocabulary**: existing catalog tones are `corporate`/`diegetic` only. The T0
   webmaster register is tagged `diegetic` here; if the copy linter is to enforce the era
   boundary (banned-word list, `!!` legality), a dedicated tone or era tag may be wanted.
7. **Over-length strings**: `system.offline_progress.tooltip` and `satire.readme.body` exceed
   140 chars by design (tooltip/long-form slots); everything else is within budget.

### The five strings I'd defend hardest

1. `satire.order_form.confirmation` — "Thank you, {founder}!! Your $0.00 has been processed.
   Nothing will ship. You already had everything."
2. `curriculum.scripted_first_failure.next_run` — "The second company gets founded by someone
   who has already failed once. Historically, that is the good kind of founder."
3. `generator.crt_wall.description` — "Six monitors on one desk. Nothing requires six
   monitors. Investors require you to have six monitors."
4. `system.resync.body` — "The two copies of the books disagreed, so everything was
   recounted. The server's copy wins. It always does. Nothing was lost."
5. `desk.manual.meter_tooltip` — "You have two hands. The meter refills on its own, free,
   always. It is a speed limit, not a store."
