package main

import (
	"github.com/danilov-go/gophermart/internal/config"
	"github.com/danilov-go/gophermart/internal/handler"
	"github.com/danilov-go/gophermart/internal/logger"
	"github.com/danilov-go/gophermart/internal/repository"
	"github.com/danilov-go/gophermart/internal/server"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	configs := config.ConfigServer{
		Net: config.NetAddress{
			Host: "localhost",
			Port: 8080,
		},
		DatabaseUri:   "host=localhost user=gophermart password=123 dbname=gophermart sslmode=disable",
		AccrualAddres: "",
		Key:           "my_secret_key",
	}
	if err := logger.Initialize("info"); err != nil {
		panic(err)
	}
	configs.Get()
	storage := repository.InitMemStorage()
	r := chi.NewRouter()
	r.Use(middleware.StripSlashes)
	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", handler.RegisterUser(storage, configs.Key))
		r.Post("/login", handler.LoginUser(storage, configs.Key))
		r.Post("/orders", handler.AuthMiddleware(storage, configs.Key, handler.SaveOrderHandler()))
		r.Get("/orders", handler.AuthMiddleware(storage, configs.Key, handler.GetOrderHandler()))
	})
	serv := server.New(configs.Net.String(), logger.Log.Sugar(), r)
	if err := serv.Run(); err != nil {
		panic(err)
	}
}
