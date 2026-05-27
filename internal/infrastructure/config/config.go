package config

import (
	"flag"
	"os"
	"sync"
	"time"
)

type config struct {
	ServerAddress  string
	DatabaseURI    string
	AccrualAddress string
	TokenExp       time.Duration
	SecretKey      []byte
	RateLimit      int
	BufferSize     int
	MigrationsDir  string
}

var (
	cfg  config
	once sync.Once
)

func Get() config {
	once.Do(func() {
		cfg.SecretKey = []byte("supersecretkey")
		cfg.TokenExp = time.Hour * 24
		cfg.RateLimit = 100
		cfg.BufferSize = 100
		cfg.MigrationsDir = "migrations"
		flag.StringVar(&cfg.ServerAddress, "a", ":8080", "address and port to run server")
		flag.StringVar(&cfg.DatabaseURI, "d", "", "DSN for db")
		flag.StringVar(&cfg.AccrualAddress, "r", "", "address for accrual system")

		flag.Parse()
		if envServerAddress := os.Getenv("RUN_ADDRESS"); envServerAddress != "" {
			cfg.ServerAddress = envServerAddress
		}
		if envAccrualAddress := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccrualAddress != "" {
			cfg.AccrualAddress = envAccrualAddress
		}
		if envDBDSN := os.Getenv("DATABASE_URI"); envDBDSN != "" {
			cfg.DatabaseURI = envDBDSN
		}
	})
	return cfg
}
