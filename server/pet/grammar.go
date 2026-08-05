package pet

type StatID string
type StatusBand string
type Mood string
type BehaviorState string
type BehaviorEvent string
type CareRejectionDetail string

const (
	StatHunger      StatID = "hunger"
	StatEnergy      StatID = "energy"
	StatCleanliness StatID = "cleanliness"
	StatAffection   StatID = "affection"

	StatusFloor  StatusBand = "floor"
	StatusLow    StatusBand = "low"
	StatusNormal StatusBand = "normal"
	StatusHigh   StatusBand = "high"

	MoodWithdrawn Mood = "withdrawn"
	MoodRestless  Mood = "restless"
	MoodNeutral   Mood = "neutral"
	MoodEngaged   Mood = "engaged"

	BehaviorIdle         BehaviorState = "idle"
	BehaviorCareResponse BehaviorState = "care_response"
	BehaviorActive       BehaviorState = "active"
	BehaviorResting      BehaviorState = "resting"

	EventGridTick     BehaviorEvent = "grid_tick"
	EventCareApplied  BehaviorEvent = "care_applied"
	EventCareRejected BehaviorEvent = "care_rejected"

	RejectionCooldown      CareRejectionDetail = "cooldown"
	RejectionIneligible    CareRejectionDetail = "ineligible"
	RejectionSaturated     CareRejectionDetail = "saturated"
	RejectionUnknownPet    CareRejectionDetail = "unknown_pet"
	RejectionUnknownAction CareRejectionDetail = "unknown_action"

	BehaviorQueueHardcap = 8
	BehaviorPRNGLabel    = "pet.behavior.v1"
)

var statIDs = [...]StatID{StatHunger, StatEnergy, StatCleanliness, StatAffection}
var statusBands = [...]StatusBand{StatusFloor, StatusLow, StatusNormal, StatusHigh}
var moods = [...]Mood{MoodWithdrawn, MoodRestless, MoodNeutral, MoodEngaged}
var behaviorStates = [...]BehaviorState{BehaviorIdle, BehaviorCareResponse, BehaviorActive, BehaviorResting}
var behaviorEvents = [...]BehaviorEvent{EventGridTick, EventCareApplied, EventCareRejected}
var careRejectionDetails = [...]CareRejectionDetail{
	RejectionCooldown, RejectionIneligible, RejectionSaturated, RejectionUnknownPet, RejectionUnknownAction,
}

func StatIDs() []StatID               { return append([]StatID(nil), statIDs[:]...) }
func StatusBands() []StatusBand       { return append([]StatusBand(nil), statusBands[:]...) }
func Moods() []Mood                   { return append([]Mood(nil), moods[:]...) }
func BehaviorStates() []BehaviorState { return append([]BehaviorState(nil), behaviorStates[:]...) }
func BehaviorEvents() []BehaviorEvent { return append([]BehaviorEvent(nil), behaviorEvents[:]...) }
func CareRejectionDetails() []CareRejectionDetail {
	return append([]CareRejectionDetail(nil), careRejectionDetails[:]...)
}

func ValidStatID(value StatID) bool               { return member(value, statIDs[:]) }
func ValidStatusBand(value StatusBand) bool       { return member(value, statusBands[:]) }
func ValidMood(value Mood) bool                   { return member(value, moods[:]) }
func ValidBehaviorState(value BehaviorState) bool { return member(value, behaviorStates[:]) }
func ValidBehaviorEvent(value BehaviorEvent) bool { return member(value, behaviorEvents[:]) }
func ValidCareRejectionDetail(value CareRejectionDetail) bool {
	return member(value, careRejectionDetails[:])
}
func ValidBehaviorQueueLength(length int) bool { return length >= 0 && length <= BehaviorQueueHardcap }

func member[T comparable](value T, values []T) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
