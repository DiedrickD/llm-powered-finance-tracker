package main

import (
	initializers "github.com/DiedrickD/llm-powered-finance-tracker/initializers"
)

func main() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDB()
}
