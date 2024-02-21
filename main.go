package main

import (
	"fmt"
	"time"
)

type MainConfig struct {
	Cycle string
}

func (c *MainConfig) init() {
	c.Cycle = getEnv("CYCLE")
}

func main() {
	var config MainConfig
	config.init()

	dur, err := time.ParseDuration(config.Cycle)
	if err != nil {
		fmt.Printf("Error parsing duration : %v\n", err)

	}
	for {
		fmt.Println("Foo")
		time.Sleep(time.Second * time.Duration(dur.Seconds()))
	}
}

type PretixBankAutomation struct {
	// filtered
}

// ReminderEmails.Run() will get triggered automatically.
func (e PretixBankAutomation) Run() {
	// Queries the DB
	// Sends some email
	fmt.Printf("Every 5 sec send reminder emails \n")
}
