# Founder Attendance Foundation implementation plan

- [x] Land immutable Founder genesis and make every first Founder-log activation create it atomically.
- [x] Record Exit's Founder mutation in `founder_log` with an exact Company-run identity and closed Founder facts.
- [x] Implement Go and TypeScript Founder replay from genesis across ordinary commands and Exits.
- [ ] Implement the race-safe, offline-aware effective-attendance resolver without a second persisted cursor.
- [ ] Add shared parity vectors and real-Postgres race/rollback/retention coverage.
- [ ] Update canonical docs, run root verification, obtain independent full-range review, and archive.

Checkboxes flip only with the test that proves the behavior.
