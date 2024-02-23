package main

import (
	"fmt"
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
		fmt.Println("Error parsing duration : %v\n", err)

	}
	app := PretixBankAutomation{}
	for {
		fmt.Println("Foo")
		app.Run()
		time.Sleep(time.Second * time.Duration(dur.Seconds()))
	}
}
