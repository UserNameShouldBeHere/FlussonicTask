package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/UserNameShouldBeHere/FlussonicTask/internal/domain"
	"github.com/UserNameShouldBeHere/FlussonicTask/internal/pipeline"
	"github.com/UserNameShouldBeHere/FlussonicTask/internal/tasktracker"
)

func main() {
	taskTracker, err := tasktracker.NewTaskTracker()
	if err != nil {
		log.Fatal(err)
	}

	tasksCh := make(chan domain.Task, 100)
	p, err := pipeline.NewPipeline(10, 2, time.Second*3, tasksCh, taskTracker)
	if err != nil {
		log.Fatal(err)
	}

	errorsCh := p.Run()

	go func() {
		for err := range errorsCh {
			fmt.Println(err)
		}
	}()

	begin := time.Now()
	go func() {
		for i := range 10 {
			task := domain.Task{
				Id:  uint16(i),
				Ttl: 2,
				Ctx: context.Background(),
				Task: func() domain.Result {
					resultCh := make(chan domain.Result, 1)
					defer close(resultCh)

					<-time.After(time.Duration(rand.Int63n(int64(5 * time.Second))))

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
	fmt.Printf("Done, took: %v\n", time.Since(begin))
}
