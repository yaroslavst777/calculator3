package application

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type TaskResponse struct {
	Task Task `json:"task"`
}

type ResultRequest struct {
	ID     int     `json:"id"`
	Result float64 `json:"result"`
}

func getTaskAgent(orchestratorURL string) (*Task, error) {
	resp, err := http.Get(orchestratorURL + "/internal/task")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil // No tasks available
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var taskResp TaskResponse
	err = json.Unmarshal(body, &taskResp)
	if err != nil {
		return nil, err
	}

	return &taskResp.Task, nil
}

func submitTaskResultAgent(orchestratorURL string, task *Task, result float64) error {
	reqBody, err := json.Marshal(ResultRequest{ID: task.ID, Result: result})
	if err != nil {
		return err
	}

	resp, err := http.Post(orchestratorURL+"/internal/task", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to submit result: %d", resp.StatusCode)
	}

	return nil
}

func worker(orchestratorURL string) {
	for {
		task, err := getTaskAgent(orchestratorURL)
		if err != nil {
			log.Printf("Error getting task: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		if task == nil {
			time.Sleep(5 * time.Second) // No tasks available
			continue
		}

		// Simulate calculation
		var result float64
		switch task.Operation {
		case "+":
			result = task.Arg1 + task.Arg2
		case "-":
			result = task.Arg1 - task.Arg2
		case "*":
			result = task.Arg1 * task.Arg2
		case "/":
			if task.Arg2 != 0 {
				result = task.Arg1 / task.Arg2
			} else {
				log.Println("Division by zero")
				continue
			}
		default:
			log.Printf("Unknown operation: %s", task.Operation)
			continue
		}

		err = submitTaskResultAgent(orchestratorURL, task, result)
		if err != nil {
			log.Printf("Error submitting result: %v", err)
		}
	}
}

func RunAgent() {
	orchestratorURL := "http://localhost:8080" // Default URL
	if url := os.Getenv("ORCHESTRATOR_URL"); url != "" {
		orchestratorURL = url
	}

	computingPower := 1
	if cp := os.Getenv("COMPUTING_POWER"); cp != "" {
		var err error
		computingPower, err = strconv.Atoi(cp)
		if err != nil || computingPower <= 0 {
			log.Fatalf("Invalid COMPUTING_POWER value: %s", cp)
		}
	}

	var wg sync.WaitGroup
	for i := 0; i < computingPower; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			worker(orchestratorURL)
		}()
	}

	rand.Seed(time.Now().UnixNano())
	wg.Wait()
}
