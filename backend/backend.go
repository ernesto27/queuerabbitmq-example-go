package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

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

func failOnError(err error, msg string, lastInsertId int64, db *sql.DB) {
	if err != nil {
		// Set db item no rabbit saved
		log.Println("Error on queue rabbit connection, set row ...")
		query := `UPDATE posts SET queue = (?) WHERE id = (?)`
		statement, err := db.Prepare(query)
		if err != nil {
			log.Fatalln(err.Error())
		}

		_, errExec := statement.Exec(0, lastInsertId)
		if errExec != nil {
			log.Fatalln(err.Error())
		}
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {

	// SQLITE INSERT POST
	sqliteDatabase, _ := sql.Open("sqlite3", "../sqlite-database.db") // Open the created SQLite File
	defer sqliteDatabase.Close()
	log.Println("Inserting post record...")
	insertStudentSQL := `INSERT INTO posts(email, body, image, queue) VALUES (?, ?, ?, ?)`
	statement, err := sqliteDatabase.Prepare(insertStudentSQL) // Prepare statement.
	// This is good to avoid SQL injections
	if err != nil {
		fmt.Println("Prepare fail")
		log.Fatalln(err.Error())
	}

	email := os.Getenv("TO_EMAIL")
	bodyPost := "Body of the post"
	image := "postimages/image.jpg"
	queue := 1

	res, err := statement.Exec(email, bodyPost, image, queue)
	if err != nil {
		fmt.Println("INSERT FAILS")
		log.Fatalln(err.Error())
	}

	lastInsertId, errLastInsert := res.LastInsertId()
	if errLastInsert != nil {
		log.Fatalln(err.Error())
	}

	// PUB SUB EXAMPLE
	conn, err := amqp.Dial(os.Getenv("RABBIT_CONN"))
	failOnError(err, "Failed to connect to RabbitMQ", lastInsertId, sqliteDatabase)
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel", lastInsertId, sqliteDatabase)
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
	failOnError(err, "Failed to declare an exchange", lastInsertId, sqliteDatabase)

	body := `{"email": "` + email + `", "image": "` + image + `"}`
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
	failOnError(err, "Failed to publish a message", lastInsertId, sqliteDatabase)

	log.Printf(" [x] Sent %s", body)

}
