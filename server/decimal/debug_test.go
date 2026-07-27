package decimal

import (
	"math"
	"strconv"
	"testing"
)

func TestDebug(t *testing.T) {
	a := FromString("-29.6871392847948")
	b := FromString("-9814193.13885174")
	ac, bc := a.components(), b.components()
	pa, _ := strconv.ParseFloat("-29.6871392847948", 64)
	pb, _ := strconv.ParseFloat("-9814193.13885174", 64)
	r := a.Add(b)
	rc := r.components()
	sum := ac.sign*ac.mag + bc.sign*bc.mag
	direct := FromFloat64(sum)
	t.Fatalf("am=%g ae=%d av=%.18g bits=%x parse=%.18g bits=%x bm=%g be=%d bv=%.18g bits=%x parse=%.18g bits=%x sum=%.18g formatted=%s directm=%.18g rm=%.18g re=%d rv=%.18g string=%s", a.mantissa, a.exponent, ac.sign*ac.mag, math.Float64bits(ac.sign*ac.mag), pa, math.Float64bits(pa), b.mantissa, b.exponent, bc.sign*bc.mag, math.Float64bits(bc.sign*bc.mag), pb, math.Float64bits(pb), sum, strconv.FormatFloat(sum, 'e', -1, 64), direct.mantissa, r.mantissa, r.exponent, rc.sign*rc.mag, r.String())
}
