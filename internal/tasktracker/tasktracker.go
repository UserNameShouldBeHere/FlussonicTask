package tasktracker

import "sync"

type Status int

const (
	NotInList Status = iota
	Done
	Failed
	Canceled
	Executing
	Waiting
)

type TaskTracker struct {
	statuses  map[uint16]Status
	forCancel map[uint16]struct{}
	mutex     *sync.Mutex
}

func NewTaskTracker() (*TaskTracker, error) {
	return &TaskTracker{
		statuses:  make(map[uint16]Status),
		forCancel: make(map[uint16]struct{}),
		mutex:     &sync.Mutex{},
	}, nil
}

func (t *TaskTracker) GetStatus(id uint16) Status {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	status, ok := t.statuses[id]
	if !ok {
		return 0
	}

	return status
}

func (t *TaskTracker) GetAllStatuses() map[uint16]Status {
	return t.statuses
}

func (t *TaskTracker) UpdateStatus(id uint16, newStatus Status) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.statuses[id] = newStatus
}

func (t *TaskTracker) Cancel(id uint16) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.statuses[id] = Canceled
	t.forCancel[id] = struct{}{}
}
