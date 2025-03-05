package application

import (
	"bytes"
	"distributed-calc/pkg/calculation"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Agent struct {
	ComputingPower  int
	OrchestratorURL string
}

const Delay = 3 * time.Second

func NewAgent() *Agent {
	cp, err := strconv.Atoi(os.Getenv("COMPUTING_POWER"))
	if err != nil || cp < 1 {
		cp = 1
	}
	fmt.Printf("ENV::COMPUTING_POWER is %v\n", cp)
	orchestratorURL := os.Getenv("ORCHESTRATOR_URL")
	if orchestratorURL == "" {
		orchestratorURL = "http://localhost:8080"
	}
	fmt.Printf("ENV::ORCHESTRATOR_URL is %v\n", orchestratorURL)
	return &Agent{
		ComputingPower:  cp,
		OrchestratorURL: orchestratorURL,
	}
}

func (a *Agent) Run() {
	for i := range a.ComputingPower {
		go a.worker(i)
	}
	select {}
}

type orchestratorResponse struct {
	Task struct {
		ID            string  `json:"id"`
		Arg1          float64 `json:"arg1"`
		Arg2          float64 `json:"arg2"`
		Operation     string  `json:"operation"`
		OperationTime int     `json:"operation_time"`
	} `json:"task"`
}

func (a *Agent) worker(id int) {
	for {
		requestUrl := fmt.Sprintf("%s/internal/task", a.OrchestratorURL)
		resp, err := http.Get(requestUrl)
		if err != nil {
			log.Printf("Worker %d: error getting task: %v", id, err)
			time.Sleep(Delay)
			continue
		}
		if resp.StatusCode == http.StatusNotFound {
			_ = resp.Body.Close()
			time.Sleep(Delay)
			continue
		}

		var taskResponse orchestratorResponse

		err = json.NewDecoder(resp.Body).Decode(&taskResponse)

		if err != nil {
			_ = resp.Body.Close()
			log.Printf("Worker %d: error decoding task: %v", id, err)
			time.Sleep(Delay)
			continue
		}
		task := taskResponse.Task
		log.Printf("Worker %d: received task %s: %f %s %f, simulating %d ms", id, task.ID, task.Arg1, task.Operation, task.Arg2, task.OperationTime)
		time.Sleep(time.Duration(task.OperationTime) * time.Millisecond)
		result, err := calculation.Evaluate(task.Operation, task.Arg1, task.Arg2)
		resultPayload := map[string]interface{}{
			"id":     task.ID,
			"result": result,
			"status": "",
		}
		if err != nil {
			log.Printf("Worker %d: error computing task %s: %v", id, task.ID, err)
			resultPayload["status"] = "error"
		}
		payloadBytes, _ := json.Marshal(resultPayload)
		_, err = http.Post(requestUrl, "application/json", bytes.NewReader(payloadBytes))
		if err != nil {
			log.Printf("Worker %d: error posting result for task %s: %v", id, task.ID, err)
			continue
		}
	}
}