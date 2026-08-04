package achievements

type Observation struct {
	Facts      map[string]bool
	Counters   map[string]int64
	ExitCount  int64
	Generators map[string]int64
}

func Eligible(condition Condition, observation Observation) bool {
	switch condition.Kind {
	case ConditionFactPresent:
		return observation.Facts[condition.FactKind]
	case ConditionCounterAtLeast:
		return observation.Counters[condition.Counter] >= condition.Minimum
	case ConditionExitCountAtLeast:
		return observation.ExitCount >= condition.Minimum
	case ConditionOwnsGeneratorAtLeast:
		return observation.Generators[condition.GeneratorID] >= condition.Minimum
	case ConditionAllOf:
		for _, child := range condition.Children {
			if !Eligible(child, observation) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// NewlyEarned evaluates every definition against one pre-achievement snapshot.
// Returned definitions retain catalog byte order and never include a lifetime latch.
func (catalog *Catalog) NewlyEarned(runEarned, lifetimeEarned map[string]bool, run, career Observation) ([]Definition, error) {
	if catalog == nil || runEarned == nil || lifetimeEarned == nil {
		return nil, ErrInvalidCatalog
	}
	result := []Definition{}
	for _, definition := range catalog.Definitions {
		if runEarned[definition.ID] || lifetimeEarned[definition.ID] {
			continue
		}
		observation := run
		if definition.ConditionScope == ScopeCareer {
			observation = career
		}
		if Eligible(definition.Condition, observation) {
			result = append(result, cloneDefinition(definition))
		}
	}
	return result, nil
}
