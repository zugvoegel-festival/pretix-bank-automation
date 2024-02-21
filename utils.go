package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func getEnv(key string) string {
	godotenv.Load(".env")
	value := os.Getenv(key)
	if len(value) == 0 {
		log.Fatalf("Environment variable %s not set", key)
	}
	return value
}
