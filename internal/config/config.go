package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/caarlos0/env/v11"
)

type NetAddress struct {
	Host string
	Port int
}

type ConfigServer struct {
	Net           NetAddress `env:"RUN_ADDRESS"`
	DatabaseUri   string     `env:"DATABASE_URI"`
	AccrualAddres string     `env:"ACCRUAL_SYSTEM_ADDRESS"`
	Key           string     `env:"SECRET_KEY"`
	Interval      int        `env:"INTERVAL"`
}

func (n NetAddress) String() string {
	return n.Host + ":" + strconv.Itoa(n.Port)
}

func (n *NetAddress) UnmarshalText(adr []byte) error {
	return n.Set(string(adr))
}
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

func (s *ConfigServer) Get() {
	f := flag.NewFlagSet("Run server", flag.ContinueOnError)
	f.Var(&s.Net, "a", "Net address host:port")
	f.StringVar(&s.DatabaseUri, "d", s.DatabaseUri, "DATABASE_URI")
	f.StringVar(&s.AccrualAddres, "r", s.AccrualAddres, "ACCRUAL_SYSTEM_ADDRESS")
	f.StringVar(&s.Key, "k", s.Key, "SECRET_KEY")
	f.IntVar(&s.Interval, "i", s.Interval, "INTERVAL")
	err := f.Parse(os.Args[1:])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = env.Parse(s)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
