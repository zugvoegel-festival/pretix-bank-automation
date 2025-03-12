package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
)

type RequisitionData struct {
	RequisitionID string `json:"requisition_id"`
}

var requisitionData RequisitionData

func loadRequisitionData() {
	file, err := os.Open("requisitionData.json")
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("Requisition data file does not exist. Creating a new one.")
			saveRequisitionData(getEnv("NORDIGEN_REQUISITION_ID"))
			return
		}
		log.Fatalf("Failed to open requisition data file: %v", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("Failed to read requisition data file: %v", err)
	}

	err = json.Unmarshal(data, &requisitionData)
	if err != nil {
		log.Fatalf("Failed to unmarshal requisition data: %v", err)
	}
}

func saveRequisitionData(id string) {
	requisitionData.RequisitionID = id
	data, err := json.Marshal(requisitionData)
	if err != nil {
		log.Fatalf("Failed to marshal requisition data: %v", err)
	}

	err = os.WriteFile("requisitionData.json", data, 0644)
	if err != nil {
		log.Fatalf("Failed to write requisition data file: %v", err)
	}
}
