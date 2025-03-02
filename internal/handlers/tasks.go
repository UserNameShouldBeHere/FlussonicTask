package handlers

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/UserNameShouldBeHere/FlussonicTask/internal/domain"
	"github.com/UserNameShouldBeHere/FlussonicTask/internal/pipeline"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.uber.org/zap"
)

type TasksHandler struct {
	p          *kafka.Producer
	pipeline   *pipeline.Pipeline
	tasksTopic string
	logger     *zap.SugaredLogger
}

func NewTasksHandler(
	p *kafka.Producer,
	pipeline *pipeline.Pipeline,
	logger *zap.SugaredLogger) (*TasksHandler, error) {

	tasksHandler := &TasksHandler{
		p:          p,
		pipeline:   pipeline,
		tasksTopic: "tasks",
		logger:     logger,
	}

	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					logger.Errorf("Delivery failed: %v", ev.TopicPartition)
				} else {
					logger.Infof("Delivered message to %v", ev.TopicPartition)
				}
			}
		}
	}()

	return tasksHandler, nil
}

func (h *TasksHandler) AddTask(w http.ResponseWriter, req *http.Request) {
	message := domain.TaskMessage{
		Timeout: time.Second * 3,
		Data:    "hello",
	}
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		err = WriteResponse(w, ResponseData{
			Status: 500,
			Data:   ErrorResponse{Errors: err.Error()},
		})
		if err != nil {
			h.logger.Errorf("error at writing response: %v", err)
		}
	}

	err = h.p.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &h.tasksTopic, Partition: kafka.PartitionAny},
		Value:          jsonMessage,
	}, nil)
	if err != nil {
		err = WriteResponse(w, ResponseData{
			Status: 500,
			Data:   ErrorResponse{Errors: err.Error()},
		})
		h.logger.Errorf("error at writing response: %v", err)
	}

	err = WriteResponse(w, ResponseData{
		Status: 200,
		Data:   nil,
	})
	h.logger.Errorf("error at writing response: %v", err)
}

func (h *TasksHandler) CancelTask(w http.ResponseWriter, req *http.Request) {
	id, err := strconv.ParseUint(req.PathValue("id"), 10, 64)
	if err != nil {
		err = WriteResponse(w, ResponseData{
			Status: 500,
			Data:   ErrorResponse{Errors: err.Error()},
		})
		h.logger.Errorf("error at writing response: %v", err)
	}

	h.pipeline.Cancel(uint16(id))

	err = WriteResponse(w, ResponseData{
		Status: 200,
		Data:   nil,
	})
	h.logger.Errorf("error at writing response: %v", err)
}

type StatusResponse struct {
	Id     uint16 `json:"id"`
	Status string `json:"status"`
}

func (h *TasksHandler) GetTask(w http.ResponseWriter, req *http.Request) {
	id, err := strconv.ParseUint(req.PathValue("id"), 10, 64)
	if err != nil {
		err = WriteResponse(w, ResponseData{
			Status: 500,
			Data:   ErrorResponse{Errors: err.Error()},
		})
		h.logger.Errorf("error at writing response: %v", err)
	}

	status := h.pipeline.GetTaskStatus(uint16(id))

	err = WriteResponse(w, ResponseData{
		Status: 200,
		Data: StatusResponse{
			Id:     uint16(id),
			Status: ConvertStatusToStr(status),
		},
	})
	h.logger.Errorf("error at writing response: %v", err)
}

type AllStatusesResponse struct {
	Tasks []StatusResponse `json:"tasks"`
}

func (h *TasksHandler) GetAllTasks(w http.ResponseWriter, req *http.Request) {
	tasks := []StatusResponse{}

	statuses := h.pipeline.GetAllTaskStatuses()
	for id, status := range statuses {
		tasks = append(tasks, StatusResponse{
			Id:     id,
			Status: ConvertStatusToStr(status),
		})
	}

	sort.Slice(tasks, func(i, j int) bool { return tasks[i].Id < tasks[j].Id })

	err := WriteResponse(w, ResponseData{
		Status: 200,
		Data: AllStatusesResponse{
			Tasks: tasks,
		},
	})
	h.logger.Errorf("error at writing response: %v", err)
}
