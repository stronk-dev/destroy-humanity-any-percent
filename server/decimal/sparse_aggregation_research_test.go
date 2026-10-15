package decimal

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// Unapplied prototype. Integer chunks represent the exact binary inputs divided
// by 2^52; no intermediate Decimal addition or exponent-sized integer is used.
const sparseBaseResearch int64 = 1_000_000_000

type sparseDigitResearch struct{ index, digit int64 }
type sparsePieceResearch struct{ first, last, digit int64 }
type sparseWorkResearch struct{ chunks, normalize, events, runs, reads int }

func sparsePositionResearch(exponent int64) (int64, int64) {
	index, remainder := exponent/9, exponent%9
	if remainder < 0 {
		index--
		remainder += 9
	}
	return index, remainder
}

func sparseNormalizeResearch(source map[int64]*big.Int, work *sparseWorkResearch) []sparseDigitResearch {
	keys := make([]int64, 0, len(source))
	for key := range source {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	result := []sparseDigitResearch{}
	var carry, value, remainder big.Int
	base := big.NewInt(sparseBaseResearch)
	index, next := int64(0), 0
	for next < len(keys) || carry.Sign() != 0 {
		if carry.Sign() == 0 {
			index = keys[next]
		}
		value.Set(&carry)
		if next < len(keys) && keys[next] == index {
			value.Add(&value, source[index])
			next++
		}
		carry.QuoRem(&value, base, &remainder)
		if remainder.Sign() != 0 {
			result = append(result, sparseDigitResearch{index, remainder.Int64()})
		}
		work.normalize++
		index++
	}
	return result
}

func sparseCompareResearch(a, b []sparseDigitResearch) int {
	for i, j := len(a)-1, len(b)-1; i >= 0 || j >= 0; {
		if i < 0 {
			return -1
		}
		if j < 0 {
			return 1
		}
		if a[i].index != b[j].index {
			if a[i].index < b[j].index {
				return -1
			}
			return 1
		}
		if a[i].digit != b[j].digit {
			if a[i].digit < b[j].digit {
				return -1
			}
			return 1
		}
		i--
		j--
	}
	return 0
}

func sparseSubtractResearch(a, b []sparseDigitResearch, work *sparseWorkResearch) ([]sparsePieceResearch, error) {
	pieces := []sparsePieceResearch{}
	borrow, previous := int64(0), int64(0)
	for i, j := 0, 0; i < len(a) || j < len(b); {
		index := int64(0)
		if j == len(b) || i < len(a) && a[i].index < b[j].index {
			index = a[i].index
		} else {
			index = b[j].index
		}
		if work.events > 0 && borrow != 0 && index > previous+1 {
			pieces = append(pieces, sparsePieceResearch{previous + 1, index - 1, sparseBaseResearch - 1})
			work.runs++
		}
		value := -borrow
		if i < len(a) && a[i].index == index {
			value += a[i].digit
			i++
		}
		if j < len(b) && b[j].index == index {
			value -= b[j].digit
			j++
		}
		borrow = 0
		if value < 0 {
			value += sparseBaseResearch
			borrow = 1
		}
		if value != 0 {
			pieces = append(pieces, sparsePieceResearch{index, index, value})
		}
		previous = index
		work.events++
	}
	if borrow != 0 {
		return nil, fmt.Errorf("invalid final sparse borrow")
	}
	return pieces, nil
}

func sparseChunkResearch(pieces []sparsePieceResearch, index int64) int64 {
	position := sort.Search(len(pieces), func(i int) bool { return pieces[i].last >= index })
	if position == len(pieces) || pieces[position].first > index {
		return 0
	}
	return pieces[position].digit
}

func sparseFloorResearch(pieces []sparsePieceResearch, exponent int64, work *sparseWorkResearch) (*big.Int, bool, error) {
	index, remainder := sparsePositionResearch(exponent)
	high := pieces[len(pieces)-1].last
	// From the 16-digit common denominator and normalized output, at most
	// three base-1e9 chunks can intersect either final rounding query.
	if high-index > 2 || high < index {
		return nil, false, fmt.Errorf("unproved rounding query width %d", high-index)
	}
	value, base := new(big.Int), big.NewInt(sparseBaseResearch)
	for at := high; at >= index; at-- {
		value.Mul(value, base).Add(value, big.NewInt(sparseChunkResearch(pieces, at)))
		work.reads++
	}
	scale := int64(1)
	for i := int64(0); i < remainder; i++ {
		scale *= 10
	}
	sticky := pieces[0].first < index || sparseChunkResearch(pieces, index)%scale != 0
	return value.Quo(value, big.NewInt(scale)), sticky, nil
}

func sparseSumResearch(values []Decimal) (Decimal, sparseWorkResearch, error) {
	work := sparseWorkResearch{}
	// The exact sum of no inputs is canonical zero; one valid input is
	// already its exactly represented sum. Preserve canonical positive zero
	// and validate before returning, including diagnostic/unnormalized inputs.
	if len(values) == 0 {
		return Zero, work, nil
	}
	if len(values) == 1 {
		value := values[0]
		if !value.IsStateValue() {
			return NaN, work, nil
		}
		if value.mantissa == 0 {
			return Zero, work, nil
		}
		return value, work, nil
	}
	positive, negative := map[int64]*big.Int{}, map[int64]*big.Int{}
	// Scratch storage is call-local. Map entries still own distinct Ints;
	// Add copies each digit before the next in-place QuoRem overwrites it.
	var coefficient, digit, scale big.Int
	base := big.NewInt(sparseBaseResearch)
	for _, value := range values {
		if !value.IsStateValue() {
			return NaN, work, nil
		}
		if value.mantissa == 0 {
			continue
		}
		bits := math.Float64bits(value.mantissa)
		binaryExponent := int((bits>>52)&2047) - 1023
		if binaryExponent < 0 || binaryExponent > 3 {
			return NaN, work, fmt.Errorf("invalid normalized significand")
		}
		coefficient.SetUint64((bits & ((1 << 52) - 1)) | (1 << 52))
		coefficient.Lsh(&coefficient, uint(binaryExponent))
		index, remainder := sparsePositionResearch(value.exponent)
		// Euclidean remainder is0..8, so this exact power is at most1e8
		// and fits int64 independently of the input count/exponent distance.
		power := int64(1)
		for i := int64(0); i < remainder; i++ {
			power *= 10
		}
		coefficient.Mul(&coefficient, scale.SetInt64(power))
		target := positive
		if value.mantissa < 0 {
			target = negative
		}
		for coefficient.Sign() != 0 {
			coefficient.QuoRem(&coefficient, base, &digit)
			if digit.Sign() != 0 {
				if target[index] == nil {
					target[index] = new(big.Int)
				}
				target[index].Add(target[index], &digit)
			}
			index++
			work.chunks++
		}
	}
	p, n := sparseNormalizeResearch(positive, &work), sparseNormalizeResearch(negative, &work)
	sign := sparseCompareResearch(p, n)
	if sign == 0 {
		return Zero, work, nil
	}
	if sign < 0 {
		p, n = n, p
	}
	pieces, err := sparseSubtractResearch(p, n, &work)
	if err != nil || len(pieces) == 0 {
		return NaN, work, fmt.Errorf("invalid sparse difference: %v", err)
	}
	last := pieces[len(pieces)-1]
	leading := 9*last.last + int64(len(strconv.FormatInt(last.digit, 10))) - 1
	exponent := leading - 15
	integer, _, err := sparseFloorResearch(pieces, exponent, &work)
	if err != nil {
		return NaN, work, err
	}
	denominator := big.NewInt(1 << 52)
	if integer.Cmp(denominator) < 0 {
		exponent--
		integer, _, err = sparseFloorResearch(pieces, exponent, &work)
		if err != nil {
			return NaN, work, err
		}
	}
	if exponent < -maxExponent || exponent > maxExponent {
		return NaN, work, nil
	}
	binaryExponent := 0
	var threshold big.Int
	for binaryExponent < 3 && integer.Cmp(threshold.Lsh(denominator, uint(binaryExponent+1))) >= 0 {
		binaryExponent++
	}
	integer, sticky, err := sparseFloorResearch(pieces, exponent-1, &work)
	if err != nil {
		return NaN, work, err
	}
	divisor := big.NewInt(10 << binaryExponent)
	quotient, remainder := new(big.Int), new(big.Int)
	quotient.QuoRem(integer, divisor, remainder)
	comparison := remainder.Lsh(remainder, 1).Cmp(divisor)
	if comparison > 0 || comparison == 0 && (sticky || quotient.Bit(0) != 0) {
		quotient.Add(quotient, big.NewInt(1))
	}
	if quotient.Cmp(denominator) < 0 || quotient.Cmp(threshold.Lsh(denominator, 1)) > 0 {
		return NaN, work, fmt.Errorf("unproved significand width")
	}
	result := New(float64(sign)*math.Ldexp(float64(quotient.Uint64()), binaryExponent-52), exponent)
	if !result.IsStateValue() {
		return NaN, work, nil
	}
	return result, work, nil
}

func TestSparseAggregationResearchIdentityPaths(t *testing.T) {
	got, _, err := sparseSumResearch(nil)
	if err != nil || math.Float64bits(got.mantissa) != 0 || got.exponent != 0 {
		t.Fatal("empty sum is not canonical positive zero", got, err)
	}
	for _, value := range []Decimal{
		Zero, {mantissa: math.Copysign(0, -1)}, One, One.Neg(),
		New(math.Nextafter(1, 2), 0), New(math.Nextafter(10, 0), maxExponent),
		New(-math.Nextafter(10, 0), -maxExponent), New(1, -maxExponent),
		NaN, Inf, NegInf, {mantissa: 10}, {mantissa: 1, exponent: maxExponent + 1},
		{mantissa: 1, exponent: -maxExponent - 1}, {exponent: 1},
	} {
		// Zero padding deliberately forces the unchanged general accumulator.
		// At full-domain endpoints the literal identity is also checked; the
		// existing bounded rational population remains the independent oracle.
		want, _, generalErr := sparseSumResearch([]Decimal{value, Zero})
		got, _, err := sparseSumResearch([]Decimal{value})
		if err != nil || generalErr != nil || value.IsStateValue() &&
			(math.Float64bits(got.mantissa) != math.Float64bits(want.mantissa) || got.exponent != want.exponent) ||
			!value.IsStateValue() && (!got.IsNaN() || !want.IsNaN()) {
			t.Fatalf("identity differs from general sum: input(%g,e%d) got(%g,e%d) want(%g,e%d) errors%v/%v",
				value.mantissa, value.exponent, got.mantissa, got.exponent, want.mantissa, want.exponent, err, generalErr)
		}
		if value.IsStateValue() && value.mantissa != 0 &&
			(math.Float64bits(got.mantissa) != math.Float64bits(value.mantissa) || got.exponent != value.exponent) {
			t.Fatal("single valid input is not its exact represented value", value, got)
		}
		for _, inputs := range [][]Decimal{{value, Zero}, {Zero, value}} {
			result, _, err := sparseSumResearch(inputs)
			if err != nil || !value.IsStateValue() && !result.IsNaN() {
				t.Fatal("zero padding hides invalid input", value, result, err)
			}
		}
	}
}

func TestSparseAggregationResearchEdges(t *testing.T) {
	u := math.Ldexp(1, -52)
	for _, row := range []struct {
		name   string
		values []Decimal
		want   Decimal
	}{
		{"empty", nil, Zero},
		{"tiny cancellation", []Decimal{FromString("1e300"), FromString("-1e300"), FromString("1e-300")}, FromString("1e-300")},
		{"full-domain cancellation", []Decimal{New(1, maxExponent), New(-5, maxExponent-1), New(-5, maxExponent-1), New(1, -maxExponent)}, New(1, -maxExponent)},
		{"full-domain borrow", []Decimal{New(1, maxExponent), New(-1, -maxExponent)}, New(1, maxExponent)},
		{"even tie", []Decimal{One, New(1+5*u, -1), New(-1, -1)}, One},
		{"remote above tie", []Decimal{One, New(1+5*u, -1), New(-1, -1), New(1, -maxExponent)}, New(1+u, 0)},
		{"remote below tie", []Decimal{One, New(1+5*u, -1), New(-1, -1), New(-1, -maxExponent)}, One},
		{"odd tie", []Decimal{New(1+u, 0), New(1+5*u, -1), New(-1, -1)}, New(1+2*u, 0)},
		{"odd below tie", []Decimal{New(1+u, 0), New(1+5*u, -1), New(-1, -1), New(-1, -maxExponent)}, New(1+u, 0)},
		{"negative remote tie", []Decimal{One.Neg(), New(-1-5*u, -1), New(1, -1), New(-1, -maxExponent)}, New(-1-u, 0)},
		{"recoverable overflow", []Decimal{New(9, maxExponent), New(9, maxExponent), New(-9, maxExponent)}, New(9, maxExponent)},
		{"net overflow", []Decimal{New(9, maxExponent), New(9, maxExponent)}, NaN},
		{"net underflow", []Decimal{New(1+u, -maxExponent), New(-1, -maxExponent)}, NaN},
		{"invalid", []Decimal{NaN}, NaN},
	} {
		t.Run(row.name, func(t *testing.T) {
			got, work, err := sparseSumResearch(row.values)
			if err != nil || !(row.want.IsNaN() && got.IsNaN()) && !got.Eq(row.want) {
				t.Fatalf("got(%g,e%d) want(%g,e%d) work%+v err%v", got.mantissa, got.exponent, row.want.mantissa, row.want.exponent, work, err)
			}
			t.Logf("work%+v", work)
		})
	}
}

// An independent bounded dense rational oracle: no sparse coefficient/chunk,
// subtraction, sticky-bit or significand-rounding implementation is reused.
func sparseRationalOracleResearch(t *testing.T, values []Decimal) Decimal {
	t.Helper()
	total := new(big.Rat)
	for _, value := range values {
		if !value.IsStateValue() || value.exponent < -32 || value.exponent > 32 {
			t.Fatal("input outside declared dense-oracle population")
		}
		term := new(big.Rat).SetFloat64(value.mantissa)
		e := value.exponent
		if e < 0 {
			e = -e
		}
		power := new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(e), nil))
		if value.exponent < 0 {
			term.Quo(term, power)
		} else {
			term.Mul(term, power)
		}
		total.Add(total, term)
	}
	if total.Sign() == 0 {
		return Zero
	}
	sign := total.Sign()
	total.Abs(total)
	one, ten := big.NewRat(1, 1), big.NewRat(10, 1)
	exponent := int64(0)
	for total.Cmp(ten) >= 0 {
		total.Quo(total, ten)
		exponent++
	}
	for total.Cmp(one) < 0 {
		total.Mul(total, ten)
		exponent--
	}
	mantissa, _ := total.Float64()
	return New(float64(sign)*mantissa, exponent)
}

