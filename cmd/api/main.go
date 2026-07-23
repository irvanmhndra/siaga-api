package main

import (
	"log/slog"
	"os"

	"github.com/irvanmhndra/siaga-api/config"
	"github.com/irvanmhndra/siaga-api/internal/app"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.Load()

	application, err := app.New(cfg)
	if err != nil {
		slog.Error("Failed to initialize app", "error", err)
		os.Exit(1)
	}
	defer application.Close()

	application.Run()
}
