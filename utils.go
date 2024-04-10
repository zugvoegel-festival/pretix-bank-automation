package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"log"
	"net/smtp"
	"os"
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

	err := smtp.SendMail(mailConfig.SmtpServer+":"+fmt.Sprint(mailConfig.SmtpPort), auth, mailConfig.SenderEmail, []string{mailConfig.RecipientEmail}, msg)

	if err != nil {
		log.Fatalf("Error sending mail: %v", err)
		return err

	}
	return nil
}

type BankAutomationError struct {
	Code                  string `json:"code"`
	FromAccount           string `json:"from_account"`
	BookingDate           string `json:"booking_date"`
	RemittanceInformation string `json:"remittance_information"`
	Reason                string `json:"reason"`
}
type BankAutomationSuccess struct {
	Code                  string `json:"code"`
	FromAccount           string `json:"from_account"`
	BankTransactionCode   string `json:"bank_transaction_code"`
	BookingDate           string `json:"booking_date"`
	RemittanceInformation string `json:"remittance_information"`
}

var bankAutomationErrors []BankAutomationError = []BankAutomationError{}
var bankAutomationError BankAutomationError
var bankAutomationSuccesses []BankAutomationSuccess = []BankAutomationSuccess{}
var bankAutomationSuccess BankAutomationSuccess

func addBankAutomationError(errorMessage string) {
	bankAutomationError.Reason = errorMessage
	bankAutomationErrors = append(bankAutomationErrors, bankAutomationError)
}
func addBankAutomationSuccess() {
	bankAutomationSuccesses = append(bankAutomationSuccesses, bankAutomationSuccess)
}
func createEmailBody() string {
	var errorBody string
	var successBody string

	if len(bankAutomationErrors) == 0 {
		errorBody = "No errors today. JUHU \n"
	} else {
		errorBody = convertErrorToCSV(bankAutomationErrors)
	}

	if len(bankAutomationErrors) == 0 {
		successBody = "No success today. moooeeepp \n"
	} else {
		successBody = convertSuccessToCSV(bankAutomationSuccesses)
	}

	return errorBody + "\n\n\n" + successBody

}
func convertErrorToCSV(data []BankAutomationError) string {
	buf := new(bytes.Buffer)
	w := csv.NewWriter(buf)
	w.Write([]string{"BookingDate", "OrderCode", "FromAccount", "RemittanceInformation", "Reason"})

	for _, row := range data {
		w.Write([]string{row.BookingDate, row.Code, row.FromAccount, row.RemittanceInformation, row.Reason})
	}
	w.Flush()
	result := fmt.Sprintf("%d errors marking orders as paid\n\n%s", len(data), buf.String())
	return result
}
func convertSuccessToCSV(data []BankAutomationSuccess) string {
	buf := new(bytes.Buffer)
	w := csv.NewWriter(buf)
	w.Write([]string{"BookingDate", "OrderCode", "FromAccount", "BankTransactionCode", "RemittanceInformation"})

	for _, row := range data {
		w.Write([]string{row.BookingDate, row.Code, row.FromAccount, row.BankTransactionCode, row.RemittanceInformation})
	}
	w.Flush()

	result := fmt.Sprintf("%d successfull orders marked as paid\n\n%s", len(data), buf.String())
	return result
}
