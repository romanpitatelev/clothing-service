package main

import (
	"github.com/romanpitatelev/clothing-service/internal/app"
	"github.com/romanpitatelev/clothing-service/internal/configs"
)

func main() {
	cfg := configs.New(true)

	if err := app.Run(cfg); err != nil {
		panic(err)
	}
}
