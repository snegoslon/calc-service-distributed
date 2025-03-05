package main

import (
	"distributed-calc/internal/application"
	"log"
)

func main() {
	app := application.NewOrchestrator()
	log.Println("Orchestrator port", app.Config.Port)
	if err := app.RunServer(); err != nil {
		log.Fatal(err)
	}
}