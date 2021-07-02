package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"project/structs"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	amqp "github.com/rabbitmq/amqp091-go"
)

func init() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	fmt.Println("Queue resend")

	// Get all post set by 0 queue error
	sqliteDatabase, _ := sql.Open("sqlite3", "../sqlite-database.db") // Open the created SQLite File
	defer sqliteDatabase.Close()

	row, err := sqliteDatabase.Query("SELECT id, email, image FROM posts WHERE queue = 0")
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()

	var messages []structs.Message
	for row.Next() {
		var id int
		var email string
		var image string
		row.Scan(&id, &email, &image)
		log.Println("Post: ", id, email, image)
		m := structs.Message{Id: id, Email: email, Image: image}
		messages = append(messages, m)
	}

	// Send messages to queue again
	conn, err := amqp.Dial(os.Getenv("RABBIT_CONN"))
	if err != nil {
		log.Fatalf("%s: %s", "Failed to connect to RabbitMQ", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("%s: %s", "Failed to open a channel", err)
	}
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
	if err != nil {
		log.Fatalf("%s: %s", "Failed to declare an exchange", err)
	}

	for _, m := range messages {
		body := `{"email": "` + m.Email + `", "image": "` + m.Image + `"}`
		err = ch.Publish(
			"checkout2", // exchange
			"",          // routing key
			false,       // mandatory
			false,       // immediate
			amqp.Publishing{
				DeliveryMode: amqp.Persistent,
				ContentType:  "application/json",
				Body:         []byte(body),
			})

		if err != nil {
			log.Fatalf("%s: %s", "Failed to publish a message", err)
		}

		// Update post mark queue success
		query := `UPDATE posts SET queue = (?) WHERE id = (?)`
		statement, err := sqliteDatabase.Prepare(query)
		if err != nil {
			log.Fatalln(err.Error())
		}

		_, errExec := statement.Exec(1, m.Id)
		if errExec != nil {
			log.Fatalln(err.Error())
		}
		log.Printf(" [x] Sent %s", body)
	}

}
