package application

import (
	"calculator3/pkg/calculation"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Config struct {
	Addr string
}

func ConfigFromEnv() *Config {
	config := new(Config)
	config.Addr = os.Getenv("PORT")
	if config.Addr == "" {
		config.Addr = "8080"
	}
	return config
}

type Application struct {
	config *Config
}

func New() *Application {
	return &Application{
		config: ConfigFromEnv(),
	}
}

type Task struct {
	ID            int     `json:"id"`
	Arg1          float64 `json:"arg1"`
	Arg2          float64 `json:"arg2"`
	Operation     string  `json:"operation"`
	OperationTime int     `json:"operation_time"`
}

// Функция запуска сервера
func (a *Application) RunServer() {

	// Запускаем оркестратор
	go RunOrchestrator()

	// Даем время оркестратору запуститься
	time.Sleep(1 * time.Second)

	// Запускаем агента
	go RunAgent()

	log.Println("Application started successfully")

}

type Request struct {
	Expression string `json:"expression"`
}

type RequestData struct {
	Expression string `json:"expression"`
}

func makeResponse(w http.ResponseWriter, statusCode int, answer float64) {
	// Отправка JSON-ответа с ошибкой
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode) //код ответа

	var response map[string]interface{}

	//Status code 200
	if statusCode == http.StatusOK {
		response = map[string]interface{}{
			"result": answer,
		}
	}

	//Status code 422
	if statusCode == http.StatusUnprocessableEntity {
		response = map[string]interface{}{
			"error": "Expression is not valid",
		}
	}

	//Status code 500
	if statusCode == http.StatusInternalServerError {
		response = map[string]interface{}{
			"error": "Internal server error",
		}
	}
	json.NewEncoder(w).Encode(response)
}

func CalcHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		makeResponse(w, http.StatusInternalServerError, 0)
		return
	}

	var requestData RequestData

	data := make([]byte, 1024)
	num, errRead := r.Body.Read(data)
	defer r.Body.Close()
	if errRead != nil && errRead != io.EOF {
		makeResponse(w, http.StatusInternalServerError, 0)
		return
	}

	data = data[:num]

	errUnmarshal := json.Unmarshal(data, &requestData)

	if errUnmarshal != nil {
		makeResponse(w, http.StatusInternalServerError, 0)
	}

	// Получение значения expression из формы
	expression := requestData.Expression

	answer, errCalc := calculation.Calc(expression)
	if errCalc != nil {
		makeResponse(w, http.StatusUnprocessableEntity, 0)
		return
	}

	// Отправка JSON-ответа
	makeResponse(w, http.StatusOK, answer)
	fileName := "log.txt"
	logMessage := fmt.Sprintf("expression = %s, answer = %.2f", expression, answer)
	err := WriteToLogFile(logMessage, fileName)
	if err != nil {
		return
	}
}

func WriteToLogFile(message string, fileName string) error {
	// Открываем файл
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	// Закрываем файл после выхода из main
	defer file.Close()
	// Конфигурируем логгер, чтобы он выводил лог в файл
	log.SetOutput(file)

	log.Println(message)
	return nil
}
