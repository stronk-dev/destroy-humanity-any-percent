"""Exact conditional WL2 models; no runtime/balance/formula selection."""
import hashlib
import json
from fractions import Fraction as F
from pathlib import Path


def text(value):
    return f"{value.numerator}/{value.denominator}"


def uniform_total(count, multiplier):
    return count * multiplier(F(1, count))  # each player has raw output 1


def meaningful_invariance(totals):
    return all(value > 0 for value in totals) and len(set(totals)) == 1


populations = [1, 200, 20_000, 200_000]
linear = [uniform_total(n, lambda share: share) for n in populations]
decreasing = [uniform_total(n, lambda share: 1 - share) for n in populations]
zero = [F(0) for _ in populations]
assert linear == [F(1)] * 4
assert decreasing == [F(0), F(199), F(19_999), F(199_999)]
assert meaningful_invariance(linear)
assert not meaningful_invariance(decreasing)
assert not meaningful_invariance(zero)

# Equal-population invariance requires 200*d(1/200) = 200000*d(1/200000).
# The common positive normalizer cancels, so the multiplier ratio is forced.
high_share, low_share = F(1, 200), F(1, 200_000)
forced_multiplier_ratio = F(200_000, 200)
raw_ratio = F(1000)
forced_credit_ratio = raw_ratio * forced_multiplier_ratio
assert high_share / low_share == raw_ratio == forced_multiplier_ratio
assert forced_credit_ratio == 1_000_000 > raw_ratio

groups = [(1, F(1000)), (199_000, F(1))]
total_raw = sum(count * output for count, output in groups)
assert total_raw == 200_000
assert groups[0][1] / total_raw == high_share
assert groups[1][1] / total_raw == low_share
linear_credits = [output * (output / total_raw) for _, output in groups]
assert linear_credits[0] / linear_credits[1] == forced_credit_ratio
decreasing_ratio = raw_ratio * (1 - high_share) / (1 - low_share)
assert decreasing_ratio < raw_ratio

# Feasibility control only: concave FINAL weights, normalized to a unit total.
# This introduces a peer-weight sum and removes the second raw-output multiply.
def weight(share):
    return 2 * share - share * share


weight_sum = sum(count * weight(output / total_raw) for count, output in groups)
normalized = [weight(output / total_raw) / weight_sum for _, output in groups]
normalized_total = sum(count * credit for (count, _), credit in zip(groups, normalized))
assert normalized_total == 1
alternative_ratio = normalized[0] / normalized[1]
assert 0 < alternative_ratio < raw_ratio
for n in populations:
    one_weight = weight(F(1, n))
    assert one_weight > 0 and n * (one_weight / (n * one_weight)) == 1

sources = ["rfc/world-layer-foundation.md", "design/05-mmo.md",
           "design/02-economy-balancing.md", "server/gameserver/world_source.go"]
result = {
    "schema_version": 1,
    "source_base": "40c14203a813450febf6afb16cd414b88adae119",
    "source_sha256": {name: hashlib.sha256(Path(name).read_bytes()).hexdigest() for name in sources},
    "scope": "conditional exact-rational model; not an admitted runtime or selected balance formula",
    "uniform_counts": populations,
    "linear_multiplier_uniform_totals": [text(x) for x in linear],
    "decreasing_multiplier_uniform_totals": [text(x) for x in decreasing],
    "all_zero_invariance_valid": meaningful_invariance(zero),
    "heterogeneous_groups": [{"count": n, "raw_per_player": text(q)} for n, q in groups],
    "raw_ratio": text(raw_ratio),
    "forced_multiplier_ratio": text(forced_multiplier_ratio),
    "forced_credit_ratio": text(forced_credit_ratio),
    "decreasing_factor_credit_ratio": text(decreasing_ratio),
    "separately_normalized_final_weight_ratio": text(alternative_ratio),
    "separately_normalized_total": text(normalized_total),
    "property_outcomes": {
        "linear_multiplier_meaningful_uniform_invariance": meaningful_invariance(linear),
        "linear_multiplier_pair_non_amplification": forced_credit_ratio <= raw_ratio,
        "decreasing_multiplier_meaningful_uniform_invariance": meaningful_invariance(decreasing),
        "decreasing_multiplier_pair_non_amplification": decreasing_ratio <= raw_ratio,
        "decreasing_multiplier_solo_credit_positive": decreasing[0] > 0,
        "normalized_final_weight_pair_non_amplification": alternative_ratio <= raw_ratio,
        "normalized_final_weight_heterogeneous_unit_total": normalized_total == 1
    },
    "counterfactual_formula_selected": False,
    "objective_completed": True,
    "limitations": ["active fraction / raw share / non-amplification are explicit model interpretations",
                    "no curve grammar, ppm rounding, Decimal implementation, real population or pacing proof",
                    "no per-faction ranking, NPC production, zero-population or reward implementation"]
}
print(json.dumps(result, indent=2))
