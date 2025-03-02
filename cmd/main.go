package main

import "calculator3/internal/application"

func main() {
	app := application.New()
	app.RunServer()
}
