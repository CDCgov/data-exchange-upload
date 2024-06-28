package event

type Event struct {
	ID string
}

type FileReadyEvent struct {
	Event
	Manifest      map[string]string
	DeliverTarget string
}

type Publisher interface {
	Publish(event FileReadyEvent)
}

type MemoryPublisher struct {
	FileReadyChannel chan FileReadyEvent
}

func (mp *MemoryPublisher) Publish(event FileReadyEvent) {
	mp.FileReadyChannel <- event
}
