package application

import (
	"distributed-calc/internal/parsing"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type Configuration struct {
	Port                 string
	TimeAdditionMs       int
	TimeSubtractionMs    int
	TimeMultiplicationMs int
	TimeDivisionMs       int
}

const DefaultTimeMs int = 100

func GetConfigFromEnvironment() *Configuration {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("ENV::PORT is %v\n", port)
	timeAddition, _ := strconv.Atoi(os.Getenv("TIME_ADDITION_MS"))
	if timeAddition == 0 {
		timeAddition = DefaultTimeMs
	}
	fmt.Printf("ENV::TIME_ADDITION_MS is %v\n", timeAddition)
	timeSubtraction, _ := strconv.Atoi(os.Getenv("TIME_SUBTRACTION_MS"))
	if timeSubtraction == 0 {
		timeSubtraction = DefaultTimeMs
	}
	fmt.Printf("ENV::TIME_SUBTRACTION_MS is %v\n", timeSubtraction)
	timeMultiplication, _ := strconv.Atoi(os.Getenv("TIME_MULTIPLICATIONS_MS"))
	if timeMultiplication == 0 {
		timeMultiplication = DefaultTimeMs
	}
	fmt.Printf("ENV::TIME_MULTIPLICATIONS_MS is %v\n", timeMultiplication)
	timeDivision, _ := strconv.Atoi(os.Getenv("TIME_DIVISIONS_MS"))
	if timeDivision == 0 {
		timeDivision = DefaultTimeMs
	}
	fmt.Printf("ENV::TIME_DIVISISIONS_MS is %v\n", timeDivision)
	return &Configuration{
		Port:                 port,
		TimeAdditionMs:       timeAddition,
		TimeSubtractionMs:    timeSubtraction,
		TimeMultiplicationMs: timeMultiplication,
		TimeDivisionMs:       timeDivision,
	}
}

type Orchestrator struct {
	Config      *Configuration
	exprStore   map[string]*Expression
	taskStore   map[string]*Task
	taskQueue   []*Task
	mu          sync.Mutex
	exprCounter int64
	taskCounter int64
}

func NewOrchestrator() *Orchestrator {
	return &Orchestrator{
		Config:    GetConfigFromEnvironment(),
		exprStore: make(map[string]*Expression),
		taskStore: make(map[string]*Task),
		taskQueue: make([]*Task, 0),
	}
}

type Expression struct {
	ID     string        `json:"id"`
	Expr   string        `json:"expression"`
	Status string        `json:"status"`
	Result *float64      `json:"result,omitempty"`
	Node   *parsing.Node `json:"-"`
	Error  bool          `json:"-"`
}

type Task struct {
	ID            string        `json:"id"`
	ExprID        string        `json:"-"`
	Arg1          float64       `json:"arg1"`
	Arg2          float64       `json:"arg2"`
	Operation     string        `json:"operation"`
	OperationTime int           `json:"operation_time"`
	Node          *parsing.Node `json:"-"`
}

func (o *Orchestrator) CalculationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Expression string `json:"expression"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Expression == "" {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	node, err := parsing.ParseExpression(req.Expression)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	o.exprCounter++
	exprID := fmt.Sprintf("%d", o.exprCounter)
	expr := &Expression{
		ID:     exprID,
		Expr:   req.Expression,
		Status: "pending",
		Node:   node,
		Error:  false,
	}
	o.exprStore[exprID] = expr
	o.scheduleTasks(expr)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"id": exprID})
}

func (o *Orchestrator) listExpressionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	expressions := make([]*Expression, 0, len(o.exprStore))
	for _, expr := range o.exprStore {
		if expr.Node != nil && expr.Node.IsLeaf {
			if expr.Error {
				expr.Status = "error"
			} else {
				expr.Status = "completed"
			}
			expr.Result = &expr.Node.Value
		}
		expressions = append(expressions, expr)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"expressions": expressions})
}

const baseUrlExpressions = "/api/v1/expressions"

func (o *Orchestrator) getExpressionByIdHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	id := r.URL.Path[len(baseUrlExpressions)+1:]
	o.mu.Lock()
	expr, ok := o.exprStore[id]
	o.mu.Unlock()
	if !ok {
		http.Error(w, fmt.Sprintf(`{"error":"Expression %s not found"}`, id), http.StatusNotFound)
		return
	}
	if expr.Node != nil && expr.Node.IsLeaf {
		if expr.Error {
			expr.Status = "error"
		} else {
			expr.Status = "completed"
		}
		expr.Result = &expr.Node.Value
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"expression": expr})
}

func (o *Orchestrator) getTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	if len(o.taskQueue) == 0 {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	task := o.taskQueue[0]
	o.taskQueue = o.taskQueue[1:]
	if expr, exists := o.exprStore[task.ExprID]; exists {
		expr.Status = "in_progress"
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"task": task})
}

func (o *Orchestrator) postTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		ID     string  `json:"id"`
		Result float64 `json:"result"`
		Status string  `json:"status"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.ID == "" {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	task, ok := o.taskStore[req.ID]
	if !ok {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	task.Node.IsLeaf = true
	task.Node.Value = req.Result
	delete(o.taskStore, req.ID)

	if expr, exists := o.exprStore[task.ExprID]; exists {
		if req.Status != "" {
			expr.Error = true
		} else {
			o.scheduleTasks(expr)
			if expr.Node.IsLeaf {
				expr.Status = "completed"
				expr.Result = &expr.Node.Value
			}
		}
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"OK"}`))
}

func (o *Orchestrator) scheduleTasks(expr *Expression) {
	var traverse func(node *parsing.Node)
	traverse = func(node *parsing.Node) {
		if node == nil || node.IsLeaf {
			return
		}
		traverse(node.Left)
		traverse(node.Right)
		if node.Left != nil && node.Right != nil && node.Left.IsLeaf && node.Right.IsLeaf {
			if !node.TaskScheduled {
				o.taskCounter++
				taskID := fmt.Sprintf("%d", o.taskCounter)
				var opTime int
				switch node.Operator {
				case "+":
					opTime = o.Config.TimeAdditionMs
				case "-":
					opTime = o.Config.TimeSubtractionMs
				case "*":
					opTime = o.Config.TimeMultiplicationMs
				case "/":
					opTime = o.Config.TimeDivisionMs
				default:
					opTime = 100
				}
				task := &Task{
					ID:            taskID,
					ExprID:        expr.ID,
					Arg1:          node.Left.Value,
					Arg2:          node.Right.Value,
					Operation:     node.Operator,
					OperationTime: opTime,
					Node:          node,
				}
				node.TaskScheduled = true
				o.taskStore[taskID] = task
				o.taskQueue = append(o.taskQueue, task)
			}
		}
	}
	traverse(expr.Node)
}

func (o *Orchestrator) RunServer() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/calculate", o.CalculationHandler)
	mux.HandleFunc(baseUrlExpressions, o.listExpressionsHandler)
	mux.HandleFunc(baseUrlExpressions+"/", o.getExpressionByIdHandler)
	mux.HandleFunc("/internal/task", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			o.getTaskHandler(w, r)
		} else if r.Method == http.MethodPost {
			o.postTaskHandler(w, r)
		} else {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})
	go func() {
		for {
			time.Sleep(2 * time.Second)
			o.mu.Lock()
			pendingTasks := len(o.taskQueue)
			if pendingTasks > 0 {
				log.Printf("Pending: %d\n", pendingTasks)
			}
			o.mu.Unlock()
		}
	}()
	address := fmt.Sprintf(":%s", o.Config.Port)
	return http.ListenAndServe(address, mux)
}