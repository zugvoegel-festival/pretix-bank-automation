package main

import (
	"fmt"
	"log"
	"time"

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

func getTransactionsFromToday() ([]go_nordigen.Transaction, error) {

	dateTo := time.Now().UTC().Format("2006-01-02")

	dateFrom := time.Now().UTC().Add(time.Duration(-24) * time.Hour).Format("2006-01-02")

	txs, err := nordigenConfig.Client.GetAccountTransactions(nordigenConfig.AccountId, dateFrom, dateTo)
	if err != nil {
		log.Printf("%v", err)
		return nil, fmt.Errorf("%v", err)
	}

	return txs.Booked, err
}
