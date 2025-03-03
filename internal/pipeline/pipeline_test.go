package pipeline

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/UserNameShouldBeHere/FlussonicTask/internal/domain"
	"github.com/UserNameShouldBeHere/FlussonicTask/internal/tasktracker"
)

func TestRunAndGetAllStatuses(t *testing.T) {
	taskTracker, err := tasktracker.NewTaskTracker()
	if err != nil {
		log.Fatal(err)
	}

	tasksCh := make(chan domain.Task, 100)
	p, err := NewPipeline(
		uint16(10),
		uint16(10),
		time.Second*3,
		tasksCh,
		taskTracker)
	if err != nil {
		log.Fatal(err)
	}

	errorsCh := p.Run()

	go func() {
		for err := range errorsCh {
			t.Errorf("unexpected error: %v", err)
		}
	}()

	go func() {
		for i := range 5 {
			task := domain.Task{
				Id:  uint16(i),
				Ttl: 2,
				Ctx: context.Background(),
				Task: func() domain.Result {
					resultCh := make(chan domain.Result, 1)
					defer close(resultCh)

					return domain.Result{
						Result: nil,
						Err:    nil,
					}
				},
			}
			tasksCh <- task
		}
		close(tasksCh)
	}()

	p.Wait()

	statuses := p.GetAllTaskStatuses()

	for _, status := range statuses {
		if status != tasktracker.Done {
			t.Errorf("expected: %v; got: %v", tasktracker.Done, status)
		}
	}
}

func TestCancelAndGetStatus(t *testing.T) {
	taskTracker, err := tasktracker.NewTaskTracker()
	if err != nil {
		log.Fatal(err)
	}

	tasksCh := make(chan domain.Task, 100)
	p, err := NewPipeline(
		uint16(10),
		uint16(10),
		time.Second*3,
		tasksCh,
		taskTracker)
	if err != nil {
		log.Fatal(err)
	}

	taskId := uint16(1)

	p.Cancel(taskId)

	errorsCh := p.Run()

	go func() {
		for err := range errorsCh {
			t.Errorf("unexpected error: %v", err)
		}
	}()

	go func() {
		for i := range 2 {
			task := domain.Task{
				Id:  uint16(i),
				Ttl: 2,
				Ctx: context.Background(),
				Task: func() domain.Result {
					resultCh := make(chan domain.Result, 1)
					defer close(resultCh)

					return domain.Result{
						Result: nil,
						Err:    nil,
					}
				},
			}
			tasksCh <- task
		}
		close(tasksCh)
	}()

	p.Wait()

	status := p.GetTaskStatus(0)
	if status != tasktracker.Done {
		t.Errorf("expected: %v; got: %v", tasktracker.Done, status)
	}

	status = p.GetTaskStatus(taskId)
	if status != tasktracker.Canceled {
		t.Errorf("expected: %v; got: %v", tasktracker.Canceled, status)
	}
}
