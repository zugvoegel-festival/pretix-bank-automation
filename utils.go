package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

func getEnv(key string) string {
	godotenv.Load(".env")
	value := os.Getenv(key)
	if len(value) == 0 {
		log.Fatalf("Environment variable %s not set", key)
	}
	return value
}

type MailConfig struct {
	SmtpPort       string `json:"smtp_port"`
	SmtpServer     string `json:"smtp_server"`
	SmtpUsername   string `json:"smtp_username"`
	SmtpPassword   string `json:"smtp_password"`
	SenderEmail    string `json:"smtp_email"`
	RecipientEmail string `json:"recipient_email"`
}

func NewMailConfig() *MailConfig {
	config := &MailConfig{}

	config.SmtpPort = getEnv("SMTP_PORT")
	config.SmtpServer = getEnv("SMTP_SERVER")
	config.SmtpUsername = getEnv("SMTP_USER")
	config.SmtpPassword = getEnv("SMTP_PASSWORD")
	config.SenderEmail = getEnv("SENDER_MAIL")
	config.RecipientEmail = getEnv("RECIPIENT_MAIL")

	return config
}

var mailConfig = NewMailConfig()

func sendEmailNotification(body string) error {
	// Email content
	subject := "Pretix Bank Automatisierung " + time.Now().Format("02-01-2006 15:04")

	// Authentication
	auth := smtp.PlainAuth("", mailConfig.SmtpUsername, mailConfig.SmtpPassword, mailConfig.SmtpServer)

	// Sending email
	msg := []byte("To: " + mailConfig.RecipientEmail + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body + "\r\n")

	recipients := strings.Split(mailConfig.RecipientEmail, ",")
	err := smtp.SendMail(mailConfig.SmtpServer+":"+fmt.Sprint(mailConfig.SmtpPort), auth, mailConfig.SenderEmail, recipients, msg)

	if err != nil {
		log.Fatalf("Error sending mail: %v", err)
		return err

	}
	return nil
}

type BankAutomationLogType int

const (
	bankError BankAutomationLogType = iota
	bankWarning
	bankSuccess
	bankVerbose
)

type BankAutomationLog struct {
	Code                  string                `json:"code"`
	FromAccount           string                `json:"from_account"`
	BankTransactionCode   string                `json:"bank_transaction_code"`
	BookingDate           string                `json:"booking_date"`
	RemittanceInformation string                `json:"remittance_information"`
	Reason                string                `json:"reason"`
	Type                  BankAutomationLogType `json:"bank_automation_error_type"`
}

var BankAutomationLogs []BankAutomationLog = []BankAutomationLog{}
var BankAutomationSingleLog BankAutomationLog

func addBankAutomationLog(bankAutomationType BankAutomationLogType, errorMessage string) {

	BankAutomationSingleLog.Reason = errorMessage
	BankAutomationSingleLog.Type = bankAutomationType
	BankAutomationLogs = append(BankAutomationLogs, BankAutomationSingleLog)
}

func convertToCSV() string {
	errorBuff := new(bytes.Buffer)
	errorWriter := csv.NewWriter(errorBuff)
	errorWriter.Write([]string{"BookingDate", "OrderCode", "FromAccount", "RemittanceInformation", "Reason"})

	successBuff := new(bytes.Buffer)
	successWriter := csv.NewWriter(successBuff)
	successWriter.Write([]string{"BookingDate", "OrderCode", "FromAccount", "BankTransactionCode", "RemittanceInformation"})

	warningBuff := new(bytes.Buffer)
	warningWriter := csv.NewWriter(warningBuff)
	warningWriter.Write([]string{"BookingDate", "OrderCode", "FromAccount", "RemittanceInformation", "Reason"})

	errorCount := 0
	warningCount := 0
	successCount := 0

	for _, row := range BankAutomationLogs {
		switch row.Type {
		case bankError:
			errorWriter.Write([]string{row.BookingDate, row.Code, row.FromAccount, row.RemittanceInformation, row.Reason})
			errorCount++
			break
		case bankWarning:
			warningWriter.Write([]string{row.BookingDate, row.Code, row.FromAccount, row.RemittanceInformation, row.Reason})
			warningCount++
			break
		case bankSuccess:
			successWriter.Write([]string{row.BookingDate, row.Code, row.FromAccount, row.BankTransactionCode, row.RemittanceInformation})
			successCount++
			break
		}
	}

	errorWriter.Flush()
	successWriter.Flush()
	warningWriter.Flush()

	errorResult := fmt.Sprintf("%d errors marking orders as paid\n\n%s", errorCount, errorBuff.String())
	successResult := fmt.Sprintf("%d successfull marked orders as paid\n\n%s", successCount, successBuff.String())
	warningResult := fmt.Sprintf("%d warnings marking orders as paid\n\n%s", warningCount, warningBuff.String())

	return errorResult + "\n\n\n" + successResult + "\n\n\n" + warningResult
}
