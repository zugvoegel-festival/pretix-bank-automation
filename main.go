package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"
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

var envNordigenAPIKey string
var envPretixAPIKey string
var envNordigenAccountID string
var envPretixEventSlug string
var envPretixOrganizerSlug string
var envPretixBaseUrl string

func getenv(key string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		log.Fatalf("Environment variable %s not set", key)
	}
	return value
}

func init() {
	envNordigenAPIKey = getenv("NORDIGEN_API_KEY")
	envPretixAPIKey = getenv("PRETIX_API_KEY")
	envNordigenAccountID = getenv("NORDIGEN_ACCOUNT_ID")
	envPretixEventSlug = getenv("PRETIX_EVENT_SLUG")
	envPretixOrganizerSlug = getenv("PRETIX_ORGANIZER_SLUG")
	envPretixBaseUrl = getenv("PRETIX_BASE_URL")
}

type Nordigen struct{}
type Pretix struct{}

func main() {

	// 1. Get all transactions from the last 24 hours
	transactions, err := getTransactionsFromLast24Hours()
	if err != nil {
		log.Fatalf("Error getting transactions: %v", err)
	}

	// 2. Scan the remittanceInformationUnstructured for the keyword {{EVENT_SLUG}}{{ORDER_CODE}}
	for _, transaction := range transactions {
		result, orderCode := parseRemittanceInformation(transaction.RemittanceInformationUnstructured)
		if result {
			// 3. Get order from Pretix using orderCode
			order, err := getPretixOrder(orderCode)
			if err != nil {
				log.Printf("Error getting order from Pretix for keyword %s: %v", orderCode, err)
				continue
			}
			if order.Status == "p" {
				log.Printf("Order %s is already paid. No further actions required", orderCode)
				continue
			}
			if order.Status == "e" {
				log.Printf("Order %s is expired. Please check Order", orderCode)
				continue
			}
			if order.Status == "c" {
				log.Printf("Order %s is canceled paid. Please check Order", orderCode)
				continue
			}
			// 4. if order is unpaid and amount is fitting . Mark as paid
			if order.Total == transaction.TransactionAmount.Amount {

				// TODO, SECURITY: check for currency!

				err := markAsPaid(orderCode)
				if err != nil {
					log.Printf("Error marking order in Pretix for order %s: %v", orderCode, err)
					continue
				}
			} else {
				log.Printf("Order %s is unpaid but amount doesent match %s  %s", orderCode, order.Total, transaction.TransactionAmount.Amount)
				continue
			}
		}
	}
}

func getTransactionsFromLast24Hours() ([]NordigenTransaction, error) {
	// Calculate start and end time for the last 24 hours
	currentTime := time.Now().UTC()
	startTime := currentTime.Add(-24 * time.Hour).Format("2006-01-02")
	endTime := currentTime.Format("2006-01-02")

	// Construct URL for Nordigen API endpoint
	url := fmt.Sprintf("https://bankaccountdata.gocardless.com/api/v2/accounts/%s/transactions/?date_from=%s&date_to=%s", envNordigenAccountID, startTime, endTime)

	// Create HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+envNordigenAPIKey)

	// Send HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Parse JSON response
	var nordigenResp NordigenTransactionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&nordigenResp); err != nil {
		return nil, err
	}

	return nordigenResp.Transactions.Booked, nil
}

func parseRemittanceInformation(remittanceInfo string) (bool, string) {
	pattern := envPretixEventSlug + `[A-Z0-9]{5}`

	// Compile regular expression
	re := regexp.MustCompile(pattern)

	// Find matches in the remittance information
	matches := re.FindStringSubmatch(remittanceInfo)

	// Check if any match is found
	if len(matches) > 0 {
		// First match is the keyword
		keyword := matches[0]

		// Rest of the information after the keyword
		rest := remittanceInfo[len(keyword):]

		return true, rest
	}

	// If no match is found, return empty values
	return false, ""
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
