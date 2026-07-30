# Faction & Incorporation — implementation log

## 2026-07-30 — accepted and started

- Owner answered the acceptance bounce with FA/FB/FC: stocks are run-scoped int64 Company fields,
  accrual is one unit per attended minute with carried milliseconds and accrual-only saturation,
  and all four Phase-0 faction objects are literal.
- Acceptance re-read confirmed these contracts close the previously recorded resource/rate/data
  gaps. RFC moved to implementing; no inferred balance mechanics are required.
- Implementation order follows the dependency surface: catalog/identity → save shape → intents and
  Commons binding → accrual → boards/reset/docs.
