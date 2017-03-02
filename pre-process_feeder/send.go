package main

import (
	"log"
	"github.com/streadway/amqp"
	"io/ioutil"
	"encoding/json"
	"os"
	"fmt"
	"regexp"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

type Configuration struct {
    spec_dir         string
    rabbit_host      string
    rabbit_port      string
    rabbit_user      string
    rabbit_password  string
}

func main() {
	file, _ := os.Open("conf.json")
	decoder := json.NewDecoder(file)
	var configuration map[string]string
	err := decoder.Decode(&configuration)
	if err != nil {
	  fmt.Println("error:", err)
	}

	conn, err := amqp.Dial("amqp://" + configuration["rabbit_user"] + ":" + configuration["rabbit_password"] + "@" + configuration["rabbit_host"] + ":" + configuration["rabbit_port"] + "/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"pre-process", // name
		true,    // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")

	files, _ := ioutil.ReadDir(configuration["spec_dir"])
	for _, f := range files {
		if f.IsDir() { continue }

		//re := regexp.MustCompile(".*\\.spec")
		matched, err := regexp.MatchString(".*\\.spec", f.Name())
		//matched := re.FindStringSubmatch(f.Name())
		if (matched) {
		body := f.Name()
		err = ch.Publish(
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(body),
			})
		log.Printf(" [x] Sent %s", body)
		failOnError(err, "Failed to publish a message")
		}
	}
}
