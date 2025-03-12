package main

import (
	"fmt"
	"log"
	"strconv"
	"time"
	"unicode/utf8"

	go_nordigen "github.com/ricardograndecros/go-nordigen"
)

type NordigenConfig struct {
	SecretId  string             `json:"secret_id"`
	SecretKey string             `json:"secret_key"`
	AccountId string             `json:"account_id"`
	Client    go_nordigen.Client `json:"client"`
}

func NewNordigenConfig() *NordigenConfig {
	config := &NordigenConfig{}
	config.SecretId = getEnv("NORDIGEN_SECRET_ID")
	config.SecretKey = getEnv("NORDIGEN_SECRET_KEY")
	config.AccountId = getEnv("NORDIGEN_ACCOUNT_ID")
	localClient, err := go_nordigen.NewClient(config.SecretId, config.SecretKey)
	config.Client = *localClient
	if err != nil {
		log.Fatal(err)
	}

	return config
}

var nordigenConfig = NewNordigenConfig()

func decodeErr(errorObject error) (string, int) {
	errStr := errorObject.Error()
	if len(errStr) > 0 {
		// Find the last rune and its position
		r, size := utf8.DecodeLastRuneInString(errStr)
		if r == utf8.RuneError {
			return "Error decoding the last rune", 500
		}

		// Convert the last rune to an integer
		runeValue := int(r)

		// Remove the last rune and append its integer representation
		newErrStr := errStr[:len(errStr)-size] + strconv.Itoa(runeValue)

		return newErrStr, runeValue
	} else {
		return "The error string is empty.", 500
	}
}

func checkAuthentication() (*go_nordigen.Account, error, int) {

	AccountInfo, err := nordigenConfig.Client.GetAccountInfo(nordigenConfig.AccountId)
	if err != nil {
		errStr, errorCode := decodeErr(err)
		return nil, fmt.Errorf("%v", errStr), errorCode
	}
	return AccountInfo, nil, 200

}
func checkRequisition() (*go_nordigen.Requisition, error, int) {

	Requisition, err := nordigenConfig.Client.GetRequisitionsById(requisitionData.RequisitionID)
	if err != nil {
		errStr, errorCode := decodeErr(err)
		return nil, fmt.Errorf("%v", errStr), errorCode
	}
	return Requisition, nil, 200

}
func reAuthorize() {

	//nordigenConfig.Client.GetRequisitionsById()
	/*EndUserAgreement, err := nordigenConfig.Client.CreateUserAgreement("GLS_GEMEINSCHAFTSBANK_GENODEM1GLS")
	if err != nil {
		log.Fatal("Error authenticating with Nordigen")
	}
	log.Printf("%s", EndUserAgreement.Id)

	req := go_nordigen.Requisition{
		InstitutionID: "GLS_GEMEINSCHAFTSBANK_GENODEM1GLS",
		Redirect:      "http://tickets.zugvoegelfestival.org",
		Agreement:     EndUserAgreement.Id,
		Reference:     "Zugvögel Festival",
		Language:      "DE",
	}
	Requisition, err := nordigenConfig.Client.NewRequisition(req)
	if err != nil {
		errStr := err.Error()
		if len(errStr) > 0 {
			// Find the last rune and its position
			r, size := utf8.DecodeLastRuneInString(errStr)
			if r == utf8.RuneError {
				fmt.Println("Error decoding the last rune.")
				return fmt.Errorf("error decoding the last rune")
			}

			// Convert the last rune to an integer
			runeValue := int(r)

			// Remove the last rune and append its integer representation
			newErrStr := errStr[:len(errStr)-size] + strconv.Itoa(runeValue)

			log.Printf("%v", newErrStr)

		} else {
			fmt.Println("The error string is empty.")
		}
	}
	log.Printf("Requisition ID: %s", Requisition.Id)
	return nil
	*/
}
func getTransactionsFromToday() ([]go_nordigen.Transaction, error, int) {

	dateTo := time.Now().UTC().Format("2006-01-02")

	dateFrom := time.Now().UTC().Add(time.Duration(-24*3) * time.Hour).Format("2006-01-02")

	txs, err := nordigenConfig.Client.GetAccountTransactions(nordigenConfig.AccountId, dateFrom, dateTo)
	if err != nil {
		errStr, errorCode := decodeErr(err)
		return nil, fmt.Errorf("%v", errStr), errorCode
	}
	return txs.Booked, nil, 200
}
