package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

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

type PretixConfig struct {
	ApiKey        string `json:"api_key"`
	EventSlug     string `json:"event_Slug"`
	OrganizerSlug string `json:"organizer_slug"`
	BaseUrl       string `json:"base_url"`
}

func NewPretixConfig() *PretixConfig {
	config := &PretixConfig{}
	config.ApiKey = getEnv("PRETIX_API_KEY")
	config.EventSlug = getEnv("PRETIX_EVENT_SLUG")
	config.OrganizerSlug = getEnv("PRETIX_ORGANIZER_SLUG")
	config.BaseUrl = getEnv("PRETIX_BASE_URL")

	return config
}

var pretixConfig = NewPretixConfig()

func getPretixOrder(orderCode string) (PretixOrder, error) {

	var order PretixOrder

	url, err := url.JoinPath("https://", pretixConfig.BaseUrl, "/api/v1/organizers/", pretixConfig.OrganizerSlug, "/events/", pretixConfig.EventSlug, "/orders/", orderCode, "/")
	if err != nil {
		log.Fatalf("%v", err)
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return order, err
	}

	err = order.fromRequest(req)
	return order, err
}

func markAsPaid(orderCode string) error {

	url, err := url.JoinPath("https://", pretixConfig.BaseUrl, "/api/v1/organizers/", pretixConfig.OrganizerSlug, "/events/", pretixConfig.EventSlug, "/orders/", orderCode, "/mark_paid/")
	if err != nil {
		log.Fatalf("%v", err)
	}
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}

	var order PretixOrder
	return order.fromRequest(req)
}

func (p *PretixOrder) fromRequest(req *http.Request) error {

	req.Header.Set("Authorization", "Token "+pretixConfig.ApiKey)

	// Send HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("pretix API returned non-200 status code: %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(&p)
}
