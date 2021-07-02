package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	log.Println("Creating sqlite-database.db...")
	file, err := os.Create("sqlite-database.db") // Create SQLite file
	if err != nil {
		log.Fatal(err.Error())
	}
	file.Close()
	log.Println("sqlite-database.db created")

	sqliteDatabase, _ := sql.Open("sqlite3", "./sqlite-database.db") // Open the created SQLite File
	defer sqliteDatabase.Close()                                     // Defer Closing the database

	createPostTableSQL := `CREATE TABLE posts (
		"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,		
		"email" TEXT,
		"body" TEXT,
		"image" TEXT,
		"queue" INTEGER		
	  );` // SQL Statement for Create Table

	log.Println("Create student table...")
	statement, err := sqliteDatabase.Prepare(createPostTableSQL) // Prepare SQL Statement
	if err != nil {
		log.Fatal(err.Error())
	}
	statement.Exec() // Execute SQL Statements
	log.Println("student table created")
}
