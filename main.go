package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type NordigenTransaction struct {
	MandateID         string `json:"mandateId,omitempty"`
	CreditorID        string `json:"creditorId,omitempty"`
	BookingDate       string `json:"bookingDate"`
	TransactionAmount struct {
		Amount   string `json:"amount"`
		Currency string `json:"currency"`
	} `json:"transactionAmount"`
	CreditorName    string `json:"creditorName,omitempty"`
	CreditorAccount struct {
		Iban string `json:"iban"`
	} `json:"creditorAccount,omitempty"`
	RemittanceInformationUnstructured string `json:"remittanceInformationUnstructured,omitempty"`
	ProprietaryBankTransactionCode    string `json:"proprietaryBankTransactionCode"`
	InternalTransactionID             string `json:"internalTransactionId"`
	DebtorName                        string `json:"debtorName,omitempty"`
	DebtorAccount                     struct {
		Iban string `json:"iban"`
	} `json:"debtorAccount,omitempty"`
	TransactionID string `json:"transactionId,omitempty"`
}
type NordigenTransactionsResponse struct {
	Transactions struct {
		Booked  []NordigenTransaction `json:"booked"`
		Pending []any                 `json:"pending"`
	} `json:"transactions"`
}
type PretixOrder struct {
	Code             string    `json:"code"`
	Event            string    `json:"event"`
	Status           string    `json:"status"`
	Testmode         bool      `json:"testmode"`
	Secret           string    `json:"secret"`
	Email            string    `json:"email"`
	Phone            any       `json:"phone"`
	Locale           string    `json:"locale"`
	Datetime         time.Time `json:"datetime"`
	Expires          time.Time `json:"expires"`
	PaymentDate      string    `json:"payment_date"`
	PaymentProvider  string    `json:"payment_provider"`
	Fees             []any     `json:"fees"`
	Total            string    `json:"total"`
	Comment          string    `json:"comment"`
	CustomFollowupAt any       `json:"custom_followup_at"`
	InvoiceAddress   struct {
		LastModified time.Time `json:"last_modified"`
		IsBusiness   bool      `json:"is_business"`
		Company      string    `json:"company"`
		Name         string    `json:"name"`
		NameParts    struct {
			Scheme string `json:"_scheme"`
		} `json:"name_parts"`
		Street            string `json:"street"`
		Zipcode           string `json:"zipcode"`
		City              string `json:"city"`
		Country           string `json:"country"`
		State             string `json:"state"`
		VatID             string `json:"vat_id"`
		VatIDValidated    bool   `json:"vat_id_validated"`
		CustomField       any    `json:"custom_field"`
		InternalReference string `json:"internal_reference"`
	} `json:"invoice_address"`
	Positions []struct {
		ID                int    `json:"id"`
		Order             string `json:"order"`
		Positionid        int    `json:"positionid"`
		Item              int    `json:"item"`
		Variation         any    `json:"variation"`
		Price             string `json:"price"`
		AttendeeName      string `json:"attendee_name"`
		AttendeeNameParts struct {
			Scheme     string `json:"_scheme"`
			GivenName  string `json:"given_name"`
			FamilyName string `json:"family_name"`
		} `json:"attendee_name_parts"`
		Company       any    `json:"company"`
		Street        any    `json:"street"`
		Zipcode       any    `json:"zipcode"`
		City          any    `json:"city"`
		Country       any    `json:"country"`
		State         any    `json:"state"`
		Discount      any    `json:"discount"`
		AttendeeEmail any    `json:"attendee_email"`
		Voucher       any    `json:"voucher"`
		TaxRate       string `json:"tax_rate"`
		TaxValue      string `json:"tax_value"`
		Secret        string `json:"secret"`
		AddonTo       any    `json:"addon_to"`
		Subevent      any    `json:"subevent"`
		Checkins      []any  `json:"checkins"`
		Downloads     []struct {
			Output string `json:"output"`
			URL    string `json:"url"`
		} `json:"downloads"`
		Answers            []any  `json:"answers"`
		TaxRule            any    `json:"tax_rule"`
		PseudonymizationID string `json:"pseudonymization_id"`
		Seat               any    `json:"seat"`
		Canceled           bool   `json:"canceled"`
		ValidFrom          any    `json:"valid_from"`
		ValidUntil         any    `json:"valid_until"`
		Blocked            any    `json:"blocked"`
	} `json:"positions"`
	Downloads []struct {
		Output string `json:"output"`
		URL    string `json:"url"`
	} `json:"downloads"`
	CheckinAttention bool      `json:"checkin_attention"`
	CheckinText      any       `json:"checkin_text"`
	LastModified     time.Time `json:"last_modified"`
	Payments         []struct {
		LocalID     int       `json:"local_id"`
		State       string    `json:"state"`
		Amount      string    `json:"amount"`
		Created     time.Time `json:"created"`
		PaymentDate any       `json:"payment_date"`
		Provider    string    `json:"provider"`
		PaymentURL  any       `json:"payment_url"`
		Details     struct {
		} `json:"details"`
	} `json:"payments"`
	Refunds         []any  `json:"refunds"`
	RequireApproval bool   `json:"require_approval"`
	SalesChannel    string `json:"sales_channel"`
	URL             string `json:"url"`
	Customer        any    `json:"customer"`
	ValidIfPending  bool   `json:"valid_if_pending"`
}

