package pipeline

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/UserNameShouldBeHere/FlussonicTask/internal/domain"
	"github.com/UserNameShouldBeHere/FlussonicTask/internal/tasktracker"
)

type Pipeline struct {
	taskWorkersCount  uint16
	rerunWorkersCount uint16

	wg      *sync.WaitGroup
	rerunWg *sync.WaitGroup

	tasksCh  chan domain.Task
	rerunCh  chan domain.Task
	errorsCh chan error

	cancelFuncs map[uint16]context.CancelFunc

	commonTimeout time.Duration

	worksForRerun uint16

	taskTracker *tasktracker.TaskTracker
}

func NewPipeline(
	taskWorkersCount uint16,
	rerunWorkersCount uint16,
	commonTimeout time.Duration,
	tasksCh chan domain.Task,
	taskTracker *tasktracker.TaskTracker) (*Pipeline, error) {

	if taskWorkersCount < 1 || rerunWorkersCount < 1 {
		return nil, fmt.Errorf("number of workers can't be less than 1")
	}

	wg := &sync.WaitGroup{}
	rerunWg := &sync.WaitGroup{}

	pipeline := &Pipeline{
		taskWorkersCount:  taskWorkersCount,
		rerunWorkersCount: rerunWorkersCount,
		wg:                wg,
		rerunWg:           rerunWg,
		tasksCh:           tasksCh,
		rerunCh:           make(chan domain.Task, 100),
		errorsCh:          make(chan error, 100),
		cancelFuncs:       make(map[uint16]context.CancelFunc),
		commonTimeout:     commonTimeout,
		worksForRerun:     0,
		taskTracker:       taskTracker,
	}

	return pipeline, nil
}

func (p *Pipeline) Run() <-chan error {
	for range p.taskWorkersCount {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()

			for newTask := range p.tasksCh {
				ctx, cancel := context.WithTimeout(newTask.Ctx, p.commonTimeout)
				p.cancelFuncs[newTask.Id] = cancel

				newTask.Ctx = ctx

				err := p.handleTask(newTask)
				if err != nil {
					if newTask.Ttl <= 0 {
						p.errorsCh <- fmt.Errorf("task №%d failed with error: %v", newTask.Id, err)
					} else {
						p.errorsCh <- fmt.Errorf("task №%d(ttl: %d) -> rerun", newTask.Id, newTask.Ttl)
						newTask.Ttl--
						p.rerunCh <- newTask
						p.worksForRerun++
					}
				}
			}
		}()
	}

	for range p.rerunWorkersCount {
		p.rerunWg.Add(1)
		go func() {
			defer p.rerunWg.Done()

			for newTask := range p.rerunCh {
				ctx, cancel := context.WithTimeout(context.Background(), p.commonTimeout)
				p.cancelFuncs[newTask.Id] = cancel

				newTask.Ctx = ctx

				err := p.handleTask(newTask)
				if err != nil {
					if newTask.Ttl <= 0 {
						p.errorsCh <- fmt.Errorf("task №%d failed with error: %v", newTask.Id, err)
						p.worksForRerun--
					} else {
						p.errorsCh <- fmt.Errorf("task №%d(ttl: %d) -> rerun", newTask.Id, newTask.Ttl)
						newTask.Ttl--
						p.rerunCh <- newTask
					}
				} else {
					p.worksForRerun--
				}
			}
		}()
	}

	go func() {
		p.wg.Wait()
		for {
			if p.worksForRerun <= 0 {
				close(p.rerunCh)
				return
			}

			<-time.After(time.Millisecond * 500)
		}
	}()
	go func() {
		p.rerunWg.Wait()
		close(p.errorsCh)
	}()

	return p.errorsCh
}

func (p *Pipeline) Wait() {
	p.rerunWg.Wait()
}

func (p *Pipeline) Cancel(id uint16) {
	if cancel, ok := p.cancelFuncs[id]; ok {
		cancel()
	}
	p.taskTracker.Cancel(id)
}

func (p *Pipeline) GetTaskStatus(id uint16) tasktracker.Status {
	return p.taskTracker.GetStatus(id)
}

func (p *Pipeline) GetAllTaskStatuses() map[uint16]tasktracker.Status {
	return p.taskTracker.GetAllStatuses()
}

func (p *Pipeline) handleTask(task domain.Task) error {
	if status := p.taskTracker.GetStatus(task.Id); status == tasktracker.Canceled {
		return nil
	}

	p.taskTracker.UpdateStatus(task.Id, tasktracker.Executing)

	err := task.Run()
	if err != nil {
		if task.Ttl > 0 {
			p.taskTracker.UpdateStatus(task.Id, tasktracker.Waiting)
		} else {
			p.taskTracker.UpdateStatus(task.Id, tasktracker.Failed)
		}
	} else {
		p.taskTracker.UpdateStatus(task.Id, tasktracker.Done)
	}

	return err
}
