package main

import (
	"context"

	"github.com/danilov-go/gophermart/internal/accrual"
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
		Interval:      5,
	}
	if err := logger.Initialize("info"); err != nil {
		panic(err)
	}
	configs.Get()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	storage := repository.InitMemStorage()
	agent := accrual.New(configs.Interval, configs.AccrualAddres, logger.Log.Sugar(), storage)
	go agent.Worker(ctx)

	h := handler.NewHandlers(storage, logger.Log.Sugar())
	r := chi.NewRouter()
	r.Use(middleware.StripSlashes)
	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", h.RegisterUser(configs.Key))
		r.Post("/login", h.LoginUser(configs.Key))
		r.Post("/orders", handler.AuthMiddleware(configs.Key, h.SaveOrderHandler()))
		r.Get("/orders", handler.AuthMiddleware(configs.Key, h.GetOrderHandler()))
		r.Get("/balance", handler.AuthMiddleware(configs.Key, h.GetBalanceHandler()))
		r.Post("/balance/withdraw", handler.AuthMiddleware(configs.Key, h.WithdrawtBalanceHandler()))
		r.Get("/withdrawals", handler.AuthMiddleware(configs.Key, h.GetWithdrawalsBalanceHandler()))
	})
	serv := server.New(configs.Net.String(), logger.Log.Sugar(), r)
	if err := serv.Run(); err != nil {
		panic(err)
	}
}
