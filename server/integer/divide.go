// Package integer owns checked nonnegative integer division outside combat
// modules, matching the client integer helper's domain and floor semantics.
package integer

import "errors"

var ErrInvalidDivision = errors.New("integer division requires nonnegative value and positive divisor")

func DivideFloor(value, divisor int64) (int64, error) {
	if value < 0 || divisor <= 0 {
		return 0, ErrInvalidDivision
	}
	return value / divisor, nil
}
