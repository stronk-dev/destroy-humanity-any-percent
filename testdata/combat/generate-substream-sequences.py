"""Print C4's sequence fixture; never import production code or write files.

The nth state is computed directly from seed + n*increment, independently of
the mutable Go/TypeScript generators. Existing seed-42 first draws are controls.
"""
import json

MASK = (1 << 64) - 1
INCREMENT = 0x9E3779B97F4A7C15


def draw(seed, ordinal):
    value = (seed + ordinal * INCREMENT) & MASK
    value = ((value ^ (value >> 30)) * 0xBF58476D1CE4E5B9) & MASK
    value = ((value ^ (value >> 27)) * 0x94D049BB133111EB) & MASK
    return value ^ (value >> 31)


def label_hash(label):
    value = 0xCBF29CE484222325
    for byte in label.encode("utf-8"):
        value = ((value ^ byte) * 0x100000001B3) & MASK
    return value


labels = ["crit", "obedience", "bot_policy"]
rows = []
for match_seed in [0, 42, MASK]:
    battle = draw(match_seed, 1)
    rows.append({"match_seed": str(match_seed), "battle_seed": str(battle),
                 "streams": {label: [str(draw(battle ^ label_hash(label), n))
                                     for n in range(1, 9)] for label in labels}})

control = rows[1]
assert control["battle_seed"] == "13679457532755275413"
assert [control["streams"][label][0] for label in labels] == [
    "15610968233845892314", "2306868096585562568", "13127054933185908140"]
print(json.dumps({"version": 1, "labels": labels, "draw_count": 8, "seeds": rows}, indent=2))
