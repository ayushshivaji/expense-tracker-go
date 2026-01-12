package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

var dbSchema string = `
CREATE TABLE IF NOT EXISTS expenses (
	email_id VARCHAR(255) PRIMARY KEY,
	receiver_info VARCHAR(255),
	transaction_time timestamp,
	sender_information VARCHAR(255),
	amount VARCHAR(255)
);`

func ensureDbSchema(db *sql.DB) {
	fmt.Println("")
	rows, err := db.Exec(dbSchema)
	if err != nil {
		log.Fatal("Not able to create schema", err)
		return
	}
	fmt.Println("Successfully created schema", rows)
}
func createDB() *sql.DB {
	connStr := "postgres://expense_user:expense_password@cloud-shell:8432/expense_database?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	return (db)
}

func writeTransactionToDB(db *sql.DB, emailId string, receiverInfo string, senderInfo string, amount float64, txTime time.Time) error {
	_, err := db.Exec(
		`INSERT INTO expenses (email_id, receiver_info, sender_information, amount, transaction_time) 
         VALUES ($1, $2, $3, $4, $5)`,
		emailId,
		receiverInfo,
		senderInfo,
		amount,
		txTime,
	)
	return err
}
