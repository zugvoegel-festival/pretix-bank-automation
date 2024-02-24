package main

import (
	"log"
	"time"
)

type MainConfig struct {
	Cycle string
}

var mainConfig = NewMainConfig()

func NewMainConfig() *MainConfig {
	config := &MainConfig{}
	config.Cycle = getEnv("CYCLE")

	return config
}

func main() {

	dur, err := time.ParseDuration(mainConfig.Cycle)
	if err != nil {
		log.Fatalf("Error parsing duration : %v\n", err)
	}
	app := PretixBankAutomation{}
	for {
		app.Run()
		time.Sleep(time.Second * time.Duration(dur.Seconds()))
	}
}
