package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

var NordigenAPIKey string
var PretixAPIKey string

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	NordigenAPIKey := os.Getenv("NORDIGEN_API_KEY")
	PretixAPIKey := os.Getenv("PRETIX_API_KEY")

	fmt.Println("Nordigen API key:", NordigenAPIKey)
	fmt.Println("Nordigen API key:", PretixAPIKey)
}

type NordigenTransaction struct {
	Description string `json:"transaction_description"`
	// Add other relevant fields from Nordigen response
}

type PretixOrder struct {
	OrderNumber string `json:"code"`
	// Add other relevant fields from Pretix response
}

func main() {

	nordigenEndpoint := os.Getenv("https://api.nordigen.com/account-access/v1/...")
	pretixEndpoint := os.Getenv("https://pretix.example.com/api/v1/organizers/your-organizer/events/your-event/orders/")

	// Fetch Nordigen transactions
	nordigenTransactions, err := getNordigenTransactions()
	if err != nil {
		log.Fatalf("Error fetching Nordigen transactions:", err)
		return
	}

	// Fetch Pretix orders
	pretixOrders, err := getPretixOrders()
	if err != nil {
		log.Fatalf("Error fetching Pretix orders:", err)
		return
	}

	// Compare and mark orders as paid
	markOrdersAsPaid(nordigenTransactions, pretixOrders)
}

func getNordigenTransactions() ([]NordigenTransaction, error) {
	// Make a request to Nordigen API
	resp, err := http.Get(nordigenEndpoint + "?api_key=" + nordigenAPIKey)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read and parse the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var transactions []NordigenTransaction
	err = json.Unmarshal(body, &transactions)
	if err != nil {
		return nil, err
	}

	return transactions, nil
}

func getPretixOrders() ([]PretixOrder, error) {
	// Make a request to Pretix API
	req, err := http.NewRequest("GET", pretixEndpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Token "+pretixAPIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read and parse the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var orders []PretixOrder
	err = json.Unmarshal(body, &orders)
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func markOrdersAsPaid(nordigenTransactions []NordigenTransaction, pretixOrders []PretixOrder) {
	// Implement the logic to compare transactions and orders, and mark orders as paid.
	for _, nordigenTransaction := range nordigenTransactions {
		for _, pretixOrder := range pretixOrders {
			if nordigenTransaction.Description == pretixOrder.OrderNumber {
				// Mark the order as paid (implement according to your Pretix API)
				fmt.Printf("Mark order %s as paid\n", pretixOrder.OrderNumber)
				// Add your logic here to mark orders as paid in Pretix
			}
		}
	}
}
