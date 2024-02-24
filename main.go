package main

import (
	"fmt"
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
		msg := fmt.Sprintf("Error parsing duration : %v\n", err)
		log.Println(msg)

	}
	app := PretixBankAutomation{}
	for {
		app.Run()
		time.Sleep(time.Second * time.Duration(dur.Seconds()))
	}
}
