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
	transactions, err := getTransactionsFromLast24Hours()
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
		bankAutomationError.RemittanceInformation = transaction.RemittanceInformationUnstructured
		bankAutomationError.FromAccount = transaction.DebtorAccount.IBAN

		orderCode, err := parseRemittanceInformation(transaction.RemittanceInformationUnstructured, pretixConfig.EventSlug)
		if err != nil {
			addBankAutomationError(err.Error())
			msg := fmt.Sprintf("%v RemittanceInfo: %s", err, transaction.RemittanceInformationUnstructured)
			log.Println(msg)
			continue
		}
		bankAutomationError.Code = orderCode

		// 3. Get order from Pretix using orderCode
		order, err := getPretixOrder(orderCode)
		if err != nil {
			addBankAutomationError(err.Error())
			msg := fmt.Sprintf("%v OrderCode: %s", err, orderCode)
			log.Println(msg)
			continue
		}
		if order.Status == "n" {
			addBankAutomationError("Order is pending")
			msg := fmt.Sprintf(" %s. Please check %s", bankAutomationError.Reason, orderCode)
			log.Println(msg)
			continue
		}
		if order.Status == "p" {
			addBankAutomationError("Order is already paid")
			msg := fmt.Sprintf(" %s. Please check %s", bankAutomationError.Reason, orderCode)
			log.Println(msg)
			continue
		}
		if order.Status == "e" {
			addBankAutomationError("Order is expired")
			bankAutomationErrors = append(bankAutomationErrors, bankAutomationError)
			msg := fmt.Sprintf(" %s. Please check %s", bankAutomationError.Reason, orderCode)
			log.Println(msg)
			continue
		}
		if order.Status == "c" {
			addBankAutomationError("Order is canceled")
			msg := fmt.Sprintf(" %s. Please check %s", bankAutomationError.Reason, orderCode)
			log.Println(msg)
			continue
		}
		// 4. if order is unpaid and amount is fitting . Mark as paid
		if order.Total == transaction.TransactionAmount.Amount && transaction.TransactionAmount.Currency == "EUR" {
			err := markAsPaid(orderCode)
			if err != nil {
				addBankAutomationError(err.Error())
				msg := fmt.Sprintf(" %s. Please check %s", bankAutomationError.Reason, orderCode)
				log.Println(msg)
				continue
			}
		} else {
			addBankAutomationError(fmt.Sprintf("amount doesn't match Order: %s  Transaction: %s %s", order.Total, transaction.TransactionAmount.Amount, transaction.TransactionAmount.Currency))
			msg := fmt.Sprintf(" %s. Please check %s", bankAutomationError.Reason, orderCode)
			log.Println(msg)
			continue
		}
	}

	var body string

	if len(bankAutomationErrors) == 0 {
		body = "No errors today. JUHU"
	} else {
		body = convertToCSV(bankAutomationErrors)
	}

	sendEmailNotification(body)

}

func parseRemittanceInformation(input string, eventSlug string) (string, error) {

	pattern := "^(?i)" + eventSlug + "-([A-Z0-9]{5})$"

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
