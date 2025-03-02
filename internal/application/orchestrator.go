package application

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

type Expression struct {
	ID     string  `json:"id"`
	Status string  `json:"status"`
	Result float64 `json:"result"`
}

var expressions = map[string]*Expression{}
var tasks = []Task{}
var taskIDCounter = 1
var expressionIDCounter = 1
var mu sync.Mutex

func init() {
	loadEnvironmentVariables()
}

func loadEnvironmentVariables() {
	if os.Getenv("TIME_ADDITION_MS") == "" || os.Getenv("TIME_SUBTRACTION_MS") == "" ||
		os.Getenv("TIME_MULTIPLICATIONS_MS") == "" || os.Getenv("TIME_DIVISIONS_MS") == "" {
		log.Fatal("All operation time environment variables must be set")
	}
}

func calculate(expression string) (float64, error) {
	// В реальной ситуации используйте пакет calculation.go
	return 0, nil
}

func addExpression(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Expression string `json:"expression"`
	}

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusUnprocessableEntity)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	expressions[fmt.Sprintf("%d", expressionIDCounter)] = &Expression{
		ID:     fmt.Sprintf("%d", expressionIDCounter),
		Status: "pending",
		Result: 0,
	}
	expressionIDCounter++

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"id": fmt.Sprintf("%d", expressionIDCounter-1)})
}

func listExpressions(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]*Expression{"expressions": getExpressions()})
}

func getExpressions() []*Expression {
	result := []*Expression{}
	for _, expr := range expressions {
		result = append(result, expr)
	}
	return result
}

func getExpressionByID(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	mu.Lock()
	defer mu.Unlock()

	expr, exists := expressions[id]
	if !exists {
		http.Error(w, "Expression not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]*Expression{"expression": expr})
}

func getTask(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	if len(tasks) == 0 {
		http.Error(w, "No tasks available", http.StatusNotFound)
		return
	}

	task := tasks[0]
	tasks = tasks[1:]

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]Task{"task": task})

	time.Sleep(time.Duration(task.OperationTime) * time.Millisecond)
}

func submitTaskResult(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ID     int     `json:"id"`
		Result float64 `json:"result"`
	}

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusUnprocessableEntity)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	// Здесь можно обновить статус выражения
	fmt.Printf("Task %d completed with result: %f\n", input.ID, input.Result)

	w.WriteHeader(http.StatusOK)
}

func RunOrchestrator() {
	router := mux.NewRouter()

	router.HandleFunc("/api/v1/calculate", addExpression).Methods("POST")
	router.HandleFunc("/api/v1/expressions", listExpressions).Methods("GET")
	router.HandleFunc("/api/v1/expressions/{id}", getExpressionByID).Methods("GET")
	router.HandleFunc("/internal/task", getTask).Methods("GET")
	router.HandleFunc("/internal/task", submitTaskResult).Methods("POST")

	log.Println("Starting orchestrator on :8080")
	go http.ListenAndServe(":8080", router)
}
