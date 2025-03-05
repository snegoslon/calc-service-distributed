package main

import (
	"distributed-calc/internal/application"
	"log"
)

func main() {
	agent := application.NewAgent()
	log.Println("Starting agent...")
	agent.Run()
}