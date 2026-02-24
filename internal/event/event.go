package event

type EventType int

const (
	EvenTimeTicker EventType = iota
	EventCarSensor
	EventPedestrianButton
)

type Event struct {
	Type  EventType
	Metas string
}
