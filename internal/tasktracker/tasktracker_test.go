package tasktracker

import (
	"testing"
)

func TestGetStatus(t *testing.T) {
	taskTracker, err := NewTaskTracker()
	if err != nil {
		t.Error(err)
	}

	status := taskTracker.GetStatus(0)

	if status != NotInList {
		t.Errorf("expected: %v; got: %v", NotInList, status)
	}

	testData := []struct {
		TestName  string
		TaskId    uint16
		NewStatus Status
	}{
		{
			"done status",
			0,
			Done,
		},
		{
			"failed status",
			0,
			Failed,
		},
		{
			"canceled status",
			0,
			Canceled,
		},
		{
			"executing status",
			0,
			Executing,
		},
		{
			"waiting status",
			0,
			Waiting,
		},
	}

	for _, testCase := range testData {
		t.Run(testCase.TestName, func(t *testing.T) {
			taskTracker.UpdateStatus(testCase.TaskId, testCase.NewStatus)

			status = taskTracker.GetStatus(testCase.TaskId)

			if status != testCase.NewStatus {
				t.Errorf("expected: %v; got: %v", testCase.NewStatus, status)
			}
		})
	}
}

func TestGetAllStatuses(t *testing.T) {
	taskTracker, err := NewTaskTracker()
	if err != nil {
		t.Error(err)
	}

	testData := []struct {
		TaskId    uint16
		NewStatus Status
	}{
		{
			1,
			Done,
		},
		{
			2,
			Failed,
		},
		{
			3,
			Canceled,
		},
		{
			4,
			Executing,
		},
		{
			5,
			Waiting,
		},
	}

	for _, testCase := range testData {
		taskTracker.UpdateStatus(testCase.TaskId, testCase.NewStatus)
	}

	statuses := taskTracker.GetAllStatuses()

	for _, testCase := range testData {
		status, ok := statuses[testCase.TaskId]
		if !ok {
			t.Errorf("task №%d not in list", testCase.TaskId)
		}
		if status != testCase.NewStatus {
			t.Errorf("expected: %v; got: %v", testCase.NewStatus, status)
		}
	}
}

func TestUpdateStatus(t *testing.T) {
	taskTracker, err := NewTaskTracker()
	if err != nil {
		t.Error(err)
	}

	taskId := uint16(0)
	newStatus := Done

	status := taskTracker.GetStatus(taskId)

	if status != NotInList {
		t.Errorf("expected: %v; got: %v", NotInList, status)
	}

	taskTracker.UpdateStatus(taskId, newStatus)

	status = taskTracker.GetStatus(taskId)

	if status != newStatus {
		t.Errorf("expected: %v; got: %v", newStatus, status)
	}
}

func TestCancel(t *testing.T) {
	taskTracker, err := NewTaskTracker()
	if err != nil {
		t.Error(err)
	}

	taskId := uint16(0)

	taskTracker.Cancel(taskId)

	status := taskTracker.GetStatus(taskId)

	if status != Canceled {
		t.Errorf("expected: %v; got: %v", Canceled, status)
	}
}
