package combat

import "errors"

var ErrInvalidIdentity = errors.New("invalid combat identity")

type Runtime string

const (
	RuntimeContainer  Runtime = "container"
	RuntimeVM         Runtime = "vm"
	RuntimeBareMetal  Runtime = "bare_metal"
	RuntimeServerless Runtime = "serverless"
	RuntimeEdge       Runtime = "edge"
	RuntimeMainframe  Runtime = "mainframe"
)

// Identity is C1's temperament/runtime pair, not a species, catalog or wire
// envelope. Runtime may carry catalog tags later; it adds no Phase-0 chart rule.
type Identity struct {
	Temperament Temperament
	Runtime     Runtime
}

func NewIdentity(temperament Temperament, runtime Runtime) (Identity, error) {
	if _, valid := temperamentIndex(temperament); !valid {
		return Identity{}, ErrInvalidIdentity
	}
	switch runtime {
	case RuntimeContainer, RuntimeVM, RuntimeBareMetal, RuntimeServerless, RuntimeEdge, RuntimeMainframe:
		return Identity{Temperament: temperament, Runtime: runtime}, nil
	default:
		return Identity{}, ErrInvalidIdentity
	}
}

// ChartIdentities validates both axes even for a directly constructed Identity.
// The existing chart remains the single arithmetic authority.
func ChartIdentities(attacker, defender Identity) (ChartResult, error) {
	if _, err := NewIdentity(attacker.Temperament, attacker.Runtime); err != nil {
		return Neutral, err
	}
	if _, err := NewIdentity(defender.Temperament, defender.Runtime); err != nil {
		return Neutral, err
	}
	return Chart(attacker.Temperament, defender.Temperament)
}
