package main

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/danilov-go/gophermart/internal/accrual"
	"github.com/danilov-go/gophermart/internal/config"
	"github.com/danilov-go/gophermart/internal/handler"
	"github.com/danilov-go/gophermart/internal/logger"
	"github.com/danilov-go/gophermart/internal/repository/db"
	"github.com/danilov-go/gophermart/internal/repository/memory"
	"github.com/danilov-go/gophermart/internal/server"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/sync/errgroup"
)

func main() {
	configs := config.ConfigServer{
		Net: config.NetAddress{
			Host: "localhost",
			Port: 8080,
		},
		DatabaseUri:   "",
		AccrualAddres: "",
		Key:           "secret_key",
		Interval:      5,
		RetryDefault:  60,
	}
	err := logger.Initialize("info")
	if err != nil {
		panic(err)
	}
	err = configs.Get()
	if err != nil {
		logger.Log.Sugar().Fatalw("ошибка загрузки конфигурации", "error", err)
	}
	g, ctx := errgroup.WithContext(context.Background())
	var storage handler.Storage
	dbStorage, err := db.InitDB(configs.DatabaseUri)
	if err != nil {
		logger.Log.Sugar().Infow("ошибка инициализации базы данных", "error", err)
		storage = memory.InitMemStorage()
	} else {
		pingCtx, pingCancel := context.WithTimeout(ctx, 3*time.Second)
		err = dbStorage.Ping(pingCtx)
		pingCancel()
		if err != nil {
			logger.Log.Sugar().Infow("база данных недоступна", "error", err)
			storage = memory.InitMemStorage()
		} else {
			storage = handler.NewErrorMiddleware(dbStorage, 2*time.Second, 2*time.Second)
		}
	}
	g.Go(func() error {
		agent := accrual.New(configs.Interval, configs.RetryDefault, configs.AccrualAddres, logger.Log.Sugar(), storage)
		return agent.Worker(ctx)
	})
	h := handler.NewHandlers(storage, logger.Log.Sugar())
	r := chi.NewRouter()
	r.Use(middleware.StripSlashes)
	r.Use(handler.RequestLogger(logger.Log))
	r.Use(handler.GzipMiddleware)
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
	g.Go(func() error {
		if err := serv.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})
	g.Go(func() error {
		<-ctx.Done()
		stopCtx, stopCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer stopCancel()
		return serv.Stop(stopCtx)
	})
	if err := g.Wait(); err != nil {
		if !errors.Is(ctx.Err(), context.Canceled) {
			logger.Log.Sugar().Fatalw("ошибка в фоновом процессе", "error", err)
		}
	}
}
