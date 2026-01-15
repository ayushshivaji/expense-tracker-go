package main

import (
	"database/sql"
	"encoding/hex"
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
var userSchema string = `
CREATE TABLE IF NOT EXISTS users (
username VARCHAR(255) PRIMARY KEY,
password_hash VARCHAR(255) 
);`

func ensureDbSchema(db *sql.DB) {
	rows, err := db.Exec(dbSchema)
	if err != nil {
		log.Fatal("Not able to create schema", err)
		return
	}
	_, errUserSchema := db.Exec(userSchema)
	if errUserSchema != nil {
		log.Fatal("Not able to create schema", errUserSchema)
		return
	}
	fmt.Println("Successfully created all schema", rows)
}
func createDB() *sql.DB {
	connStr := "postgres://expense_user:expense_password@cloud-shell:8432/expense_database?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	ensureDbSchema(db)
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

func checkIfValidLogin(db *sql.DB, username string, hash [16]byte) bool {
	var storedHash []byte
	err := db.QueryRow(
		`SELECT password_hash FROM users WHERE username = $1`,
		username,
	).Scan(&storedHash)

	if err != nil {
		if err == sql.ErrNoRows {
			return false
		}
		log.Println("Database error:", err)
		return false
	}
	return string(storedHash) == string(hex.EncodeToString(hash[:]))
}

func addUser(db *sql.DB, username string, hash [16]byte) bool {
	fmt.Print(username, hex.EncodeToString(hash[:]))
	_, err := db.Exec(
		`INSERT INTO users (username, password_hash) VALUES ($1, $2)`,
		username,
		hex.EncodeToString(hash[:]),
	)
	if err == nil {
		fmt.Print("User added")
	} else {
		fmt.Print("User not added")
	}
	return err == nil
}
