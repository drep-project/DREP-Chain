package event

type EventType int

const (
	StartSyncBlock EventType = 1
	StopSyncBlock  EventType = 2
)

type SyncBlockEvent struct {
	EventType EventType
}
