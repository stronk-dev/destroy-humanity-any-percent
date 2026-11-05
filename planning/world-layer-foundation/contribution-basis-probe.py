"""Exact candidate comparison only; no adopted formula or product implementation."""
import hashlib
import json
from fractions import Fraction as F
from pathlib import Path


def text(x):
    return f"{x.numerator}/{x.denominator}"


def subtotal(groups, cap=F(10), weight=F(1)):
    if cap <= 0 or weight < 0:
        raise ValueError("invalid cap/weight")
    if any(type(n) is not int or n <= 0 or x < 0 for n, x in groups):
        raise ValueError("invalid count/output")
    count = sum(n for n, _ in groups)
    return weight * sum(n * min(x, cap) for n, x in groups) / count if count else F(0)


def meaningful_invariance(values):
    return all(x > 0 for x in values) and len(set(values)) == 1


counts = [1, 200, 20_000, 200_000]
uniform = [subtotal([(n, F(1))]) for n in counts]
uncapped_sum = [F(n) for n in counts]
zero = [subtotal([(n, F(0))]) for n in counts]
assert uniform == [F(1)] * 4 and meaningful_invariance(uniform)
assert not meaningful_invariance(uncapped_sum)
assert not meaningful_invariance(zero)
assert subtotal([]) == 0
assert subtotal([(200, F(1000))]) == 10

replicas = [1, 10, 1000]
factions = [subtotal([(50 * k, F(1))], weight=F(1, 2)) +
            subtotal([(150 * k, F(3))], weight=F(1, 2)) for k in replicas]
assert factions == [F(2)] * 3
empty_faction_total = subtotal([(200, F(1))], weight=F(1, 2)) + subtotal([], weight=F(1, 2))
assert empty_faction_total == F(1, 2)  # no silent redistribution

cap, n = F(10), 200_000
high, low = F(1000), F(1)
high_credit, low_credit = min(high, cap) / n, min(low, cap) / n
heterogeneous_total = subtotal([(1, high), (n - 1, low)])
assert heterogeneous_total == F(200_009, 200_000)
assert high_credit / low_credit == 10 <= high / low == 1000
assert high_credit == cap / n
single_positive_total = subtotal([(1, high), (n - 1, F(0))])
assert single_positive_total > 0
single_positive_share = high_credit / single_positive_total
assert single_positive_share == 1  # a strict fractional-share bound FAILS

# Finite exact controls complement the algebra; they are not a universal
# floating-point/rounding or game-balance proof.
samples = [F(0), F(1, 10), F(1), F(9), F(10), F(11), F(1000)]
for a, b in zip(samples, samples[1:]):
    assert min(a, cap) <= min(b, cap)
for a in samples:
    assert F(0) <= min(a, cap) / n <= cap / n
    for b in samples:
        if a >= b > 0:
            assert min(a, cap) / min(b, cap) <= a / b
refused = []
for name, groups, limit in [("negative_output", [(1, F(-1))], cap),
                            ("zero_cap", [(1, F(1))], F(0)),
                            ("negative_cap", [(1, F(1))], F(-1))]:
    try:
        subtotal(groups, cap=limit)
    except ValueError:
        refused.append(name)
    else:
        raise AssertionError(f"invalid input admitted: {name}")
assert len(refused) == 3

sources = ["rfc/world-layer-foundation.md", "design/05-mmo.md",
           "planning/world-layer-foundation/contribution-basis-proposal.md",
           "planning/world-layer-foundation/contribution-basis-probe.py",
           "planning/world-layer-foundation/producer-boundary.json"]
result = {
    "schema_version": 1,
    "source_base": "3aa86e6d",
    "scope": "owner-choice candidate, exact rational model; not adopted gameplay",
    "source_sha256": {p: hashlib.sha256(Path(p).read_bytes()).hexdigest() for p in sources},
    "mathematical_unit_cap_not_balance": text(cap),
    "uniform_counts": counts,
    "uniform_totals": [text(x) for x in uniform],
    "uncapped_sum_totals": [text(x) for x in uncapped_sum],
    "all_zero_meaningful_invariance": meaningful_invariance(zero),
    "faction_replication_factors": replicas,
    "faction_totals": [text(x) for x in factions],
    "empty_faction_total_no_redistribution": text(empty_faction_total),
    "heterogeneous_total": text(heterogeneous_total),
    "high_to_low_raw_ratio": text(high / low),
    "high_to_low_credit_ratio": text(high_credit / low_credit),
    "per_founder_absolute_bound": text(cap / n),
    "single_positive_total": text(single_positive_total),
    "single_positive_fraction_of_human_total": text(single_positive_share),
    "strict_fractional_share_bound_below_one_satisfied": False,
    "invalid_inputs_refused": refused,
    "adopted": False,
    "implementation_authority": False,
    "limitations": ["no qualified-source/time/population/rank/rounding grammar",
                    "no NPC production, solo viability or full anti-domination proof",
                    "no Decimal/ppm/runtime or actual player/pacing measurement",
                    "original World/design bodies unchanged; owner direction choice required"]
}
print(json.dumps(result, indent=2))
