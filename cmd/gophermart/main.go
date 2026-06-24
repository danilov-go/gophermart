package main

import (
	"github.com/danilov-go/gophermart/internal/config"
	"github.com/danilov-go/gophermart/internal/logger"
	"github.com/danilov-go/gophermart/internal/server"
	"github.com/go-chi/chi/v5"
)

func main() {
	configs := config.ConfigServer{
		Net: config.NetAddress{
			Host: "localhost",
			Port: 8080,
		},
		DatabaseUri:   "host=localhost user=gophermart password=123 dbname=gophermart sslmode=disable",
		AccrualAddres: "",
	}
	if err := logger.Initialize("info"); err != nil {
		panic(err)
	}
	configs.Get()
	r := chi.NewRouter()
	serv := server.New(configs.Net.String(), logger.Log.Sugar(), r)
	if err := serv.Run(); err != nil {
		panic(err)
	}
}
