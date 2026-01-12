package main

import (
	"bytes"
	"encoding/base64"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jhillyerd/enmime"
)

func ParseMessage(message string) (string, time.Time, float64, string, string) {
	data, err := base64.URLEncoding.DecodeString(message)
	if err != nil {
		log.Fatalf("Unable to decode message: %v", err)
	}
	msg, err := enmime.ReadEnvelope(bytes.NewReader(data))
	if err != nil {
		log.Fatalf("Unable to read envelope: %v", err)
	}
	var email_time time.Time
	var merchant, creditCardNumber, transactionType string
	var amount float64

	if strings.Contains(string(data), "Scapia") {
		merchant, email_time, amount, creditCardNumber, transactionType = parseScapiaMail(msg)
	} else if strings.Contains(string(data), "Axis") {
		merchant, email_time, amount, creditCardNumber, transactionType = parseAxisMail(msg)
	} else if strings.Contains(string(data), "HDFC") {
		merchant, email_time, amount, creditCardNumber, transactionType = parseHdfcMail(msg)
	} else if strings.Contains(string(data), "ICICI") {
		merchant, email_time, amount, creditCardNumber, transactionType = parseIciciMail(msg)
	} else if strings.Contains(string(data), "SBI") {
		merchant, email_time, amount, creditCardNumber, transactionType = parseSbiMail(msg)
	} else {
		merchant, email_time, amount, creditCardNumber, transactionType = genericParser(msg)
	}
	return merchant, email_time, amount, creditCardNumber, transactionType
}

func transactionTypeFetcher(msg string) string {
	var transactionType string = "undefined"
	if strings.Contains(msg, "Scapia") || strings.Contains(msg, "HDFC Bank Credit Card") {
		transactionType = "Credit Card"
	} else if strings.Contains(msg, "BLOCKUPI") {
		transactionType = "UPI"
	} else if strings.Contains(msg, "ICICI Bank Credit Card") {
		transactionType = "ICICI Credit Card"
	} else if strings.Contains(msg, "SBI") {
		transactionType = "SBI UPI/ Debit Card"
	}
	return transactionType
}

func regexMatch(regexString string, msg string) string {
	regex := regexp.MustCompile(regexString)
	regexMatch := regex.FindStringSubmatch(msg)
	var regexItem string
	if len(regexMatch) > 1 {
		regexItem = regexMatch[1]
	}
	return strings.TrimSpace(regexItem)
}

func parseMailTime(date_header string) time.Time {
	emailTime, err := time.Parse("Mon, 2 Jan 2006 15:04:05 -0700", date_header)
	if err != nil {
		log.Fatalf("Error parsing the time: %v", err)
	}
	return emailTime
}

func parseScapiaMail(msg *enmime.Envelope) (string, time.Time, float64, string, string) {
	emailTime := parseMailTime(msg.GetHeader("Date"))
	receiverInformation := regexMatch(`Merchant\s*(.+?)(?:\n|Not you)`, msg.Text)
	senderInformation := regexMatch(`Credit Card ending in (\d+) has`, msg.Text)
	transactionType := transactionTypeFetcher(msg.Text)
	amount, errorFloat := strconv.ParseFloat(strings.ReplaceAll(regexMatch(`Amount\s*₹([\d,]+\.?\d*)`, msg.Text), ",", ""), 64)
	if errorFloat != nil {
		log.Fatalf("Error parsing the amount: %v", errorFloat)
	}
	return receiverInformation,
		emailTime,
		amount,
		senderInformation,
		transactionType
}

func parseAxisMail(msg *enmime.Envelope) (string, time.Time, float64, string, string) {
	emailTime := parseMailTime(msg.GetHeader("Date"))
	receiverInformation := regexMatch(`Transaction Info:\s*(.*)\s*`, msg.Text)
	senderInformation := regexMatch(`Account Number:\s*(.*)\s`, msg.Text)
	transactionType := transactionTypeFetcher(msg.Text)
	amount, errorFloat := strconv.ParseFloat(strings.ReplaceAll(regexMatch(`Amount Debited:\s*INR\s*([\d,]+\.?\d*)`, msg.Text), ",", ""), 64)
	if errorFloat != nil {
		log.Fatalf("Error parsing the amount: %v", errorFloat)
	}
	return receiverInformation,
		emailTime,
		amount,
		senderInformation,
		transactionType
}

func parseHdfcMail(msg *enmime.Envelope) (string, time.Time, float64, string, string) {
	emailTime := parseMailTime(msg.GetHeader("Date"))
	receiverInformation := regexMatch(`towards\s+(.+?)\s+on\s+\d`, msg.Text)
	senderInformation := regexMatch(`HDFC\sBank\sCredit\sCard\sending\s(.*)\stowards`, msg.Text)
	transactionType := transactionTypeFetcher(msg.Text)
	amount, errorFloat := strconv.ParseFloat(strings.ReplaceAll(regexMatch(`Rs.\s*([\d,]+\.?\d*) is`, msg.Text), ",", ""), 64)
	if errorFloat != nil {
		log.Fatalf("Error parsing the amount: %v", errorFloat)
	}
	return receiverInformation,
		emailTime,
		amount,
		senderInformation,
		transactionType
}

func parseIciciMail(msg *enmime.Envelope) (string, time.Time, float64, string, string) {
	emailTime := parseMailTime(msg.GetHeader("Date"))
	receiverInformation := regexMatch(`Info:\s*(.+?)\.\s`, msg.Text)
	senderInformation := regexMatch(`Your\sICICI\sBank\sCredit\sCard\s+(.*?)\shas`, msg.Text)
	transactionType := transactionTypeFetcher(msg.Text)
	amount, errorFloat := strconv.ParseFloat(strings.ReplaceAll(regexMatch(`transaction\sof\sINR\s+([\d,]+\.?\d*)\son`, msg.Text), ",", ""), 64)
	if errorFloat != nil {
		log.Fatalf("Error parsing the amount: %v", errorFloat)
	}
	return receiverInformation,
		emailTime,
		amount,
		senderInformation,
		transactionType
}

func parseSbiMail(msg *enmime.Envelope) (string, time.Time, float64, string, string) {
	emailTime := parseMailTime(msg.GetHeader("Date"))
	receiverInformation := regexMatch(`debit\sby\s(.*?)\sof`, msg.Text)
	senderInformation := regexMatch(`Your\sA\/C\s(.*?)\shas`, msg.Text)
	transactionType := transactionTypeFetcher(msg.Text)
	amount, errorFloat := strconv.ParseFloat(strings.ReplaceAll(regexMatch(`Rs\s(.*?)\son`, msg.Text), ",", ""), 64)
	if errorFloat != nil {
		log.Fatalf("Error parsing the amount: %v", errorFloat)
	}
	return receiverInformation,
		emailTime,
		amount,
		senderInformation,
		transactionType
}

func genericParser(msg *enmime.Envelope) (string, time.Time, float64, string, string) {
	return "", time.Time{}, 0.0, "", ""
}
