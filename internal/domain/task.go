package domain

import (
	"context"
	"fmt"
)

type Result struct {
	Result interface{}
	Err    error
}

type TaskFunc func() Result

type Task struct {
	Id uint16

	Ttl uint8

	Task TaskFunc

	Ctx context.Context

	StopCh <-chan struct{}
}

func (t *Task) Run() error {
	resultCh := make(chan Result)
	go func() {
		resultCh <- t.Task()
	}()

	select {
	case <-t.Ctx.Done():
		return fmt.Errorf("task was canceled or took too much time")
	case res := <-resultCh:
		if res.Err != nil {
			return fmt.Errorf("error recieved while executing task")
		}
		return nil
	}
}
