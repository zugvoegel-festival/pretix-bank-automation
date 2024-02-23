package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
)

type PretixBankAutomation struct {
}

func (e PretixBankAutomation) Run() {
	// 1. Get all transactions from the last 24 hours
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
			log.Println("%v RemittanceInfo: %s", err, transaction.RemittanceInformationUnstructured)
			continue
		}
		bankAutomationError.Code = orderCode

		// 3. Get order from Pretix using orderCode
		order, err := getPretixOrder(orderCode)
		if err != nil {
			addBankAutomationError(err.Error())
			log.Println("%v OrderCode: %s", err, orderCode)
			continue
		}
		if order.Status == "p" {
			addBankAutomationError("Order is already paid")
			log.Println(" %s. Please check %s", bankAutomationError.Reason, orderCode)
			continue
		}
		if order.Status == "e" {
			addBankAutomationError("Order is expired")
			bankAutomationErrors = append(bankAutomationErrors, bankAutomationError)
			log.Println(" %s. Please check %s", bankAutomationError.Reason, orderCode)
			continue
		}
		if order.Status == "c" {
			addBankAutomationError("Order is canceled")
			log.Println(" %s. Please check %s", bankAutomationError.Reason, orderCode)
			continue
		}
		// 4. if order is unpaid and amount is fitting . Mark as paid
		if order.Total == transaction.TransactionAmount.Amount && transaction.TransactionAmount.Currency == "EUR" {
			// TODO, SECURITY: check for currency!
			err := markAsPaid(orderCode)
			if err != nil {
				addBankAutomationError(err.Error())
				log.Println(" %s. Please check %s", bankAutomationError.Reason, orderCode)
				continue
			}
		} else {
			addBankAutomationError(fmt.Sprintf("amount doesn't match Order: %s  Transaction: %s %s", order.Total, transaction.TransactionAmount.Amount, transaction.TransactionAmount.Currency))
			log.Println(" %s. Please check %s", bankAutomationError.Reason, orderCode)
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
