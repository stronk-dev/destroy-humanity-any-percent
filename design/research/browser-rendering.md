# Research: Browser Game Rendering & Engine Tech — 2026 State of the Art

> Web research agent report, July 2026. Worked via direct WebFetch of primary sources (caniuse, MDN, vendor docs, GitHub, engine blogs); unverifiable items flagged. Feeds the client-engine RFC.

## 0. Executive summary

| Question | Answer |
|---|---|
| Spectacle layer tech | **PixiJS v8 (v8.19.x), `preference: 'webgpu'` with automatic WebGL2 fallback**, `ParticleContainer` for mass-unit/particle layers |
| Can it hit "render every rack/agent"? | Yes — Pixi v8 `ParticleContainer` benchmarks at **1,000,000 particles @ 60fps**; regular sprites+Container at **200,000 @ 60fps** |
| Does the CSS-sprite pet system hold up? | **Yes, under a hard node budget** (~a few hundred DOM nodes; Lighthouse warns at 800 body nodes, errors at 1,400). CSS pets while they're a *cast*; Pixi the moment they're a *crowd*. |
| Threading | Sim in a **dedicated Worker** (`postMessage` + transferable `ArrayBuffer` — **avoid SharedArrayBuffer/COOP-COEP unless forced**). Render on main thread; OffscreenCanvas is a later optimization. |
| WASM | **No.** Not worth it for an idle sim; JS JIT + closed-form catch-up wins. |
| Tweening | **GSAP** — now **100% free including all plugins** (Webflow). Changes the calculus. |

## 1. Rendering tech

### 1.1 WebGPU support (2026)

