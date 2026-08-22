# Mobile Browser / PWA Reality, 2026 — the phone half of our players

> **Dated research snapshot, not current repository authority.** External observations were
> compiled in August 2026; repository observations were frozen at commit
> `ad06e03b7e84079e9d66c7a8385d3573c8692c8a` (2026-08-08). Later fixes and present work are owned
> by the current docs, RFC index, backlog and platform-alignment queues. Every settings line,
> notification/install card, ticker string, mechanic, threshold, budget and routing proposal in
> this file is a non-adopted research example unless a current design/RFC explicitly says
> otherwise. `[M]`, `[P]`, vendor figures and §10's verify-before-shipping limits remain in force.
>
> **Feeds:** `design/06-tech.md` (BINDING stack doc — which, as of this writing, contains **zero
> occurrences of "mobile", "touch", "viewport", or "PWA"**; the desktop shape of our tech research
> is structural, not incidental) · `docs/client-shell.md` (lifecycle + prediction Worker) ·
> `docs/transport.md` + `balance/transport/phase0.json` (reconnect/recovery tuning) ·
> `rfc/ui-foundation.md` (token matrix — no breakpoints or touch tokens exist yet) ·
> `rfc/archive/game-ui-screens.md` · the notification policy (currently one line in `docs/client-shell.md`:
> "Other notifications remain badges owned by later feature surfaces").
>
> **Builds on, does not redo:** `research/browser-rendering.md §5` (mobile GPU + PWA one-pager —
> this file is the deep pass it deferred) · `research/compliance-2026-refresh.md §4` (web push
> legality — **its ruling stands and this file routes around it, not over it**: push is
> banner-compatible but imports Google/Mozilla/Apple as recipients + a US transfer; "ship service
> push or nothing; refuse retention push on design law 1's logic") · `research/tech-stack.md`
> (desktop-shaped; this file is the mobile correction layer).
>
> **Provenance** (house convention, `research/README.md`): `[V]` verified against a fetched URL ·
> `[P]` plausible/secondary source · `[M]` model knowledge, unverified · `[sim]`/`[repo]` my own
> reasoning or a fact checked against this repository. Compiled 2026-08 from three web-research
> passes; each major claim carries its glyph and, where `[V]`, its URL lives in §11. No legal
> matrix in this file — the legal surface (push, storage) is owned by `compliance.md` +
> `compliance-2026-refresh.md`; nothing here introduces a new legal exposure beyond what those
> files already ruled on.

---

## 0. The one-paragraph answer

Mobile browsers in 2026 are hostile to exactly one thing we do — **pretending to be alive while
not on screen** — and indifferent-to-friendly toward everything else we do. Our two founding
architecture laws turn out to be the correct mobile survival strategy, chosen before we knew it:
**server-authoritative lazy closed-form progress** means iOS suspending the entire WebContent
process at screen-lock costs the player *nothing* (the return is just an offline-progress claim
with a smaller gap), and **DOM-first rendering with numbers throttled to 10 Hz** is within
low-end-Android budgets *if* we contain layout (fixed-width tabular numerals, `contain`, no
per-frame layout-triggering styles). The real mobile work is therefore not progress, not
rendering, but **four seams**: (1) the *live-UI/reconnect* seam — the socket dies silently on
every screen lock and network flip, so "hidden ⇒ disconnect deliberately; visible ⇒ reconnect +
recover-or-resync" must become policy, not accident (Centrifugal's own docs mandate this `[V]`);
(2) the *installability* seam — iOS 26 made every Add-to-Home-Screen open standalone with **zero
installability requirements** `[V]`, which converts "install our PWA" from an evangelism problem
into a one-card honesty problem, and installing is also what exempts a player from Safari's 7-day
storage wipe `[V]`; (3) the *input* seam — the ~20–25/s click-clamp world was calibrated for
mouse; sustained one-finger touch tapping is materially slower, so the clamp is unreachable
headroom on phones rather than a limit, and the button needs multi-touch + `pointerdown` +
long-press/zoom suppression; (4) the *notification* seam — push works on installed iOS PWAs since
16.4 `[V]` and opt-in reality is single-digit-to-low-teens percent `[V]`, which makes the honest
design cheap: we were never going to build retention pings (design law 1 + the compliance ruling
already refused them), so the curtain-pulled stance — "this game will not notify you; here are the
one or two state-truth exceptions you can opt into, with their trigger formulas printed" — costs
us nothing and is the satire. Notably, **no successful web incremental treats mobile web as
first-class** (§6): they either ignore it or funnel to native wrappers. Doing mobile web *well* is
an open niche, and our architecture is unusually pre-adapted to it.

---

## 1. Background execution: what actually keeps running (topic 1)

### 1.1 Main-thread timer throttling, by browser

| Browser | Hidden-tab timer policy | Prov |
|---|---|---|
| Chrome (desktop + Android) | Three tiers: **minimal** (visible, or audible in last 30 s) → normal timers; **1 s alignment** when hidden < 5 min (or nesting < 5, or live WebRTC); **intensive: once per MINUTE, batched** when hidden > 5 min ∧ nesting ≥ 5 ∧ silent ≥ 30 s ∧ no WebRTC. Shipped Chrome 88 (2021). | `[V]` Chrome timer-throttling doc |
| Chrome, budget tier | Since Chrome 57: after 10 s hidden, timers run only while a per-tab CPU budget (regen 0.01 s/s) is positive. | `[V]` |
| Firefox desktop | 1 s minimum timeout in inactive tabs; exempt if the tab has an `AudioContext`. | `[V]` MDN |
| **Firefox for Android** | **15-minute** minimum timeout for inactive tabs, "and may unload them entirely." | `[V]` MDN |
| Safari | ~1 s hidden-tab DOM-timer alignment with auto-increasing interval (ceiling ~30 s); mostly moot on iOS because the whole process suspends (§1.3). | `[M]` |

**The WebSocket exemption is GONE.** Chrome 57's budget throttling explicitly exempted pages with
open WebSockets `[V]`; the Chrome 88 intensive-throttling regime exempts **only WebRTC and
audio** — the flag `AllowAggressiveThrottlingWithWebSocket` shipped enabled (~Chrome 107, 2022),
meaning an open socket no longer protects main-thread timers `[V doc / M version]`. **Do not rely
on "we hold a socket" to keep anything ticking in a hidden tab.**

### 1.2 Dedicated Workers — our sim loop's actual contract

- No spec'd or MDN-documented worker-specific timer throttling; Chrome's 1 s/1 min clamps are
  implemented in the main-thread scheduler. A 50 ms loop in a dedicated Worker keeps firing in a
  merely-hidden tab on Chrome (desktop and Android, while the renderer lives) `[P]` — this is the
  ecosystem's standard workaround, but no primary sentence states it; **verify on device** (§10).
- **Throttling is not the only mechanism.** Page *freezing* (Lifecycle) and *discarding* stop the
  entire frame — Workers included — and discarding kills the renderer outright `[V/M]`.
- **iOS: Workers buy nothing.** Screen lock or app switch suspends the whole WebContent process —
  main thread, Workers, timers, rAF — within seconds-to-tens-of-seconds (no public number) `[M]`.
  Centrifugal's protocol doc says it plainly: *"a mobile OS can kill the connection at some point
  without any callbacks called"* `[V]`.
- **Practical contract for the 20 Hz Worker: it is guaranteed only while
  `document.visibilityState === "visible"`.** Everything else is best-effort.

**Routing — client shell (`docs/client-shell.md`), mostly already correct `[repo]`:** the shipped
shell already (a) anchors prediction to authoritative amounts and evaluates *total elapsed ms*
through `accrueConstant` rather than counting ticks, so a starved Worker self-corrects; (b) treats
a >5,000 ms clock gap as "no local catch-up, request the authoritative offline path"; (c) measures
the hidden interval independently of Worker throttling so a still-ticking Worker can't bypass the
return-story rule; (d) uses `visibilitychange`/`pagehide`/`freeze` and has **no `unload` handler**.
This is exactly the architecture the platform data demands. The mobile deltas it still needs are
in §8.

### 1.3 Freezing, discarding, lifecycle events on mobile

- `freeze`/`resume`: Chromium-only (Chrome 68+); not implemented by Safari or Firefox `[V/M]`.
- **Discard has no event** — detection is only `document.wasDiscarded === true` on the reload
  `[V]`. On low-RAM Android, assume a backgrounded tab can be gone at any moment `[M]`.
- `unload`: never fired in typical mobile situations (tab-switcher close, app-switcher kill);
  mobile browsers bfcache pages despite unload listeners `[V]`. We correctly have none `[repo]`.
- **The complete dependable last-gasp set on iOS is `visibilitychange→hidden` and `pagehide`.**
  `freeze` is a Chromium bonus `[V]`.
- bfcache 2025-26: all majors; per current web.dev text, Safari and Chrome (~149+ `[P]`) no longer
  block bfcache on open WebSockets — the page is paused; you close the socket yourself in
  `pagehide`/`freeze` `[V/P]`. A bfcache restore fires `pageshow` with `persisted: true` — that is
  a reconnect trigger we do not currently consume `[repo]`.

### 1.4 WebSocket lifetime, by scenario

| Scenario | What happens to the socket | Prov |
|---|---|---|
| Desktop, tab hidden | Survives until freeze/discard; incoming messages still delivered; your scheduled *handlers* are throttled, message events are not | `[M]` |
| iOS, screen lock / app switch | Process suspended; TCP dies while JS is frozen; **no close event at death**; on resume the socket object often still reads OPEN (half-open) and fails only on next write or ping deadline | `[M]`, corroborated `[V]` by Centrifugal's no-callbacks warning |
| Android, screen off + Doze | Network access deferred to maintenance windows; renderer gets no pass; long idle silently starves/kills the connection; OEM battery managers (Samsung/Xiaomi/Huawei/OnePlus, all max severity on dontkillmyapp) are worse | `[M]` + `[V]` dontkillmyapp |
| Carrier NAT | Idle TCP mappings expire in ~5–30 min (classic measurement: Wang et al., SIGCOMM 2011); expiry is silent — no RST reaches the client | `[M]` |

### 1.5 What actually still runs on mobile web, 2026

Audible media keeps a page foreground-class on Chrome `[V]`; service workers are event-driven
only (woken for push/fetch, killed after ~30 s idle) `[M]`; web push works (§3); Background/
Periodic Sync is Chromium-only and rare-interval `[M]`. **Nothing else.** No timers, Workers, or
sockets survive iOS backgrounding. Every return-to-foreground is architecturally identical to an
offline-progress claim — which is the system we already shipped. **Mobile makes our lazy
closed-form law (design law 2/7) load-bearing rather than elegant.**

---

## 2. PWA installability, 2026 (topic 2)

### 2.1 iOS — the surprise: installability requirements are gone

- **iOS 26 (Sept 2025): "By default, every website added to the Home Screen opens as a web
  app"** — no manifest required; the user gets an "Open as Web App" toggle to opt out; Apple:
  "there are now zero requirements for 'installability' in Safari" `[V]` WebKit Safari-26.0 post.
  Largest install-UX shift since 16.4. Install friction is now purely *discovery* (Share sheet),
  not qualification.
- Capabilities of an installed iOS web app, current through iOS 26: standalone display (11.3+),
  service workers (11.3+), **web push (16.4+, installed-only, permission request must be inside a
  user gesture, APNs transport)** `[V]`, Badging API (16.4+) `[V]`, Screen Wake Lock (16.4+)
  `[V]`, third-party (WebKit-skin) browsers may offer Add to Home Screen (16.4+) `[V]`.
- **Storage: installing is the durability upgrade.** Safari's ITP deletes *all* script-writable
  storage after 7 days of Safari use without interacting with the site — but Home-Screen web apps
  keep their own counter and Apple states first-party data deletion there is not expected
  (bug-reportable) `[V]` WebKit third-party-cookie-blocking post. In-tab Safari, a lapsed player
  can lose local state; installed, effectively durable. Our server-authoritative saves make this
  cosmetic for progress `[repo]` — but it can silently wipe an anonymous/local session identity
  or cached catalogs; treat all client storage as cache (§8).
- Still true on iOS: **no `beforeinstallprompt`, no programmatic prompt** (non-standard,
  Chromium-only `[V]` MDN); no element fullscreen on iPhone `[M]`; **no `ScreenOrientation.lock()`**
  (limited availability; design for both orientations) `[V/M]`.
- **DMA final state:** after the Feb 2024 kill-and-revert drama, Apple's compliance page commits
  to Home-Screen web apps in the EU, **built on WebKit — alt engines are authorized only for
  browser apps, explicitly not for Home-Screen web apps** `[V]` Apple DMA page. No Gecko/Blink
  browser has shipped on iOS as of mid-2026 (a Blink prototype exists, 28.6% faster than Safari
  on Speedometer 3.1, publicized by OWA June 2026) `[V]` OWA blog. **Plan for WebKit as the only
  iOS engine for the game's lifetime.**

### 2.2 Android Chrome

- Installability criteria (current web.dev text): manifest (`name`/`short_name`, 192+512 px
  icons, `start_url`, `display` standalone-class), HTTPS, not installed, plus an engagement
  heuristic (≥1 interaction, ≥30 s). **Service worker/offline is no longer a criterion** `[V]`.
- Richer install bottom-sheet (Chrome 94+) requires ≥1 manifest `screenshots` entry +
  `description` `[V]`. WebAPK minting gives a real launcher entry on Play-Services devices `[M]`;
  no Play Services (post-2019 Huawei) ⇒ no WebAPK and no FCM push `[M]`.
- In-app browsers (Instagram/TikTok/Gmail webviews, CCT) don't fire `beforeinstallprompt`;
  players arriving from social links can't install until they escape to a real browser `[M]`.

### 2.3 Install-prompt norms — and our honest version

Canonical: capture `beforeinstallprompt`, `preventDefault()`, show a **contextual** install
button at a value moment, `prompt()` on tap `[M]` (web.dev patterns). Anti-patterns: prompting on
load, install interstitials, re-prompting after dismissal, "install to continue" `[M]`.

**Curtain-pulled install card (ours):** one dismissible card, shown at a value moment (e.g. first
Exit or first hardcap hit), that states exactly what install changes and what it doesn't: *"real
notifications (only if you opt in, §3), an icon, durable saves on iPhone — nothing else; the game
is identical and free either way."* On iOS the card is instructions (Share → Add to Home Screen);
on Android it triggers the deferred native prompt. Honest AND per the double-permission
literature the higher-converting shape `[M inference from V'd web.dev guidance]`.

---

## 3. Notifications: support, opt-in reality, and the honest design (topic 3)

### 3.1 Support matrix, mid-2026

| Platform | State | Prov |
|---|---|---|
| iOS/iPadOS 16.4→26 | Installed Home-Screen web apps ONLY; user-gesture-gated permission; APNs. Unchanged through iOS 26 (verified by absence in Safari 26.0/27-beta posts) | `[V]` |
| **Declarative Web Push** | iOS/iPadOS 18.4 + macOS 15.5 (Mar 2025): JSON payload displayed without waking a service worker; subscription without SW; classic push's punish-silent-push revocation avoided. WebKit-only so far; adopt additively | `[V]` |
| Android Chrome | Classic push via FCM, works in-tab (no install needed); Android 13+ also needs Chrome's app-level notification permission | `[M]` |
| Chrome hygiene | Quieter UI (bell icon) for habitual blockers/disruptive sites; Chrome **may auto-revoke** notification permission for disruptive or long-unvisited sites | `[V]` Chrome support |
| Android delivery | FCM subject to Doze batching; OEM battery managers break delivery (Huawei/Xiaomi/OnePlus/Samsung max-severity per dontkillmyapp); Huawei sans Play Services = no web push at all | `[V]` dontkillmyapp + `[M]` |

**Engineering consequence: a granted permission is not permanent and a subscription is not
reliable.** Detect endpoint death server-side (410/404 on send) and reflect it honestly in-game;
never treat a dead subscription as a player action; never make a notification load-bearing —
which design law 7 (offline progress default) already guarantees `[repo]`.

### 3.2 Opt-in and fatigue numbers

- Web push opt-in benchmarks (PushEngage, updated 2026-03): up to ~20% for single-step native
  prompts in the best segments; ~7–16% by region/vertical `[V]`; other sources ~6% average `[P]`.
  App-push contrast: Android ~81%, iOS ~51% medians `[P]`. **Planning number for a game website:
  single-digit to low-teens — and on iOS only within the installed minority.** Push is a bonus
  channel for a small self-selected slice, never a retention pillar.
- web.dev permission-UX (updated 2025-03): never prompt on load; a native-prompt denial is
  near-permanent; the soft prompt's chief value is *protecting* the native prompt `[V]`.
- NN/g: 56 notifications/day average (Telefonica 2016); guidelines — delay until value shown,
  state exactly what notifications contain, consolidate, opt-out **inside the app** `[V]`.
- Idle-game notification culture `[M — unverified, do not quote as fact]`: capacity pings
  ("storage full"), lapse-guilt pings ("your capitalists miss you!"), FOMO/streak pings. The
  defensible kernel in vendor retention claims is that *state-truth* notifications outperform
  generic "come back" blasts (segmented 1.3–1.4× CTR is the verified datum `[V]`). Guilt pings
  are a recurring uninstall-story trope. Honest-stance exemplars: Egg Inc. (per-type toggles,
  concrete-state triggers); the browser canon (Cookie Clicker, Universal Paperclips, Melvor web)
  sends nothing `[M]`.

### 3.3 Our notification policy — the curtain-pulled design

`compliance-2026-refresh.md §4` already ruled: push is compatible with the no-banner rule, but it
adds Google/Mozilla/Apple as data recipients + a US transfer to a deliberately EU-clean
architecture, **and a re-engagement push is ePrivacy direct marketing and the exact dark pattern
the game mocks — "ship service push or nothing."** This file's research strengthens that ruling
(low opt-in ceilings make the third-party cost buy little) and specifies what "service push," if
ever shipped, looks like:

1. **The default is the satire.** The settings panel says: *"This game will not notify you. Idle
   games use notifications to manufacture guilt; ours literally cannot — your company runs at 90%
   while you're gone either way (formula: `docs/production-engine.md`)."* That line is design
   laws 1, 7, 9, and 10 in one surface.
2. **At most one or two narrow, named opt-ins**, each a separate toggle stating its exact
   state-truth trigger and hard frequency cap as visible numbers (law 5), e.g. *"the Commons
   event I explicitly joined is starting (max 1 per event)"*. No "enable notifications" blanket.
3. **Soft prompt first; native prompt only on an explicit tap** (iOS mandates the gesture anyway
   `[V]`); a decline is stored and never re-asked — which also keeps us out of Chrome's
   disruptive-site auto-revocation regime `[V]`.
4. **Mute as prominent as enable**, in-app (NN/g `[V]`).
5. **Each opt-in's tooltip pulls the curtain** (law 10): what the ping is engineered to do, and
   its trigger formula (law 9's Helldivers-transparency applied to attention).
6. If shipped at all: Declarative Web Push on Apple platforms, classic elsewhere `[V]`; and the
   DPIA/recipients documentation cost from `compliance-2026-refresh.md §4.5` is paid first.
   **Default remains: don't ship it.** Nothing in this dossier argues the third-party import is
   worth it for a single-digit-percent channel; it argues that *if* Marco ever wants the one
   honest exception, this is its shape.

---

## 4. Mobile performance for a DOM-heavy incremental (topic 4)

### 4.1 The 10 Hz number problem — layout, not paint

- Layout is "almost always scoped to the entire document" and cost scales with DOM size (web.dev's
  example trace: 1,618 elements checked, 28 ms in layout — ~2× a 60 fps frame budget) `[V]`.
  Changing a text node's content is a **layout** operation, not just paint — text is shaped during
  layout, and a width change propagates into flex/grid redistribution `[V mechanism / M
  propagation]`. An *uncontained* counter in a flex row re-lays-out its panel 10×/s.
- **The fix is a contained, fixed-width number cell:** `contain: layout paint` (or `strict` with
  explicit size) confines recalculation to the cell's subtree `[V]` MDN (Baseline since 2022) +
  `font-variant-numeric: tabular-nums` (equal-width figures, Baseline since 2020 `[V]`) + a
  `ch`-based fixed width so the cell never resizes. With that recipe, a dozen 10 Hz counters are
  comfortably inside budget even at 4–6× CPU slowdown `[M]`.
- **Routing:** the shell already throttles formatting and Svelte refresh to 100 ms `[repo]` — the
  cadence is right; what does not exist yet is the *containment/tabular-nums counter-cell spec*,
  which belongs in the UI Foundation token/component contract (§9).

### 4.2 Layers, `will-change`, filters

- Only `transform` and `opacity` animate compositor-only `[V]`. web.dev's comparative trace:
  animating `top/left` = 37 ms render + 79 ms paint, 50% dropped frames; the same via `transform`
  ≈ 0 ms, 1% dropped `[V]`.
- Every promoted layer is an uncompressed texture (~w×h×4×DPR² — a full-screen layer on a
  1080×2400 phone ≈ 10 MB `[M]`); MDN: excessive `will-change` = "excessive memory use…worse
  performance"; last resort, toggled around the animation, smallest possible subtree `[V]`.
  This confirms `browser-rendering.md §2`'s compositor-layer-explosion warning with numbers.
- `backdrop-filter` reached cross-engine Baseline only Sept 2024 `[V]`; on tile-based GPUs an
  animated backdrop blur forces per-frame read-back + multi-pass blur — among the most expensive
  effects on low-end Android `[M]`; animated large-radius blur/shadow is the same family `[V/M]`.
  **Rule: no animated blur/backdrop-filter/box-shadow on the mobile path; pre-baked shadow
  sprites + opacity crossfades.** Static shadows/gradients are fine (painted once, cached) `[M]`.

### 4.3 CSS sprite animation done compositor-only

Animating `background-position` (the naive `steps()` sprite) triggers **paint** per step per
sprite `[M mechanism / V paint-vs-transform split]`. The compositor-only pattern: sprite strip in
a fixed-size `overflow:hidden`/`contain: paint` window, animated with `transform: translateX()` +
`steps(N)` — one texture, transform snaps. Budget guidance (no rigorous published dataset exists
`[V absence]`): dozens of transform/opacity-only animations are fine on low-end; more than a
handful of *paint-triggering* ones drop a Moto-G-class phone below 60 fps; pets ≤10–15 visible
CSS-animated sprites, `animation-play-state: paused` off-screen `[M]`. This tightens
`browser-rendering.md §2`'s desktop cap (~60 visible) to a **mobile cap roughly a quarter of it**
and specifies the required animation technique.

### 4.4 Battery

- **Radio:** LTE RRC idle <15 mW vs connected 1,000–3,500 mW, plus the post-transfer energy tail;
  HPBN's Pandora case: a 60 s analytics beacon = 0.2% of bytes but **46% of total power** `[V]`.
  A ~20–25 s keepalive on cellular keeps the radio near-permanently in its high-power duty cycle
  `[M inference]` — acceptable only while the screen (the dominant drain) is on anyway; **never
  hold a socket for a hidden page on mobile** (§7).
- **Wake Lock:** Baseline Mar 2025; auto-released on hide/low battery `[V]`. Opt-in setting only,
  re-acquired on visibility — as `browser-rendering.md §5` already ruled `[repo]`.
- rAF discipline: render work at 10 Hz, all animation compositor-side, everything paused on
  `hidden` `[M]` — our fixed-timestep-in-Worker + 100 ms presentation cadence already conforms
  `[repo]`.

### 4.5 The low-end baseline, 2026

- Android Go RAM floor: 2 GB since Android 13 Go `[V]`. Bottom-quartile reference: **Samsung
  Galaxy A51 class** — ~4.25× slower single-core than an iPhone 15 Pro; ~30% of worldwide volume
  is still low-end; derived budget **≤365 KiB compressed JS for a 3 s first load** (650 KiB for
  5 s) `[V]` infrequently.org 2024 (2025/26 update unfetched `[M]`).
- Lighthouse mobile emulation = 4× CPU slowdown, 150 ms RTT / 1.6 Mbps `[V]`; true bottom
  quartile is worse than the default.
- **INP is the game-feel metric for a clicker**: good ≤200 ms at p75, same threshold mobile and
  desktop `[V]`. After 4–6× throttling, per-click JS+render should target **<30–40 ms of
  desktop-measured main-thread time** `[M derivation]`. This is a measurable acceptance gate for
  the Game UI Screens RFC (§9).

---

## 5. Touch UX for a clicker (topic 5)

### 5.1 Tap targets — exact standards

Apple HIG **44×44 pt** `[V]` · Material/Android **48×48 dp** (~9 mm) `[V]` · web.dev ~48 px with
8 px spacing `[V]` · **WCAG 2.2 SC 2.5.8 (AA): ≥24×24 CSS px** with spacing/equivalent/inline
exceptions; SC 2.5.5 (AAA) is 44 px `[V]`. Reading: 24 px is the legal floor, 48 dp the working
minimum for buy rows, and the core click button should be thumb-zone scale (80–120+ px) `[M]`.
Our a11y posture already targets WCAG via axe-core in UI Foundation `[repo]` — target size should
join the token contract as named size tokens (§9).

### 5.2 The 300 ms delay is dead; the remaining latency is ours

Chrome 32 (2014) removed it for `width=device-width` pages; iOS followed March 2016;
`touch-action: manipulation` removes it per-element at any zoom, now supported in Safari `[V]`
Chrome blog + WebKit blog. **Non-issue in 2026 given a proper viewport meta** — ship
`touch-action: manipulation` on interactive elements as belt-and-suspenders; do NOT use
`touch-action: none` page-wide (blocks zoom, harms low-vision users — scope to the click button)
`[V]`. Remaining latency: digitizer scan (~8–16 ms), event→rAF alignment, and main-thread
busyness (= INP). **Register the click on `pointerdown`, not `click`** — fires at contact, not
lift, cutting ~50–100 ms perceived latency; suppress the compatibility click `[M]`.

### 5.3 Touch tap rates vs the click clamp — a fairness finding

All `[M]` (no citable measurement survived verification — playtest before tuning, §10): sustained
single-finger ~5–7 taps/s, bursts ~8–10/s; practiced two-thumb alternation ~10–14/s; the 15–25+
CPS mouse techniques (butterfly/drag-click) **have no touchscreen equivalent**.

**Consequence:** a ~20–25/s manual clamp calibrated for mouse is unreachable headroom for one
honest finger and barely reachable two-thumbed — if manual clicking matters economically, touch
players are structurally disadvantaged. Options: (a) count multi-touch — N simultaneous
`pointerdown`s = N clicks (3–4 drumming fingers ≈ 15–20/s); (b) set the clamp low enough
(~10–12/s) that both input classes saturate it, shifting skill off raw CPS. **The clamp is
balance data, not code `[repo]`, so this is a tuning-plus-ruling question, not a law change** —
but multi-touch-counting needs an explicit design ruling (it changes what "a click" means and
what bots may emulate). Flag at the click-clamp definition site as a `DESIGN-GAP:` when the
active-play surface is next touched (§9).

- Autoclicker note: Android Accessibility-API clickers generate OS-trusted events, so
  `event.isTrusted` is insufficient; touch-side signals are periodicity/coordinate-variance/
  pressure-field heuristics — advisory only. **The server-side rate clamp remains the real
  defense, which we already have `[repo]`.** `[M]`

### 5.4 Mobile interaction hazards checklist

| Hazard | Fix | Prov |
|---|---|---|
| Double-tap zoom on the button | `touch-action: manipulation` (+ `width=device-width`) | `[V]` |
| Long-press selection/callout | `user-select: none` on game chrome + iOS `-webkit-touch-callout: none`; `contextmenu` preventDefault as backup | `[M]` |
| Pull-to-refresh / scroll chaining | `overscroll-behavior: none` on `html`, `-y: contain` on internal scrollers; Safari iOS 16+ (~95% global) — progressive hardening, not load-bearing | `[V]` MDN + caniuse |
| iOS edge-swipe back | **Cannot be disabled by web content**; keep interactive elements out of the outer ~20–30 px gutters | `[M]` |
| 100vh under the URL bar | `height: 100svh` for the shell (stable), `dvh` only where live tracking wanted (throttled updates — never bind animation to it); Chrome 108+/FF 101+/Safari 15.4+ | `[V]` web.dev |
| Keyboard occlusion (chat/name inputs) | `visualViewport` resize/scroll; Chrome `interactive-widget=resizes-content` | `[M]` |
| Haptics | `navigator.vibrate()` is **Android-only; iOS Safari has never supported it through 26.x** `[V]` caniuse — Android-only enhancement, ≤10–20 ms pulses, behind a user setting | `[V/M]` |
| Safe areas | `viewport-fit=cover` + `env(safe-area-inset-*)` — already ruled in `browser-rendering.md §5` | `[repo]` |

---

## 6. How existing browser incrementals handle mobile (topic 6)

- **Cookie Clicker:** official **Android** app (beta Aug 2019, release Oct 2020); **no iOS
  version exists** (verified two ways: Wikipedia platform list + App Store search returning only
  clones) `[V]`. The mobile *web* experience is unoptimized desktop web funneling to the Android
  app, which historically lags web content `[M]`. The genre flagship treats mobile as an app
  funnel, and its mobile app is its weakest platform `[M]`.
- **Melvor Idle:** browser-origin, Steam 1.0 Nov 2021, published by Jagex, cross-platform cloud
  saves `[V]`; mobile is **store apps** (free, Jagex-published iOS/Android) — the web game in a
  webview wrapper (Cordova-family `[M]`), UI compressed to single-column. The strongest
  "web-first incremental that took mobile seriously" — and it still chose apps `[V/M]`.
- **Kittens Game:** free web, **paid iOS app $2.99** ("no ads or micro-transactions") `[V]`.
  **Universal Paperclips:** free web, official iOS app $1.99, engine rebuilt for mobile in 2.0
  `[V]`. The pattern for free web games: *the app is the monetization*.
- **Antimatter Dimensions:** Vue web app `[V repo]`; the rare big incremental genuinely playable
  in a phone browser via responsive tabs `[M]`; iOS listing exists (July 2026) but officialness
  unconfirmed `[V listing / M officialness]`. **IdleOn:** free same-codebase iOS app `[V]`.
  **NGU Idle:** no iOS app; web + Steam `[V absence]`. **IdleMMO** (Galahad Creative): browser
  idle MMO shipping *both* a mobile-web/PWA experience and store apps — the closest existing
  "idle MMO with a first-class mobile web story" `[V listing / M characterization]`.
- **The pattern:** every commercially successful web incremental funnels mobile to store apps;
  none treats mobile web as the primary mobile surface (AD and IdleMMO nearest exceptions) `[M
  synthesis of V'd listings]`. The genre's stated reasons — save fragility on iOS, no background
  progress, dense desktop UIs — are exactly the two things our architecture already solves
  (server saves, closed-form offline) plus one thing we must still do (layout compression).
  **Mobile web done well is an open niche in this genre, and we are pre-adapted to it.** For a
  game whose design law 1 forbids monetization, the genre's app-funnel playbook (paid app or
  store economics) is unavailable anyway — mobile web + optional PWA install *is* our mobile
  strategy, by elimination and by fit. (PEGI note from `compliance-2026-refresh.md`: a storefront
  PWA would trigger PEGI 18 for the parody casino — one more reason mobile web, not stores.)

---

## 7. Transport: mobile reconnect policy (topic 7)

### 7.1 The policy the platform data dictates

1. **Hidden on mobile ⇒ disconnect deliberately.** Centrifugal's own protocol doc: *"Disconnect
   from a server when a mobile application goes to the background since a mobile OS can kill the
   connection at some point without any callbacks called"* `[V]`. This also zeroes the keepalive
   radio cost (§4.4) and sidesteps the bfcache-blocking question. Desktop hidden tabs may stay
   connected (mains/wifi, cheap) — but socket + ping handling must live in the Worker, where
   Chrome's main-thread clamps don't reach `[P]`.
2. **Reconnect immediately (skip backoff) on:** `visibilitychange→visible`, `pageshow` with
   `persisted: true` (bfcache restore — an event we don't currently consume `[repo]`), the
   `online` event, and Chromium `resume`. `navigator.onLine` is "inherently unreliable" `[V]` MDN
   — treat `online` as an *attempt now* hint, never gate features on it.
3. **Treat every mobile resume as "socket unknown, probably dead."** iOS resume typically hands
   back a half-open socket that still reads OPEN and fails only on next write or ping deadline
   `[M, V-corroborated]`.

### 7.2 Tuning against our shipped config

- **Centrifuge defaults are already mobile-sane `[V]`:** client `minReconnectDelay` 500 ms /
  `maxReconnectDelay` 20 s, full-jitter backoff (the AWS algorithm — with 100 contending clients,
  full jitter more than halved total call volume `[V]`); server ping 25 s / pong timeout 8 s;
  client `maxServerPingDelay` 10 s ⇒ **worst-case half-open detection ≈ 35 s**. 25 s pings sit
  under every common proxy/NAT idle timeout (ALB/nginx 60 s; Cloudflare ~100 s `[P]`; carrier NAT
  5–30 min `[M]`). Keep them — they only run while visible under rule 7.1(1).
- **Recovery window vs the pocket gap:** player history is 512 msgs / **10 minutes** `[repo]`.
  A phone in a pocket routinely exceeds 10 min, so **the common mobile return takes the
  `recovered: false` → full-resync path — that is by design, not failure**: Centrifugal's docs
  say the application database stays the source of truth and history is the fast shortcut `[V]`,
  and our shell already routes any >30 s absence through the authoritative snapshot + return
  recap `[repo]`. The mobile requirement is only that full resync stays cheap and unremarkable —
  one snapshot + one offline claim. Do not widen history TTL for mobile; it buys nothing.
- **World channel:** recovery returns latest-only and `world_rev` is a per-process ordering key
  treated as a new baseline on reconnect `[repo]` — already the right semantics for
  reconnect-heavy mobile.
- **Duplicate tolerance:** recovery replay and outbox redelivery can both duplicate; events carry
  immutable IDs and receipts carry intent identity + revision `[repo]` — the dedupe-by-ID rule
  must remain a client invariant under reconnect churn.

---

## 8. Gaps observed at the frozen 2026-08-08 coordinate

This table is historical audit evidence, not a current defect list. Use the live backlog and
platform-alignment execution queue for present status.

| # | Surface | Finding | Severity |
|---|---|---|---|
| 1 | `design/06-tech.md` | The BINDING stack doc contains zero mobile provisions (no "mobile", "touch", "viewport", "PWA") `[repo]`. Mobile-half-of-players is currently an undocumented assumption. | HIGH (process) |
| 2 | Client shell lifecycle | Handles `visibilitychange`/`pagehide`/`freeze`, no `unload` — correct `[repo]`. **Missing:** `pageshow(persisted)` and `online` as reconnect/resync triggers; no hidden⇒disconnect policy (transport RFC owns it). | HIGH |
| 3 | Transport | No mobile socket policy: today a screen lock leaves a silent half-open socket; with defaults, up to ~35 s of stale-but-alive-looking UI after resume unless resync-on-visible is policy (§7.1). | HIGH |
| 4 | Sim Worker | 20 Hz loop guaranteed only while visible; wall-clock anchoring + >5 s gap → authoritative path already shipped `[repo]` — conforming. Residual risk: assuming Worker liveness anywhere else. | OK / guard |
| 5 | UI Foundation tokens (`rfc/ui-foundation.md` C9) | Token matrix has no breakpoints, no touch target-size tokens, no safe-area tokens, no `svh/dvh` rule, no tabular-nums/contained counter-cell spec — the entire §4.1/§5 surface is unowned. | HIGH |
| 6 | Click button (Game UI Screens / active-play) | No `pointerdown`/`touch-action: manipulation`/multi-touch spec; ~20–25/s clamp calibrated for mouse — touch can't reach it (§5.3). Needs a design ruling on multi-touch counting or clamp retune (balance data). | MED-HIGH |
| 7 | PWA assets | No manifest/icons/screenshots exist. Android richer install sheet needs manifest + ≥1 screenshot + description `[V]`; iOS 26 needs nothing but the Share sheet `[V]`. Cheap, unowned. | MED |
| 8 | Storage assumptions | Safari 7-day wipe applies to in-tab players `[V]`; anything locally persisted (session identity, cached catalogs) must be reconstructible from the server. Audit at the save/account seam. | MED |
| 9 | Notification policy | Currently one line (badges deferred) `[repo]`. The curtain-pulled settings copy (§3.3) doesn't exist; risk is a later surface shipping a naive prompt. Push itself stays deferred per compliance ruling. | MED (policy) |
| 10 | Rendering rules | Mobile pet cap (~≤15 visible, transform-only sprites, no animated blur) tightens `browser-rendering.md`'s desktop numbers; nothing enforces it yet. | MED |
| 11 | Perf budget | 365 KiB JS / INP ≤200 ms on A51-class / 4× CPU throttle are concrete, measurable, and unmeasured for our bundle. Candidate CI/Lighthouse gate. | MED |
| 12 | CSP | `connect-src 'self'` doesn't cover `wss:` in all browsers — already found by `cicd-deploy.md §9.5`; mobile makes reconnect paths exercise it constantly. | (already filed) |

---

## 9. Historical routing proposals

These routes were research recommendations at the frozen coordinate. They do not make work READY
and do not supersede the active RFC/decision/execution queues.

| Finding | Route to | Action shape |
|---|---|---|
| Mobile execution contract (§1), socket policy (§7) | `design/06-tech.md` mobile addendum + the active transport RFC | Addendum §: "mobile lifecycle contract" — visible-only Worker guarantee, hidden⇒disconnect, resume⇒resync; transport RFC acceptance criterion: reconnect-on-visible with recovered:false full-resync proven in the browser suite |
| `pageshow(persisted)` + `online` triggers (§7.1) | client shell (`docs/client-shell.md` + its successor RFC batch) | Two added lifecycle handlers + WebKit browser-suite case |
| Ping/backoff/recovery numbers (§7.2) | `balance/transport/phase0.json` + Centrifugo config | Keep defaults; record them as reviewed values; do NOT widen history TTL for mobile |
| Counter-cell spec: `contain` + `tabular-nums` + fixed `ch` width (§4.1) | `rfc/ui-foundation.md` C9 token/component contract | A named `counter-cell` primitive; components consume it, never raw text nodes |
| Breakpoints, target-size tokens (≥48 dp buys, thumb-scale main button), safe-area + `svh` tokens (§5) | `rfc/ui-foundation.md` token matrix + `rfc/archive/game-ui-screens.md` | New token domains: `size.touch`, `inset.safe`, breakpoint set; axe-core already gates a11y `[repo]` — add WCAG 2.5.8 check |
| `pointerdown` + `touch-action` + hazard suite (§5.2, §5.4) | `rfc/archive/game-ui-screens.md` (click surface spec) | Interaction spec lines + WebKit/Chromium touch tests |
| Touch vs click-clamp fairness (§5.3) | `design/02` / active-play clamp definition site | `DESIGN-GAP:` flag when next touched — multi-touch counting vs lower clamp needs an owner ruling; constants are balance data |
| Mobile sprite/animation caps (§4.2–4.3) | `browser-rendering.md` consumers / future pet-renderer RFC | Mobile cap ≤10–15 visible, transform/opacity-only, no animated blur — acceptance criterion for the CSS pet renderer |
| Perf budget + INP gate (§4.5) | `rfc/scaffolding-and-ci.md` successors / CI obligations inventory | Lighthouse mobile-emulation job: JS ≤365 KiB, INP budget; joins the CI obligation table in `research/README.md` |
| PWA manifest + install card (§2) | new small client RFC or Game UI Screens stretch; copy via copy pipeline | Manifest + icons + screenshots; one dismissible curtain-pulled install card (value-moment trigger); BACKLOG entry |
| Storage-is-cache audit (§2.1) | save-layer / accounts docs seam | Verify nothing local is unreconstructible; note ITP exemption for installed apps |
| Notification policy §3.3 | notification policy owner (currently `docs/client-shell.md` badge line) + `design/08` voice for the settings copy | Adopt the six-point curtain-pulled policy as the standing default ("no notifications" IS the policy); push remains deferred per `compliance-2026-refresh.md §4.5` |
| Mobile-web-as-niche strategic finding (§6) | `design/00-vision.md` adjacency / BACKLOG | No action needed beyond recording: mobile web is our mobile strategy by elimination (law 1 forecloses the app-funnel playbook; PEGI-18 storefront trap) |

---

## 10. Verify before shipping

Anything below that reaches shipped copy, an RFC acceptance criterion, or a tuned constant must
be re-verified first (house convention — `[M]`/`[P]` marked in body):

1. **On-device test matrix (the big one):** does a dedicated Worker's 50 ms loop keep firing in a
   hidden tab on current Chrome Android and iOS Safari 26? Exact iOS suspension latency after
   lock (seconds vs ~30 s)? Half-open socket behavior on resume (close event synthesized or not)?
   — one afternoon with two phones settles the four highest-leverage `[M]`s in §1.
2. Current default state of Chromium's `AllowAggressiveThrottlingWithWebSocket`; the exact Chrome
   version where open WebSockets stopped blocking bfcache (web.dev currently reads ~149 `[P]`).
3. Safari hidden-tab timer-alignment ceiling (~30 s `[M]`).
4. Touch tap-rate numbers (§5.3, all `[M]`) — run a small playtest before any clamp retune; do
   not ship the 5–14/s figures as fact.
5. Idle-game notification-culture claims (§3.2 `[M]` block) — do not quote as fact in any
   player-facing or design-doc text without sourcing.
6. PushEngage/PushPushGo opt-in percentages are vendor data `[V/P]` — fine for planning, not for
   published formulas or copy.
7. Melvor's exact wrapper tech (Cordova-family `[M]`); Antimatter Dimensions iOS app
   officialness; Orteil's stated mobile position — verify before naming any of them in satire or
   docs.
8. Centrifugo `client.recovery_max_publication_limit` exact key/default (~300 `[P]`) before
   citing in transport tuning.
9. infrequently.org 2025/2026 update to the A51-class baseline before pinning the CI perf budget.
10. iOS 26 "zero installability requirements" behavior on a real device (does a manifest-less
    A2HS truly open standalone for OUR SPA shell, and does the Astro-page-vs-SPA scope matter).

---

## 11. Key primary sources

Background/lifecycle: [Chrome 88 timer throttling](https://developer.chrome.com/blog/timer-throttling-in-chrome-88) ·
[Chrome 57 budget throttling](https://developer.chrome.com/blog/background_tabs) ·
[MDN setTimeout clamps](https://developer.mozilla.org/en-US/docs/Web/API/Window/setTimeout) ·
[Page Lifecycle API](https://developer.chrome.com/docs/web-platform/page-lifecycle-api) ·
[bfcache](https://web.dev/articles/bfcache) ·
[Memory Saver / wasDiscarded](https://developer.chrome.com/blog/memory-and-energy-saver-mode)

PWA/push: [Safari 26.0 features (A2HS change)](https://webkit.org/blog/17333/webkit-features-in-safari-26-0/) ·
[iOS 16.4 Web Push](https://webkit.org/blog/13878/web-push-for-web-apps-on-ios-and-ipados/) ·
[Declarative Web Push](https://webkit.org/blog/16535/meet-declarative-web-push/) ·
[Safari 16.4 (badging, wake lock)](https://webkit.org/blog/13966/webkit-features-in-safari-16-4/) ·
[ITP 7-day cap + A2HS exemption](https://webkit.org/blog/10218/full-third-party-cookie-blocking-and-more/) ·
[Apple DMA page](https://developer.apple.com/support/dma-and-apps-in-the-eu/) ·
[OWA blog (engine-ban timeline)](https://open-web-advocacy.org/blog/) ·
[Chrome install criteria](https://web.dev/articles/install-criteria) ·
[Richer install UI](https://developer.chrome.com/blog/richer-pwa-installation) ·
[Permission UX](https://web.dev/articles/push-notifications-permissions-ux) ·
[Chrome notification auto-revocation](https://support.google.com/chrome/answer/3220216) ·
[NN/g push guidelines](https://www.nngroup.com/articles/push-notification/) ·
[PushEngage benchmarks](https://www.pushengage.com/web-push-notification-benchmark-data-by-region/) ·
[dontkillmyapp](https://dontkillmyapp.com/)

Performance/touch: [Layout thrash](https://web.dev/articles/avoid-large-complex-layouts-and-layout-thrashing) ·
[MDN contain](https://developer.mozilla.org/en-US/docs/Web/CSS/contain) ·
[MDN will-change](https://developer.mozilla.org/en-US/docs/Web/CSS/will-change) ·
[Compositor properties/layer count](https://web.dev/articles/stick-to-compositor-only-properties-and-manage-layer-count) ·
[Animations guide (trace numbers)](https://web.dev/articles/animations-guide) ·
[HPBN mobile radio](https://hpbn.co/mobile-networks/) + [optimizing for mobile networks](https://hpbn.co/optimizing-for-mobile-networks/) ·
[Android Go](https://developer.android.com/guide/topics/androidgo) ·
[Performance inequality 2024](https://infrequently.org/2024/01/performance-inequality-gap-2024/) ·
[Lighthouse throttling](https://github.com/GoogleChrome/lighthouse/blob/main/docs/throttling.md) ·
[INP](https://web.dev/articles/inp) ·
[WCAG 2.5.8](https://www.w3.org/WAI/WCAG22/Understanding/target-size-minimum.html) ·
[Apple 44pt](https://developer.apple.com/design/tips/) ·
[Android 48dp](https://support.google.com/accessibility/android/answer/7101858) ·
[300ms gone](https://developer.chrome.com/blog/300ms-tap-delay-gone-away) ·
[WebKit tapping](https://webkit.org/blog/5610/more-responsive-tapping-on-ios/) ·
[MDN touch-action](https://developer.mozilla.org/en-US/docs/Web/CSS/touch-action) ·
[MDN overscroll-behavior](https://developer.mozilla.org/en-US/docs/Web/CSS/overscroll-behavior) ·
[svh/dvh](https://web.dev/blog/viewport-units) ·
[caniuse vibration](https://caniuse.com/vibration)

Transport: [Centrifugal client API (reconnect, background-disconnect advice)](https://centrifugal.dev/docs/transports/client_api) ·
[Centrifugo server config (ping 25 s / pong 8 s)](https://centrifugal.dev/docs/server/configuration) ·
[History & recovery](https://centrifugal.dev/docs/server/history_and_recovery) ·
[centrifuge-js defaults](https://github.com/centrifugal/centrifuge-js) ·
[AWS full jitter](https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/) ·
[MDN navigator.onLine](https://developer.mozilla.org/en-US/docs/Web/API/Navigator/onLine) ·
[Cloudflare WebSockets](https://developers.cloudflare.com/network/websockets/)

Genre: [Cookie Clicker (Wikipedia)](https://en.wikipedia.org/wiki/Cookie_Clicker) ·
[Melvor Idle (Steam)](https://store.steampowered.com/app/1267910/Melvor_Idle/) ·
[Universal Paperclips iOS](https://apps.apple.com/us/app/universal-paperclips/id1300634274) ·
[Kittens Game iOS](https://apps.apple.com/us/app/kittens-game/id1198099725) ·
[AD source](https://github.com/IvarK/AntimatterDimensionsSourceCode)