func TestSparseAggregationResearchRationalPopulation(t *testing.T) {
	seed := uint64(664)
	next := func() uint64 {
		seed ^= seed << 13
		seed ^= seed >> 7
		seed ^= seed << 17
		return seed
	}
	maxWork := sparseWorkResearch{}
	var witnesses strings.Builder
	for profile := 0; profile < 2048; profile++ {
		values := make([]Decimal, int(next()%17))
		for i := range values {
			mantissa := float64(100_000_000_000_000+next()%900_000_000_000_000) / 1e14
			exponent := int64(next()%65) - 32
			if next()&1 != 0 {
				mantissa = -mantissa
			}
			values[i] = New(mantissa, exponent)
		}
		if profile%4 == 0 && len(values) != 0 {
			values = append(values, values[0].Neg())
		}
		before := append([]Decimal(nil), values...)
		want := sparseRationalOracleResearch(t, values)
		got, work, err := sparseSumResearch(values)
		if err != nil || !got.Eq(want) {
			t.Fatalf("profile%d got(%g,e%d), rational(%g,e%d), work%+v err%v", profile, got.mantissa, got.exponent, want.mantissa, want.exponent, work, err)
		}
		fmt.Fprintf(&witnesses, "%d:%d|", profile, len(values))
		for _, value := range values {
			fmt.Fprintf(&witnesses, "%016x:%d;", math.Float64bits(value.mantissa), value.exponent)
		}
		fmt.Fprintf(&witnesses, "=>%016x:%d\n", math.Float64bits(got.mantissa), got.exponent)
		for i := range values {
			if values[i] != before[i] {
				t.Fatal("candidate mutated an input")
			}
		}
		reversed := append([]Decimal(nil), values...)
		for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
			reversed[i], reversed[j] = reversed[j], reversed[i]
		}
		other, _, err := sparseSumResearch(reversed)
		if err != nil || !got.Eq(other) {
			t.Fatal("permutation changes candidate result", profile, err)
		}
		// Each coefficient contributes at most three chunks. A carry is at
		// most n, hence at most digits_base(n) extra adjacent positions per
		// occupied key. These bounds derive from representation, not timing.
		digits := 0
		for n := len(values); n > 0; n /= int(sparseBaseResearch) {
			digits++
		}
		if work.chunks > 3*len(values) || work.normalize > work.chunks*(1+digits) || work.events > work.normalize || work.runs > work.events || work.reads > 9 {
			t.Fatal("derived work bound violated", profile, work)
		}
		if work.normalize > maxWork.normalize {
			maxWork = work
		}
	}
	t.Logf("complete2048 seed664: exact rational/raw match, reverse/input controls; largest normalized-work row%+v", maxWork)
	t.Logf("input/output witness sha256 %x", sha256.Sum256([]byte(witnesses.String())))
	if fmt.Sprintf("%x", sha256.Sum256([]byte(witnesses.String()))) != "ef6cdc7626f91ff237961488f84f07a022b5ac746542d53dfb68fb258768047f" {
		t.Fatal("declared input/output population changed")
	}
}

