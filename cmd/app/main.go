package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

func main() {
	topic := "test"

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		producerConfig := &kafka.ConfigMap{
			"bootstrap.servers":     "localhost",
			"acks":                  "1",
			"request.required.acks": "1",
			"client.id":             "myProducer",
		}

		producer, err := kafka.NewProducer(producerConfig)
		if err != nil {
			fmt.Printf("Failed to create producer: %s\n", err)
			os.Exit(1)
		}
		defer producer.Close()
		fmt.Println("Producer initialized")

		for range 50 {
			err = producer.Produce(&kafka.Message{
				TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
				Value:          []byte("hello"),
			}, nil)
			if err != nil {
				log.Fatal(err)
			}
		}

		producer.Flush(15 * 1000)
	}()

	//!======================================

	go func() {
		consumerConfig := &kafka.ConfigMap{
			"bootstrap.servers": "localhost",
			"group.id":          "myGroup",
			"auto.offset.reset": "smallest"}

		consumer, err := kafka.NewConsumer(consumerConfig)
		if err != nil {
			fmt.Printf("Failed to create consumer: %s\n", err)
			os.Exit(1)
		}
		defer consumer.Close()
		fmt.Println("Consumer initialized")

		err = consumer.SubscribeTopics([]string{topic}, nil)
		if err != nil {
			log.Fatal(err)
		}

		for {
			msg, err := consumer.ReadMessage(time.Second)
			if err == nil {
				fmt.Printf("Message on %s: %s\n", msg.TopicPartition, string(msg.Value))
			} else if !err.(kafka.Error).IsTimeout() {
				// The client will automatically try to recover from all errors.
				// Timeout is not considered an error because it is raised by
				// ReadMessage in absence of messages.
				fmt.Printf("Consumer error: %v (%v)\n", err, msg)
			}
		}
	}()

	wg.Wait()
	<-time.After(time.Second * 10)
	// retrieved := consumer.Poll(0)
	// switch message := retrieved.(type) {
	// case *kafka.Message:
	// 	fmt.Println(string(message.Value))
	// case kafka.Error:
	// 	fmt.Printf("%% Error: %v\n", message)
	// default:
	// 	fmt.Printf("Ignored %v\n", message)
	// }
}
