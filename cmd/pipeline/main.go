package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/UserNameShouldBeHere/FlussonicTask/internal/domain"
	"github.com/UserNameShouldBeHere/FlussonicTask/internal/pipeline"
)

func main() {
	tasksCh := make(chan domain.Task, 100)
	p, err := pipeline.NewPipeline(50, 10, time.Second*3, tasksCh)
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
		for i := range 5 {
			task := domain.Task{
				Id:  uint16(i),
				Ttl: 2,
				Ctx: context.Background(),
				Task: func() domain.Result {
					resultCh := make(chan domain.Result, 1)
					defer close(resultCh)

					<-time.After(time.Second * 3)

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