type BankAutomationError struct {
	Code                  string `json:"code"`
	FromAccount           string `json:"fromAccount"`
	RemittanceInformation string `json:"remittanceInformation"`
	Reason                string `json:"reason"`
}

var envNordigenAPIKey string
var envPretixAPIKey string
var envNordigenAccountID string
var envPretixEventSlug string
var envPretixOrganizerSlug string
var envPretixBaseUrl string
var smtpPort string
var smtpServer string
var smtpUsername string
var smtpPassword string
var senderEmail string
var recipientEmail string
var bankAutomationErrors []BankAutomationError = []BankAutomationError{}
var bankAutomationError BankAutomationError

func getenv(key string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		log.Fatalf("Environment variable %s not set", key)
	}
	return value
}

func init() {
	// Extra no error check here. Default should be defined environment variables. This is only for development
	godotenv.Load(".env")

	envNordigenAPIKey = getenv("NORDIGEN_API_KEY")
	envPretixAPIKey = getenv("PRETIX_API_KEY")
	envNordigenAccountID = getenv("NORDIGEN_ACCOUNT_ID")
	envPretixEventSlug = getenv("PRETIX_EVENT_SLUG")
	envPretixOrganizerSlug = getenv("PRETIX_ORGANIZER_SLUG")
	envPretixBaseUrl = getenv("PRETIX_BASE_URL")
	smtpPort = getenv("SMTP_PORT")
	smtpServer = getenv("SMTP_SERVER")
	smtpUsername = getenv("SMTP_USER")
	smtpPassword = getenv("SMTP_PASSWORD")
	senderEmail = getenv("SENDER_MAIL")
	recipientEmail = getenv("RECIPIENT_MAIL")
}

type Nordigen struct{}
type Pretix struct{}

func addBankAutomationError(errorMessage string) {
	bankAutomationError.Reason = errorMessage
	bankAutomationErrors = append(bankAutomationErrors, bankAutomationError)
}

