package main

import (
	"log"
	"xks/cmd"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found, proceeding with existing environment variables.")
	}
	cmd.Execute()
}
