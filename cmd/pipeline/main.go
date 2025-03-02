package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"

	"github.com/UserNameShouldBeHere/FlussonicTask/internal/domain"
	"github.com/UserNameShouldBeHere/FlussonicTask/internal/pipeline"
	"github.com/UserNameShouldBeHere/FlussonicTask/internal/tasktracker"
)

func main() {
	taskTracker, err := tasktracker.NewTaskTracker()
	if err != nil {
		log.Fatal(err)
	}

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost",
		"group.id":          "myGroup",
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		log.Fatal(err)
	}

	err = c.SubscribeTopics([]string{"tasks"}, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

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
		run := true
		currentId := 0

		for run {
			msg, err := c.ReadMessage(time.Second)
			if err == nil {
				task := domain.Task{
					Id:  uint16(currentId),
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
				currentId++
			} else if !err.(kafka.Error).IsTimeout() {
				// The client will automatically try to recover from all errors.
				// Timeout is not considered an error because it is raised by
				// ReadMessage in absence of messages.
				fmt.Printf("Consumer error: %v (%v)\n", err, msg)
			}
		}
		// for i := range 10 {
		// 	task := domain.Task{
		// 		Id:  uint16(i),
		// 		Ttl: 2,
		// 		Ctx: context.Background(),
		// 		Task: func() domain.Result {
		// 			resultCh := make(chan domain.Result, 1)
		// 			defer close(resultCh)

		// 			<-time.After(time.Duration(rand.Int63n(int64(5 * time.Second))))

		// 			return domain.Result{
		// 				Result: nil,
		// 				Err:    nil,
		// 			}
		// 		},
		// 	}
		// 	tasksCh <- task
		// }
		// close(tasksCh)
	}()

	p.Wait()
	fmt.Printf("Done, took: %v\n", time.Since(begin))
}
