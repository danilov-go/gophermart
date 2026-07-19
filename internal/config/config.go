// Package config управляет конфигурацией приложения.
package config

import (
	"errors"
	"flag"
	"os"
	"strconv"
	"strings"

	"github.com/caarlos0/env/v11"
)

// NetAddress определяет адрес сервера.
type NetAddress struct {
	Host string
	Port int
}

// ConfigServer определяет конфигурацию приложения.
type ConfigServer struct {
	// Net содержит сетевой адрес для запуска сервера.
	Net NetAddress `env:"RUN_ADDRESS"`
	// DatabaseUri содержит строку подключения к базе данных.
	DatabaseUri string `env:"DATABASE_URI"`
	// AccrualAddres содержит адрес внешней системы начисления баллов.
	AccrualAddres string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	// Key содержит секретный ключ для подписи JWT-токенов.
	Key string `env:"SECRET_KEY"`
	// Interval содержит интервал опроса внешней системы в секундах.
	Interval int `env:"INTERVAL"`
	// RetryDefault содержит дефолтное время ожидания в секундах при ошибке 429.
	RetryDefault int `env:"RETRY_DEFAULT"`
}

// String возвращает строковое представление сетевого адреса в формате host:port.
func (n NetAddress) String() string {
	return n.Host + ":" + strconv.Itoa(n.Port)
}

// UnmarshalText десериализует сетевой адрес из текстового формата для библиотеки env.
func (n *NetAddress) UnmarshalText(adr []byte) error {
	return n.Set(string(adr))
}

// Set парсит строку в формате host:port и валидирует ее для пакета flag.
func (n *NetAddress) Set(s string) error {
	hp := strings.Split(s, ":")
	if len(hp) != 2 {
		return errors.New("Need address in a form host:port")
	}
	port, err := strconv.Atoi(hp[1])
	if err != nil {
		return err
	}
	n.Host = hp[0]
	n.Port = port
	return nil
}

// Get парсит конфигурацию приложения.
func (s *ConfigServer) Get() error {
	f := flag.NewFlagSet("Run server", flag.ContinueOnError)
	f.Var(&s.Net, "a", "Net address host:port")
	f.StringVar(&s.DatabaseUri, "d", s.DatabaseUri, "DATABASE_URI")
	f.StringVar(&s.AccrualAddres, "r", s.AccrualAddres, "ACCRUAL_SYSTEM_ADDRESS")
	f.StringVar(&s.Key, "k", s.Key, "SECRET_KEY")
	f.IntVar(&s.Interval, "i", s.Interval, "INTERVAL")
	f.IntVar(&s.RetryDefault, "t", s.RetryDefault, "RETRY_DEFAULT")
	err := f.Parse(os.Args[1:])
	if err != nil {
		return err
	}
	err = env.Parse(s)
	if err != nil {
		return err
	}
	return nil
}
