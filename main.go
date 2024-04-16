package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"
)

func main() {

	msg := fmt.Sprintf("Pretix Bank Automatisierung " + time.Now().Format("02-01-2006 15:04"))
	log.Println(msg)
	transactions, err := getTransactionsFromToday()
	if err != nil {
		msg := fmt.Sprintf("Error getting transactions: %v", err)
		sendEmailNotification(msg)
	}

	// 2. Scan the remittanceInformationUnstructured for the keyword {{EVENT_SLUG}}{{ORDER_CODE}}

	for _, transaction := range transactions {

		if transaction.DebtorAccount.IBAN == "" {
			// check if deposit or withdrawal
			continue
		}
		if transaction.RemittanceInformationUnstructured == "" {
			// check if remittance empty
			continue
		}
		BankAutomationSingleLog = BankAutomationLog{}
		BankAutomationSingleLog.RemittanceInformation = transaction.RemittanceInformationUnstructured
		BankAutomationSingleLog.FromAccount = transaction.DebtorAccount.IBAN
		BankAutomationSingleLog.BankTransactionCode = transaction.BookingDate
		BankAutomationSingleLog.BookingDate = transaction.BookingDate

		orderCode, err := parseRemittanceInformation(transaction.RemittanceInformationUnstructured, pretixConfig.EventSlug)
		if err != nil {
			addBankAutomationLog(bankError, err.Error())
			msg := fmt.Sprintf("%v RemittanceInfo: %s", err, transaction.RemittanceInformationUnstructured)
			log.Println(msg)
			continue
		}
		BankAutomationSingleLog.Code = orderCode

		// 3. Get order from Pretix using orderCode
		order, err := getPretixOrder(orderCode)
		if err != nil {
			addBankAutomationLog(bankError, err.Error())
			msg := fmt.Sprintf("%v OrderCode: %s", err, orderCode)
			log.Println(msg)
			continue
		}
		if order.Status == "n" && order.RequireApproval {
			addBankAutomationLog(bankError, "Order is pending")
			msg := fmt.Sprintf(" %s. Please check %s", BankAutomationSingleLog.Reason, orderCode)
			log.Println(msg)
			continue
		}
		if order.Status == "p" {
			addBankAutomationLog(bankWarning, "Order is already paid")
			msg := fmt.Sprintf(" %s. Please check %s", BankAutomationSingleLog.Reason, orderCode)
			log.Println(msg)
			continue
		}
		if order.Status == "e" {
			addBankAutomationLog(bankError, "Order is expired")
			msg := fmt.Sprintf(" %s. Please check %s", BankAutomationSingleLog.Reason, orderCode)
			log.Println(msg)
			continue
		}
		if order.Status == "c" {
			addBankAutomationLog(bankError, "Order is canceled")
			msg := fmt.Sprintf(" %s. Please check %s", BankAutomationSingleLog.Reason, orderCode)
			log.Println(msg)
			continue
		}
		// 4. if order is unpaid and amount is fitting . Mark as paid
		if order.Total == transaction.TransactionAmount.Amount && transaction.TransactionAmount.Currency == "EUR" {
			err := markAsPaid(orderCode)
			if err != nil {
				addBankAutomationLog(bankError, err.Error())
				msg := fmt.Sprintf(" %s. Please check %s", BankAutomationSingleLog.Reason, orderCode)
				log.Println(msg)
				continue
			}
			addBankAutomationLog(bankSuccess, "")
		} else {
			addBankAutomationLog(bankError, fmt.Sprintf("amount doesn't match Order: %s  Transaction: %s %s", order.Total, transaction.TransactionAmount.Amount, transaction.TransactionAmount.Currency))
			msg := fmt.Sprintf(" %s. Please check %s", BankAutomationSingleLog.Reason, orderCode)
			log.Println(msg)
			continue
		}
	}

	body := convertToCSV()

	sendEmailNotification(body)

}

func parseRemittanceInformation(input string, eventSlug string) (string, error) {

	pattern := "^(?i).*" + eventSlug + "-([A-Z0-9]{5}).*"

	// Compile the regular expression
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", fmt.Errorf("error compiling regex: %v", err)
	}

	input = strings.ReplaceAll(input, " ", "")
	match := re.FindStringSubmatch(input)
	if match != nil {
		return match[1], nil
	}
	return "", fmt.Errorf("couldn't parse remittance info")
}