func main() {

	// 1. Get all transactions from the last 24 hours
	transactions, err := getTransactionsFromLast24Hours()
	if err != nil {
		msg := fmt.Sprintf("Error getting transactions: %v", err)
		sendEmailNotification(msg)
		log.Fatalf(msg)

	}

	// 2. Scan the remittanceInformationUnstructured for the keyword {{EVENT_SLUG}}{{ORDER_CODE}}

	for _, transaction := range transactions {

		if transaction.DebtorAccount.Iban == "" {
			// check if deposit or withdrawal
			continue
		}
		if transaction.RemittanceInformationUnstructured == "" {
			// check if remittance empty
			continue
		}
		transaction.RemittanceInformationUnstructured = "FEST-F7CXX "
		bankAutomationError.RemittanceInformation = transaction.RemittanceInformationUnstructured
		bankAutomationError.FromAccount = transaction.DebtorAccount.Iban

		orderCode, err := parseRemittanceInformation(transaction.RemittanceInformationUnstructured)
		if err != nil {
			addBankAutomationError(err.Error())
			log.Printf("%v RemittanceInfo: %s", err, transaction.RemittanceInformationUnstructured)
			continue
		}
		bankAutomationError.Code = orderCode

		// 3. Get order from Pretix using orderCode
		order, err := getPretixOrder(orderCode)
		if err != nil {
			addBankAutomationError(err.Error())
			log.Printf("%v OrderCode: %s", err, orderCode)
			continue
		}
		if order.Status == "p" {
			addBankAutomationError("Order is already paid")
			log.Printf(" %s. Please check %s", bankAutomationError.Reason, orderCode)
			continue
		}
		if order.Status == "e" {
			addBankAutomationError("Order is expired")
			bankAutomationErrors = append(bankAutomationErrors, bankAutomationError)
			log.Printf(" %s. Please check %s", bankAutomationError.Reason, orderCode)
			continue
		}
		if order.Status == "c" {
			addBankAutomationError("Order is canceled")
			log.Printf(" %s. Please check %s", bankAutomationError.Reason, orderCode)
			continue
		}
		// 4. if order is unpaid and amount is fitting . Mark as paid
		if order.Total == transaction.TransactionAmount.Amount && transaction.TransactionAmount.Currency == "EUR" {
			// TODO, SECURITY: check for currency!
			err := markAsPaid(orderCode)
			if err != nil {
				addBankAutomationError(err.Error())
				log.Printf(" %s. Please check %s", bankAutomationError.Reason, orderCode)
				continue
			}
		} else {
			addBankAutomationError(fmt.Sprintf("amount doesn't match Order: %s  Transaction: %s %s", order.Total, transaction.TransactionAmount.Amount, transaction.TransactionAmount.Currency))
			log.Printf(" %s. Please check %s", bankAutomationError.Reason, orderCode)
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

func getTransactionsFromLast24Hours() ([]NordigenTransaction, error) {
	// Calculate start and end time for the last 24 hours
	currentTime := time.Now().UTC()
	startTime := currentTime.Add(-24 * time.Hour).Format("2006-01-02")
	endTime := currentTime.Format("2006-01-02")

	// Construct URL for Nordigen API endpoint
	var resp NordigenTransactionsResponse
	url := fmt.Sprintf("https://bankaccountdata.gocardless.com/api/v2/accounts/%s/transactions/?date_from=%s&date_to=%s", envNordigenAccountID, startTime, endTime)

	// Create HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	err = resp.fromRequest(req)
	return resp.Transactions.Booked, err
}

func parseRemittanceInformation(input string) (string, error) {

	prefix := "(?i)" + envPretixEventSlug + "-"

	pattern := "^" + prefix + "([A-Z0-9]{5})$"

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

func getPretixOrder(orderID string) (PretixOrder, error) {

	var order PretixOrder
	url := fmt.Sprintf("https://%s/api/v1/organizers/%s/events/%s/orders/%s/", envPretixBaseUrl, envPretixOrganizerSlug, envPretixEventSlug, orderID)

	// Create HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return order, err
	}

	err = order.fromRequest(req)
	return order, err

}
func (p *NordigenTransactionsResponse) fromRequest(req *http.Request) error {

	req.Header.Set("Authorization", "Bearer "+envNordigenAPIKey)

	// Send HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Nordigen API returned non-200 status code: %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(&p)
}

func (p *PretixOrder) fromRequest(req *http.Request) error {

	req.Header.Set("Authorization", "Token "+envPretixAPIKey)

	// Send HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Pretix API returned non-200 status code: %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(&p)
}

func markAsPaid(orderID string) error {
	// Construct URL for Pretix API endpoint
	url := fmt.Sprintf("https://%s/api/v1/organizers/%s/events/%s/orders/%s/mark_paid", envPretixBaseUrl, envPretixOrganizerSlug, envPretixEventSlug, orderID)

	// Create HTTP request
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}

	var order PretixOrder
	return order.fromRequest(req)
}

func convertToCSV(data []BankAutomationError) string {
	var lines []string
	lines = append(lines, "OrderCode,FromAccount,RemittanceInformation,Reason")
	for _, row := range data {
		row := fmt.Sprintf("%s,%s,%s,%s", row.Code, row.FromAccount, row.RemittanceInformation, row.Reason)
		lines = append(lines, row)
	}
	return strings.Join(lines, "\n")
}
func sendEmailNotification(body string) error {
	// Email content
	subject := "Pretix Bank Automatisierung " + time.Now().Format("02-01-2006")

	// Authentication
	auth := smtp.PlainAuth("", smtpUsername, smtpPassword, smtpServer)

	// Sending email
	msg := []byte("To: " + recipientEmail + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body + "\r\n")

	err := smtp.SendMail(smtpServer+":"+fmt.Sprint(smtpPort), auth, senderEmail, []string{recipientEmail}, msg)

	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}
	return nil
}
