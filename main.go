package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s:\n%s\n", msg, err)
	}
}

func displayUsage() {
	fmt.Printf("Usage: %s (send|receive) [message]\n", os.Args[0])
	failOnError(errors.New("parsing: error parsing inputs"), "Please reformat your inputs")
}

func main() {
	// check arguments
	if len(os.Args) < 2 {
		displayUsage()
	}
	// load mode from arguments
	var body string
	mode := func(body *string) string {
		if len(os.Args) == 3 && os.Args[1] == "send" {
			*body = os.Args[2]
			return "send"
		} else {
			return "receive"
		}
	}(&body)

	// load env configs
	err := godotenv.Load()
	failOnError(err, "Failed to load configs from .env, please check .env.example")
	SERVER_URL := os.Getenv("SERVER_URL")
	SERVER_PORT := os.Getenv("SERVER_PORT")
	USER_NAME := os.Getenv("USER_NAME")
	USER_PASSWORD := os.Getenv("USER_PASSWORD")

	// create connection to RabbitMQ instance
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s/", USER_NAME, USER_PASSWORD, SERVER_URL, SERVER_PORT))
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	// create channel to RabbitMQ instance
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"queue1",
		false,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare a queue")

	if mode == "send" {
		err = ch.Publish(
			"",
			q.Name,
			false,
			false,
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(body),
			})
		failOnError(err, "Failed to publish a message")
		fmt.Printf("Sent %s on channel %s\n", body, q.Name)
	} else {
		msgs, err := ch.Consume(
			q.Name, // queue
			"",     // consumer
			true,   // auto-ack
			false,  // exclusive
			false,  // no-local
			false,  // no-wait
			nil,    // args
		)
		failOnError(err, "Failed to register a consumer")

		forever := make(chan bool)

		go func() {
			for d := range msgs {
				log.Printf("Received a message: %s", d.Body)
			}
		}()

		log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
		<-forever
	}
}