func TestSparseAggregationResearchOriginalPair(t *testing.T) {
	var fixture struct {
		Fixed    []struct{ Rate string }
		Variable struct {
			Lower string `json:"lower_rate"`
			Upper string `json:"upper_rate"`
		}
	}
	data, err := os.ReadFile("../../testdata/numeric-sum-order-research-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &fixture); err != nil || len(fixture.Fixed) != 4 {
		t.Fatal("invalid original pair", err)
	}
	for _, source := range []string{fixture.Variable.Lower, fixture.Variable.Upper} {
		values := []Decimal{FromString(source)}
		for _, fixed := range fixture.Fixed {
			values = append(values, FromString(fixed.Rate))
		}
		got, work, err := sparseSumResearch(values)
		if err != nil || got.String() != "1.01245957117e1" || !got.Eq(sparseRationalOracleResearch(t, values)) {
			t.Fatalf("original pair: result(%g,e%d), work%+v err%v", got.mantissa, got.exponent, work, err)
		}
	}
}

// Compatibility search, not an equivalence census: seek the first committed
// difference for four positive canonical card metrics and one canonical factor.
// A 20,000-profile guard is an experiment bound, not a gameplay-domain limit.
// These are admitted numeric fields, not yet a reachable minigame session.
func TestSparseAggregationResearchFourTermConsumerSearch(t *testing.T) {
	data, err := os.ReadFile("../../testdata/pitch/aggregation-compatibility-research-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var vector struct {
		Version int
		Seed    uint64 `json:"search_seed"`
		Profile int    `json:"search_profile"`
		Cards   []struct {
			Metric string `json:"base_metric"`
		}
		Factor    string `json:"card_factor"`
		Current   string `json:"current_valuation"`
		Candidate string `json:"candidate_valuation"`
	}
	if json.Unmarshal(data, &vector) != nil || vector.Version != 1 || vector.Seed != 665 || len(vector.Cards) != 4 {
		t.Fatal("invalid consumer counterexample")
	}
	seed := vector.Seed
	next := func() uint64 {
		seed ^= seed << 13
		seed ^= seed >> 7
		seed ^= seed << 17
		return seed
	}
	canonicalInput := func() Decimal {
		coefficient := 100_000_000_000 + next()%900_000_000_000
		exponent := int64(next()%7) - 3
		value, err := ParseCanonical(New(float64(coefficient)/1e11, exponent).String())
		if err != nil {
			t.Fatal(err)
		}
		return value
	}
	for profile := 0; profile < 20_000; profile++ {
		factor := canonicalInput()
		bases := []Decimal{canonicalInput(), canonicalInput(), canonicalInput(), canonicalInput()}
		terms := make([]Decimal, len(bases))
		for i, base := range bases {
			terms[i] = base.Mul(factor)
		}
		candidate, _, err := sparseSumResearch(terms)
		if err != nil {
			t.Fatal(err)
		}
		current := SumDeterministic(terms).Quantize(CanonicalSignificantDigits)
		candidate = candidate.Quantize(CanonicalSignificantDigits)
		if !current.IsStateValue() || !candidate.IsStateValue() {
			t.Fatal("invalid declared consumer profile", profile)
		}
		if !current.Eq(candidate) {
			if !candidate.Eq(sparseRationalOracleResearch(t, terms).Quantize(CanonicalSignificantDigits)) {
				t.Fatal("consumer disagreement does not match independent rational rounding", profile)
			}
			if profile != vector.Profile || factor.String() != vector.Factor || current.String() != vector.Current || candidate.String() != vector.Candidate {
				t.Fatal("retained consumer counterexample changed", profile, bases, factor, current, candidate)
			}
			for i, base := range bases {
				if base.String() != vector.Cards[i].Metric {
					t.Fatal("retained consumer base changed", i, base)
				}
			}
			t.Logf("first committed disagreement at profile%d: bases%v factor%s current%s candidate%s", profile, bases, factor, current, candidate)
			return
		}
	}
	t.Fatal("no disagreement found within declared search; do not infer equivalence")
}