From the [GPUWeb Implementation Status wiki](https://github.com/gpuweb/gpuweb/wiki/Implementation-Status) and [caniuse](https://caniuse.com/webgpu):

- **Chrome/Edge:** shipped v113+ (macOS/Windows/ChromeOS); Android v121+ (ARM/Qualcomm/Intel GPUs; Imagination from v139); **Linux: Intel Gen12+ in v144, NVIDIA in v147**.
- **Firefox:** Windows v141 (July 2025); Apple Silicon macOS v145; other macOS v147; **Linux+Android targeted 2026**.
- **Safari:** default across macOS Tahoe 26 / iOS 26 / iPadOS 26 / visionOS 26.
- **Global usage ~83.6%.** (caniuse lags on Firefox; the wiki is authoritative.)

**Read:** WebGPU is majority, not universal. Cannot ship WebGPU-only; CAN ship WebGPU-preferred. **WebGL2 is the baseline at 94.67%** ([caniuse](https://caniuse.com/webgl2)); below that (<5%) gets a degraded/static presentation.

### 1.2 What WebGPU buys ([Chrome](https://developer.chrome.com/blog/from-webgl-to-webgpu))

Immutable pipelines (fewer CPU state stalls), fully async design, **compute shaders** (GPU particle sim/flocking), storage buffers 128MB+ vs WebGL's 64KB uniforms, render bundles, better debugging. **Caveat from PixiJS:** "WebGPU doesn't automatically guarantee better performance… PixiJS often encounters CPU limitations before GPU bottlenecks" — WebGPU wins mainly with many batch breaks (filters/masks/blends); plain 2D sprite batching is often equally fast on WebGL2.

### 1.3 PixiJS v8 (v8.19.x; no v9 announced)

**v8 vs v7 Bunnymark, 100,000 sprites** ([launch post](https://pixijs.com/blog/pixi-v8-launches)):

| Scenario | CPU v7 → v8 | GPU v7 → v8 |
|---|---|---|
| All moving | ~50ms → ~15ms | ~9ms → ~2ms |
| **Static** | **~21ms → ~0.12ms** | ~9ms → ~0.5ms |
| Restructuring | ~50ms → ~24ms | ~9ms → ~2ms |

The static number is the headline: a mostly-static datacenter of 100k sprites costs **~0.12ms CPU/frame**.

Key features: **Render Groups** (GPU-transform containers = hardware 2D camera for pan/zoom over a huge world), inherited blend modes/tints (flash a whole cluster with one property), Canvas2D-mirroring Graphics API, **`pixi.js/html-source` (v8.19)** — live interactive DOM rendered into Pixi textures — and the WebGPU **`transient`** texture flag (discard MSAA buffers; real mobile win on tile-based GPUs).

### 1.4 ParticleContainer ([blog](https://pixijs.com/blog/particlecontainer-v8))

- **1,000,000 particles @ 60fps** (vs 200k sprites); >3× v7.
- Mechanism: explicit **static vs dynamic property declaration** — static props upload only on explicit `update()`. "The fewer dynamic properties, the faster."
- Constraints: `addParticle()` API, no children/events/filters per particle.
- Quote: "the main bottleneck at that point wasn't even rendering but the logic behind the movement" — **the JS sim bottlenecks before the renderer.**
- **Implication:** the rack/agent layer = a ParticleContainer with mostly-static props, not thousands of Sprites.

### 1.5 Alternatives

- **Phaser 4** (4.2.1, active): full framework (scenes/physics/input) that wants to own the loop — competes with Svelte. Its `SpriteGPULayer` does ~1M quads BUT: WebGL-only, one texture, shader-driven animation only — **JS-driven per-member mutation triggers whole-layer buffer regeneration**. Disqualifying for units that constantly change state. Choose only if you want the batteries.
- **Three.js:** strongest WebGPU story, but a 3D stack you don't need. Only if real-3D spectacle moments become a goal.
- **Kaplay / Kontra:** prototyping/size-golf niches; not composable with a Pixi stage. No.
- **Raw Canvas2D:** fine for <500-element boards and the low-tier fallback; falls off a cliff at low thousands of draws. Modern additions: `roundRect`, `reset()`, `contextlost/restored` events.
- **`@pixi/particle-emitter` v5:** editor + 30 example effects, but **Pixi v8 compatibility unconfirmed — verify or budget ~1–2 days for a hand-rolled pooled emitter** (~200 lines on ParticleContainer).
- 100k+ particles with per-particle physics → WebGPU compute (custom); defer until actually needed.

## 2. Hybrid DOM + Canvas

### 2.1 The pattern

Canvas absolute at z-0; UI overlay at z-1 with **`pointer-events: none` on the root, `auto` on interactive panels** (that's the whole input router). Promote the canvas to its own compositor layer (`will-change: transform`). **Never animate DOM top/left/width — only `transform`/`opacity`.** Coordinate sync (tooltips pinned to canvas objects) through ONE per-frame function using a cached `getBoundingClientRect` + camera transform (invalidate via ResizeObserver). Batch DOM reads/writes; one forced sync layout costs more than 10k sprite draws. **DOM budget: Lighthouse warns ~800 body nodes, errors ~1,400** ([docs](https://developer.chrome.com/docs/lighthouse/performance/dom-size/)).

### 2.2 Svelte 5 integration

Use **`{@attach}` attachments** (not `use:` actions — [docs](https://svelte.dev/docs/svelte/@attach); actions are superseded). **Do NOT let Svelte reactivity drive per-frame canvas updates** — a `$effect` re-running 60×/s destroys the frame budget. Boundary: **Svelte owns UI state → coarse commands to Pixi (`stage.setRackCount(n)`) → Pixi owns per-frame.** No official Pixi-Svelte binding exists (only `@pixi/react`); hand-roll ~50 lines with an attachment — the better architecture at these object counts anyway.

### 2.3 How far CSS alone goes (pet system verdict)

Verified support: individual `translate/rotate/scale` **93.2%**; `@property` typed custom props **92.9%**; View Transitions **88.5%** (Chrome 111, Safari 18, Firefox 144); `prefers-reduced-motion` ~100%. (Scroll-driven animations: unverified — check before relying.)

**The cattery approach holds up for a bounded cast:**
- `steps()` + sprite sheets = compositor-driven frame animation, zero JS.
- **Individual transform properties are the underrated unlock:** hop (`translate`) + wobble (`rotate`) + squash (`scale`) as three independent animations on one element.
- **`@property` is the second unlock:** register `--wiggle: <angle>`, interpolate custom properties, compose via `calc()` — per-pet personality driven from CSS variables Svelte sets reactively. The closest CSS has to per-instance shader params.
- Free: accessibility tree, DPI independence, hit-testing, reduced-motion via one media query.

**Where it breaks:** node count (~10 nodes/pet → warning territory at 50–80 pets before UI); no z-interleaving with canvas content (pets can't walk *behind* racks); compositor-layer explosion on mobile (every animating element = a GPU layer with texture memory); no shaders (no glow/dissolve/palette-FX — CSS `filter` on 60 animating elements tanks mobile); per-frame JS position writes painful past ~200 elements.

**Recommendation: keep CSS pets with a hard cap (~60 visible, ~300 nodes) and build a `PetRenderer` interface seam** (`spawn/move/setState/despawn`) so a `PixiPetRenderer` can replace `CssPetRenderer` without touching game logic. The forcing functions to port: shader effects, crowds, or pets-behind-racks.

## 3. Performance infrastructure

### 3.1 Workers & OffscreenCanvas

OffscreenCanvas at **93.8%** (incl. iOS Safari 17). But: no style sizing (forward resize/DPR manually), **no input events in the worker** (forwarding adds a frame of click latency — bad for a *clicker*), library caveats. **Recommendation: don't start with it.** Render cost isn't the problem (~0.12ms); the **sim** is. Sim in a worker, render on main.

### 3.2 SharedArrayBuffer: no

SAB needs COOP/COEP cross-origin isolation, which breaks OAuth popups, cross-origin CDN assets, embeds — each needs CORP/CORS auditing; `credentialless` COEP lacks broad support. Our worker↔main traffic is a 10–20 Hz snapshot, not a 60 Hz physics buffer: **transferable `ArrayBuffer`s via `postMessage` (zero-copy) suffice.** If a future minigame truly needs SAB, isolate it on a separate cross-origin-isolated route.

### 3.3 WASM: no

WASM wins for tight numeric loops without boundary crossings; an idle sim is many small calls over a heterogeneous object graph. Offline catch-up's correct fix is **closed-form math**, not a faster loop (which we already committed to server-side). Revisit only if a profiled worker tick exceeds ~4ms — and fix algorithmically first (SoA layout, closed forms).

### 3.4 Background-tab behavior (critical for an idle game)

- **rAF does not run at all in hidden tabs.** Logic on rAF = game stops when backgrounded.
- Timer throttling ([Chrome 88+](https://developer.chrome.com/blog/timer-throttling-in-chrome-88)): visible or sounded-in-30s → minimal; hidden → **once per second**; hidden >5 min + chain≥5 + silent + no WebRTC → **once per minute** (intensive).
- Page Lifecycle **frozen** stops everything; **discarded** unloads (check `document.wasDiscarded`).
- **The architecture that survives all of it:** (1) progress = `f(wall-clock dt)`, never tick counts — throttle/freeze/sleep become the same code path; (2) sim tick on `setInterval` in a Worker (less throttled, not immune); (3) rAF renders only, stopped when hidden; (4) `visibilitychange` → catch-up; (5) **use `pagehide`, never `unload`** (unreliable + kills bfcache); persist on 30s interval + hidden + pagehide; (6) on `freeze`: close IndexedDB, disconnect channels, release locks. Server is authoritative anyway (MMO) — client catch-up is presentation.

### 3.5 Memory for multi-day tabs

Leak sources ranked: detached DOM nodes (floating numbers/tooltips held by arrays/closures/timers — filter "Detached" in heap snapshots); unremoved listeners (attachments' cleanup functions); Pixi textures (auto-GC exists, but `destroy()` one-offs and **stagger mass destruction**); unbounded arrays (feeds/logs/stat samples → ring buffers); per-frame `RenderTexture`/`Graphics` regeneration. Practices: **object-pool everything transient** (allocation pressure = GC pauses = micro-jank); SoA `Float64Array` entity data; debug HUD with `performance.memory` + Pixi object counts + DOM node count; **24 h accelerated soak test with hourly heap snapshots** (the only way to catch slow leaks); handle `contextlost/contextrestored` (multi-day mobile sessions WILL lose GPU context).

## 4. Juice toolkit

### 4.1 Tweening: GSAP (now fully free incl. all plugins — [pricing](https://gsap.com/pricing/))

Why: **timelines** (choreograph "hit → flash → shake → number pops → coin flies → counter ticks" declaratively); tweens **arbitrary JS objects** — one engine for Pixi sprites, camera, CSS pets, and counters; `gsap.ticker` slaves to Pixi's ticker; `quickTo()` for per-frame writes without allocation; the juice easings (`elastic`, `back`, `bounce`). CSS/WAAPI for pure UI state changes (compositor-thread, survives main-thread jank). Motion/anime.js: fine, unnecessary now.

### 4.2 The juice canon, mapped

Squash & stretch on click (GSAP `back.out`, ~120ms — cheapest, highest-impact) · floating numbers = **pooled Pixi `BitmapText`**, never DOM, never `Text` (canvas-redraw overhead), randomized drift · **screen shake = trauma model** (trauma 0–1, decay linearly, shake = trauma², sample **noise not random**, translate the camera Render Group, gate on reduced-motion) · particles on every event (8/click transforms feel) · hit-stop (ticker speed 0 for 40–80ms) · tween every number · tint pulse via inherited v8 tint · anticipation (`back.in`) · trails via ghost sprites/low-alpha clear.

**Idle-specific problem: juice at 10,000 events/sec.** Solutions: **aggregate** (batch N events into one bigger visual), **stochastic sampling** (1-in-K spawns, scaled up), per-emitter rate budgets, **representation switching at thresholds** (individual numbers → stream → ambient glow intensity). Design the aggregation strategy BEFORE building effects.

### 4.3 Audio

WebAudio autoplay: create/resume the `AudioContext` in the first click handler — free for a clicker. **Use howler.js** (~7KB gz, audio sprites, mobile unlock handled — [repo](https://github.com/goldfire/howler.js)). Patterns needed: **voice limiting** (cap concurrent instances per SFX), **pitch randomization ±8%** (the cheapest repetition-fatigue fix), ducking, one audio-sprite sheet, Master/Music/SFX/UI gain buses persisted. `AudioParam` ramp methods, never `.value=` for timed changes.

### 4.4 Haptics

`navigator.vibrate` — Android only, silently no-ops elsewhere; 8–20ms taps, user-toggleable, default conservative. No iOS web haptics; don't ship the checkbox-switch hack.

## 5. Mobile web + PWA

- **Install:** iOS 16.4+ installable from Share menu in all major browsers but **no `beforeinstallprompt`** — teach the flow with a coach mark. Android: full support + custom prompts. Manifest needs 192+512 icons, start_url, display, HTTPS.
- **Push: 95.1% global** — but iOS only for Home-Screen-installed apps. Design retention so push is a bonus for installers, never the primary mechanic ("install for notifications" incentive).
- **No reliable background execution on mobile web.** Backgrounded PWAs freeze/die in seconds-to-minutes (iOS aggressive). Server-authoritative offline progress is mandatory (already our law). No Background Sync as a mechanic.
- **Tile-based GPUs** (Apple/Mali/Adreno): avoid MSAA (use the `transient` flag), render-target switches, **overdraw** (the #1 mobile killer — small particles, not big quads), full-screen filters. `@0.5x` asset variants (native Pixi support). ≤16 textures/batch — atlas aggressively. **Cap DPR at 2** (low-tier: 1.5 or 1).
- **Touch:** Pointer Events only; `touch-action: none` on canvas; handle **`pointerdown` not `click`** (latency); support multi-touch clicking (two thumbs). Safe areas: `viewport-fit=cover` + `env(safe-area-inset-*)` with fallbacks; `dvh/svh` not `vh`. **Wake Lock** (Baseline 2025): opt-in "keep screen awake" setting only, re-acquire on visibility.
- **Battery:** 30fps when idle, stop rendering when hidden, particle budgets down on low battery (hint only).

## 6. Notable browser games (weakest section — flagged)

Verified: **Vampire Survivors** started on Phaser, **migrated to Unity from v1.6** (Wikipedia) — cautionary, though driven partly by console ports and predating Pixi v8/Phaser 4 by years. **Antimatter Dimensions** = Vue+DOM entirely (proves deep incrementals can be pure DOM; also shows the spectacle ceiling). **Cookie Clicker** = (unverified detail) hand-rolled DOM + small canvas — i.e., exactly our hybrid architecture, built in 2013 without a renderer.

**Confident claim: no browser incremental has achieved Gnorp-level presentation. That niche is open, and 2026 tech (Pixi v8 ParticleContainer + render-group cameras) is precisely the missing primitive.**

Unverified leads for follow-up: io-game tech stacks (entity interpolation at 100+ visible entities), WebGPU showcase titles, current jam standouts.

## 7. Accessibility & performance tiers

- **Reduced motion** (baseline everywhere): respect OS setting by default, offer independent in-game toggle, **replace don't remove** (scale-pulse → opacity dissolve). Kill: shake, zoom punch, parallax (vestibular triggers). Safe: opacity, color, small shifts. Listen for mid-session changes.
- Mapping table: shake off · zoom off · parallax off · 40 particles → 6 or a flash · floating numbers fade in place · panels crossfade · pet anims slowed/static · counters snap.
- **Colorblind-safe feedback** (WCAG 1.4.1): encode in **shape and motion, not just hue** (crit = bigger + different sprite; buffs rise, debuffs sink); vary luminance (3:1) not just hue; glyphs on every colored status; CVD palette presets; **never red-green as the only axis** (blue-orange safest); high-contrast mode + a "reduce visual noise" toggle (helps CVD + low-vision + low-end simultaneously).
- **Performance tiers — measure, don't guess:** seed from `@pmndrs/detect-gpu` (caveat: **benchmark data stale since Dec 2025**) + `hardwareConcurrency` + `deviceMemory` (Chromium-only) + DPR; then a **rolling 120-frame time monitor**: >20ms avg for 3s → step down (dismissible toast); <10ms for 30s → step up with hysteresis; **user override always wins**; persist in localStorage. Tier table: particles 500/5k/30k/150k · individually-rendered units aggregate/500/5k/all · DPR 1/1.5/2/2 · filters none/none/hero-only/full · pet anim static/4fps/12fps/24fps · shake off/subtle/full/full · 30/60/60/60fps.

## 8. Final architecture recommendation

**PixiJS v8, WebGPU-preferred, WebGL2 fallback:**

```js
await app.init({ preference: 'webgpu', antialias: false,
  resolution: Math.min(devicePixelRatio, 2), autoDensity: true,
  powerPreference: 'high-performance', backgroundAlpha: 0 });
```

Layer mapping: static world = Container in a RenderGroup (GPU camera) · racks/agents = **ParticleContainer** (position dynamic; tint/scale static + explicit `update()`) · particles = second ParticleContainer with pooled emitters · floating numbers = pooled BitmapText · minigame boards = regular Containers · camera shake/zoom = root RenderGroup transform · bloom/glow = high-tier only.

**Threading:** Svelte 5 (UI, `{@attach}` mount) + Pixi (render, rAF) + Howler on main; **sim in a dedicated Worker** (20 Hz `setInterval`, wall-clock dt, SoA Float64Array state, IndexedDB persistence off-main); snapshots 10–20 Hz via transferable ArrayBuffers; commands up. Server authoritative above it all. No SAB, no WASM, no OffscreenCanvas day one.

**Dependency list:** pixi.js v8.19.x · GSAP · howler.js · @pmndrs/detect-gpu (seed only) · CSS+`@property` for UI micro-animation and pets.

**Named risks:** (1) Vampire Survivors precedent — benchmark the worst-case scene (10k racks + 30k particles + full UI) on a mid-tier Android in month one; (2) particle-emitter v8 compat unconfirmed — budget the hand-rolled emitter; (3) juice-at-scale needs the aggregation strategy designed first; (4) DOM node budget — instrument node count in the debug HUD from day one; (5) set up the 24 h soak test before players run week-long sessions.

## Sources

caniuse (webgpu, webgl2, offscreencanvas, sharedarraybuffer, view-transitions, mdn-css translate/@property, push-api) · GPUWeb Implementation Status wiki · Firefox 141 notes · Chrome dev blog (WebGL→WebGPU, WebGPU overview/140, timer throttling 88, Page Lifecycle, canvas2d, Lighthouse DOM size, DevTools memory) · PixiJS (v8 launch, ParticleContainer blog+guide, perf tips, migration, releases) · @pixi/react · @pixi/particle-emitter · Phaser (news, README, SpriteGPULayer docs) · Kaplay · Kontra · three.js manual · web.dev (offscreen-canvas, coop-coep, wasm perf patterns) · MDN (prefers-reduced-motion, env(), Wake Lock, Vibration, WebAudio best practices, deviceMemory, PWA install, WAAPI) · Svelte {@attach} docs · GSAP pricing · howler.js · @pmndrs/detect-gpu · WCAG 2.2 1.4.1 · Wikipedia (Vampire Survivors) · AD source repo.

**Gaps flagged by the agent:** notable-browser-games survey (io-game stacks, WebGPU showcases); scroll-driven animation support numbers; particle-emitter v8 compat; exact Pixi 8.19 release date.
