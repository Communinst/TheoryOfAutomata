package fsm

type State int

// Q
// MX
// main is North-South, side is East-West
// X - first colour's name letter

const (
	NSG_STRAIGHT_RIGHT State = iota //	Means EWR
	NSY_STRAIGHT_RIGHT

	NSG_LEFT // EWG_RIGHT
	NSY_LEFT

	EWG_STRAIGHT_RIGHT // NSR
	EWY_STRAIGHT_RIGHT

	EWG_LEFT // NSG_RIGTH
	EWY_LEFT

	totalStates
	inital
)

var states = [totalStates]string{
	"North-South: Green forward and right turn.\n",
	"North-South: Yellow forward and right turn.\n",
	"North-South: Green left turn.\n East-West: Green right turn.\n",
	"North-South: Yellow left turn.\n East-West: Yellow right turn.\n",
	"East-West: Green forward and right turn.\n",
	"East-West: Yellow forward and right turn.\n",
	"East-West: Green left turn.\n North-South: Green right turn.\n",
	"East-West: Yellow left turn.\n North-South: Yellow right turn.\n",
}

func (s State) String() string {
	if s < 0 || s >= totalStates {
		return "Oops, something's broken. Call 911😭.\n"
	}
	if s == inital {
		return "Initial state. Please, do smth."
	}
	return states[s]
}
