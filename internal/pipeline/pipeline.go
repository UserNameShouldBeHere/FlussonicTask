package pipeline

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/UserNameShouldBeHere/FlussonicTask/internal/domain"
)

type Pipeline struct {
	taskWorkersCount  uint16
	rerunWorkersCount uint16

	wg      *sync.WaitGroup
	rerunWg *sync.WaitGroup

	tasksCh  chan domain.Task
	rerunCh  chan domain.Task
	errorsCh chan error

	commonTimeout time.Duration

	worksForRerun uint16
}

func NewPipeline(taskWorkersCount uint16, rerunWorkersCount uint16, commonTimeout time.Duration, tasksCh chan domain.Task) (*Pipeline, error) {
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
		commonTimeout:     commonTimeout,
		worksForRerun:     0,
	}

	return pipeline, nil
}

func (p *Pipeline) Run() <-chan error {
	for range p.taskWorkersCount {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()

			for newTask := range p.tasksCh {

				ctx, _ := context.WithTimeout(newTask.Ctx, p.commonTimeout)
				newTask.Ctx = ctx

				log.Printf("current running: %d\n", newTask.Id)
				err := newTask.Run()
				if err != nil {
					if newTask.Ttl <= 0 {
						p.errorsCh <- fmt.Errorf("task №%d failed with error: %v", newTask.Id, err)
					} else {
						p.errorsCh <- fmt.Errorf("task №%d(ttl: %d) -> rerun", newTask.Id, newTask.Ttl)
						newTask.Ttl--
						p.rerunCh <- newTask
						p.worksForRerun++
					}
				} else {
					p.errorsCh <- fmt.Errorf("task №%d done", newTask.Id)
				}
			}
		}()
	}

	for range p.rerunWorkersCount {
		p.rerunWg.Add(1)
		go func() {
			defer p.rerunWg.Done()

			for newTask := range p.rerunCh {

				ctx, _ := context.WithTimeout(newTask.Ctx, p.commonTimeout)
				newTask.Ctx = ctx

				log.Printf("current rerunning: %d\n", newTask.Id)
				err := newTask.Run()
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
					p.errorsCh <- fmt.Errorf("task №%d done", newTask.Id)
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
	log.Println("Done")
}
