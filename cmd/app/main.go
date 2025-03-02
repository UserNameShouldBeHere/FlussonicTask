package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/UserNameShouldBeHere/FlussonicTask/internal/domain"
	"github.com/UserNameShouldBeHere/FlussonicTask/internal/handlers"
	"github.com/UserNameShouldBeHere/FlussonicTask/internal/pipeline"
	"github.com/UserNameShouldBeHere/FlussonicTask/internal/tasktracker"
)

func main() {
	var (
		backEndPort  int
		taskTimeout  int
		taskWorkers  uint
		rerunWorkers uint
		kafkaHost    string
	)

	flag.IntVar(&backEndPort, "p", 8080, "backend port")
	flag.IntVar(&taskTimeout, "t", 3000, "task timeout in milliseconds")
	flag.UintVar(&taskWorkers, "twrk", 10, "amount of workers for tasks")
	flag.UintVar(&rerunWorkers, "rwrk", 10, "amount of workers for rerunning tasks")
	flag.StringVar(&kafkaHost, "kh", "kafka", "kafka host (e.g. 'kafka' for Docker or 'localhost' for local runs)")

	flag.Parse()

	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(zapcore.DebugLevel),
		Development:      true,
		Encoding:         "console",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, err := config.Build()
	if err != nil {
		log.Fatal(err)
	}
	sugarLogger := logger.Sugar()

	taskTracker, err := tasktracker.NewTaskTracker()
	if err != nil {
		log.Fatal(err)
	}

	tasksCh := make(chan domain.Task, 100)
	p, err := pipeline.NewPipeline(
		uint16(taskWorkers),
		uint16(rerunWorkers),
		time.Millisecond*time.Duration(taskTimeout),
		tasksCh,
		taskTracker)
	if err != nil {
		log.Fatal(err)
	}

	go startPipeline(p, tasksCh, kafkaHost, sugarLogger)

	startServer(p, backEndPort, kafkaHost, sugarLogger)
}

func startServer(
	pipeline *pipeline.Pipeline,
	backEndPort int,
	kafkaHost string,
	logger *zap.SugaredLogger) {

	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":  kafkaHost,
		"enable.idempotence": true,
	})
	if err != nil {
		logger.Fatalln(err)
	}

	defer p.Close()

	tasksHandler, err := handlers.NewTasksHandler(p, pipeline, logger)
	if err != nil {
		logger.Fatalf("error in tasks handler initialization: %v", err)
	}

	router := http.NewServeMux()

	router.HandleFunc("GET /job/all", tasksHandler.GetAllTasks)
	router.HandleFunc("GET /job/{id}/status", tasksHandler.GetTask)
	router.HandleFunc("POST /job", tasksHandler.AddTask)
	router.HandleFunc("POST /job/{id}/cancel", tasksHandler.CancelTask)

	server := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf(":%d", backEndPort),
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
	}

	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		<-sigint
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			logger.Errorf("Server shutdown error: %v", err)
		}
	}()

	logger.Infof("Starting server at %s%s", "localhost", fmt.Sprintf(":%d", backEndPort))

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		logger.Fatalln(err)
	}

	<-stopped

	logger.Infoln("Server stopped")
}

func startPipeline(
	pipeline *pipeline.Pipeline,
	tasksCh chan domain.Task,
	kafkaHost string,
	logger *zap.SugaredLogger) {

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": kafkaHost,
		"group.id":          "myGroup",
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		logger.Fatalln(err)
	}

	err = c.SubscribeTopics([]string{"tasks"}, nil)
	if err != nil {
		logger.Fatalln(err)
	}
	defer c.Close()

	errorsCh := pipeline.Run()

	go func() {
		for err := range errorsCh {
			logger.Infoln(err)
		}
	}()

	go func() {
		currentId := 0

		for {
			msg, err := c.ReadMessage(time.Second)
			if err == nil {
				var message domain.TaskMessage
				err := json.Unmarshal(msg.Value, &message)
				if err != nil {
					logger.Errorf("error in unmarshalling: %v", err)
				}

				ctx, cancel := context.WithTimeout(context.Background(), message.Timeout)

				task := domain.Task{
					Id:     uint16(currentId),
					Ttl:    2,
					Ctx:    ctx,
					Cancel: cancel,
					Task: func() domain.Result {
						resultCh := make(chan domain.Result, 1)
						defer close(resultCh)

						<-time.After(time.Duration(rand.Int63n(int64(5 * time.Second))))

						return domain.Result{
							Result: message.Data,
							Err:    nil,
						}
					},
				}
				tasksCh <- task
				currentId++
			} else if !err.(kafka.Error).IsTimeout() {
				logger.Errorf("Consumer error: %v (%v)", err, msg)
			}
		}
	}()

	pipeline.Wait()
}
