package decimal

// SumGeometricSeries returns the price of count purchases after owned existing
// purchases: base * ratio^owned * (1-ratio^count) / (1-ratio).
func SumGeometricSeries(count int64, base, ratio Decimal, owned int64) Decimal {
	return base.
		Mul(ratio.Pow(float64(owned))).
		Mul(One.Sub(ratio.Pow(float64(count)))).
		Div(One.Sub(ratio))
}

// AffordGeometricSeries returns the maximum affordable purchase count as a
// Decimal integer. Returning Decimal avoids losing counts near JavaScript's
// safe-integer boundary.
func AffordGeometricSeries(cash, base, ratio Decimal, owned int64) Decimal {
	actualStart := base.Mul(ratio.Pow(float64(owned)))
	return cash.
		Div(actualStart).
		Mul(ratio.Sub(One)).
		Add(One).
		Log10().
		Div(ratio.Log10()).
		Floor()
}
