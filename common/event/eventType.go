package event

type EventType int

const (
	StartSyncBlock = 1
	StopSyncBlock  = 2
)

type SyncBlockEvent struct {
	EventType EventType
}
