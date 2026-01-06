package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/jhillyerd/enmime"
)

func ParseMessage(message string) (string, time.Time, string, string, string) {
	data, err := base64.URLEncoding.DecodeString(message)
	if err != nil {
		log.Fatalf("Unable to decode message: %v", err)
	}
	msg, err := enmime.ReadEnvelope(bytes.NewReader(data))
	if err != nil {
		log.Fatalf("Unable to read envelope: %v", err)
	}
	var email_time time.Time
	var amount, merchant, creditCardNumber, transactionType string
	if strings.Contains(string(data), "Scapia") {
		fmt.Println("Parsing Mail from Scapia")
		merchant, email_time, amount, creditCardNumber, transactionType = parseScapiaMail(msg)
	} else {
		merchant, email_time, amount = genericParser(msg)
	}

	// fmt.Println(string(data))
	// fmt.Println(msg.GetHeader("Subject"))
	// fmt.Println(msg)
	return merchant, email_time, amount, creditCardNumber, transactionType
}

func parseScapiaMail(msg *enmime.Envelope) (string, time.Time, string, string, string) {
	email_time := parseMailTime(msg.GetHeader("Date"))
	amountRegex := regexp.MustCompile(`Amount\s*₹([\d,]+\.?\d*)`)
	amountMatch := amountRegex.FindStringSubmatch(msg.Text)
	var amount string
	if len(amountMatch) > 1 {
		amount = amountMatch[1]
	}
	merchantRegex := regexp.MustCompile(`Merchant\s*(.+?)(?:\n|Not you)`)
	merchantMatch := merchantRegex.FindStringSubmatch(msg.Text)
	var merchant string
	if len(merchantMatch) > 1 {
		merchant = strings.TrimSpace(merchantMatch[1])
	}
	var creditCardNumber string
	creditCardRegex := regexp.MustCompile(`Credit Card ending in (\d+) has`)
	creditCardMatch := creditCardRegex.FindStringSubmatch(msg.Text)
	if len(creditCardMatch) > 1 {
		creditCardNumber = strings.TrimSpace(creditCardMatch[1])
	}
	return merchant, email_time, amount, creditCardNumber, "Credit Card"
}

func parseMailTime(date_header string) time.Time {
	email_time, err := time.Parse("Mon, 2 Jan 2006 15:04:05 -0700", date_header)
	if err != nil {
		log.Fatalf("Error parsing the time: %v", err)
	}
	return email_time
}

func genericParser(msg *enmime.Envelope) (string, time.Time, string) {
	return "", time.Time{}, ""
}
