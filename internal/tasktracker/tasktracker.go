package tasktracker

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
}

func NewTaskTracker() (*TaskTracker, error) {
	return &TaskTracker{
		statuses:  make(map[uint16]Status),
		forCancel: make(map[uint16]struct{}),
	}, nil
}

func (t *TaskTracker) GetStatus(id uint16) Status {
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
	t.statuses[id] = newStatus
}

func (t *TaskTracker) Cancel(id uint16) {
	t.statuses[id] = Canceled
	t.forCancel[id] = struct{}{}
}
