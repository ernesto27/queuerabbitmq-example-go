package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"project/structs"

	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
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
		"emailqueue", // name
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
			log.Printf("Received: %s", d.Body)
			message := structs.Message{}
			err := json.Unmarshal([]byte(string(d.Body)), &message)
			if err != nil {
				fmt.Println(err.Error())
			}

			// Send email to user, save log on file
			// f, err := os.OpenFile("/tmp/emails.log",
			// 	os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			// if err != nil {
			// 	log.Println(err)
			// }
			// defer f.Close()
			// if _, err := f.WriteString(message.Email + "\n"); err != nil {
			// 	log.Println(err)
			// }

			// sendEmail(os.Getenv("FROM_EMAIL"), os.Getenv("FROM_PASSWORD"), message)
		}

	}()

	log.Printf(" [*] Waiting for messages new. To exit press CTRL+C")
	<-forever
}

func sendEmail(fromEmail string, fromPassword string, message structs.Message) {
	// Receiver email address.
	to := []string{
		message.Email,
	}

	// smtp server configuration.
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	header := make(map[string]string)
	header["From"] = fromEmail
	header["To"] = to[0]
	header["Subject"] = "This is a subject"
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=\"utf-8\""
	header["Content-Transfer-Encoding"] = "base64"

	mailMessage := ""
	for k, v := range header {
		mailMessage += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	mailMessage += "\r\n" + base64.StdEncoding.EncodeToString([]byte("this is the body"))

	// Authentication.
	auth := smtp.PlainAuth("", fromEmail, fromPassword, smtpHost)

	// Sending email.
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, fromEmail, to, []byte(mailMessage))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Email Sent Successfully!")
}
