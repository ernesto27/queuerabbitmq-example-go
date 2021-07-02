package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
	"project/structs"

	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/disintegration/imaging"
)

func init() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	conn, err := amqp.Dial(os.Getenv("RABBIT_CONN"))
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"checkout2", // name
		"fanout",    // type
		true,        // durable
		false,       // auto-deleted
		false,       // internal
		false,       // no-wait
		nil,         // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	q, err := ch.QueueDeclare(
		"imagequeue", // name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		q.Name,      // queue name
		"",          // routing key
		"checkout2", // exchange
		false,
		nil)
	failOnError(err, "Failed to bind a queue")

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
			log.Printf("Received from image service: %s", d.Body)
			message := structs.Message{}
			err := json.Unmarshal([]byte(string(d.Body)), &message)
			if err != nil {
				fmt.Println(err.Error())
			}
			saveImage(message)
		}

	}()

	log.Printf(" [*] Waiting for messages new. To exit press CTRL+C")
	<-forever
}

func saveImage(message structs.Message) {
	img, err := imaging.Open("../" + message.Image)
	if err != nil {
		panic(err)
	}
	thumb := imaging.Thumbnail(img, 200, 200, imaging.Lanczos)
	dst := imaging.New(200, 200, color.NRGBA{0, 0, 0, 0})
	dst = imaging.Paste(dst, thumb, image.Pt(0, 0))
	err = imaging.Save(dst, "../postimages/thumb.jpg")
	if err != nil {
		log.Fatalf("failed to save image: %v", err)
	}
	fmt.Println("Image saved")
}
