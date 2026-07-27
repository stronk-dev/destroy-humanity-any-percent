# Numeric Normalization Carry — Running Log

## 2026-07-28

- RFC-0002's one-million-source ledger regression failed after the tenth `1e87` entry.
- Diagnostic representation was mantissa `10`, exponent `87`; canonical rendering hid it as
  `1e88`, while `IsStateValue` correctly rejected it.
- Root cause: the one-pass `Log10`/`Pow10` normalizer can undershoot its scale by one because of
  IEEE-754 rounding and does not correct the resulting boundary mantissa.
