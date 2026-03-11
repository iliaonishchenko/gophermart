package main

import (
	"flag"
	"github.com/iliaonishchenko/gophermart/internal/config"
)

func parseFlags(cfg *config.Config) {
	var serverAddr string   // -a
	var databaseAddr string // -d
	var accrualAddr string  // -r

	flag.StringVar(&serverAddr, "a", "localhost:8080", "server address")
	flag.StringVar(&databaseAddr, "d", "localhost:27017", "database address")
	flag.StringVar(&accrualAddr, "r", "", "accrual address")

	flag.Parse()

	if cfg.ServerAddr == "" {
		cfg.ServerAddr = serverAddr
	}
	if cfg.DatabaseURI == "" {
		cfg.DatabaseURI = databaseAddr
	}
	if cfg.AccrualAddr == "" {
		cfg.AccrualAddr = accrualAddr
	}
}
