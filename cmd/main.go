package main

import (
	"fmt"
	"log/slog"
	"os"

	_ "github.com/a-h/templ" // needed to prevent "go mod tidy" from breaking templ functions
	"github.com/joho/godotenv"

	"github.com/Ewan-Greer09/remote-colab/internal/server"
)

func main() {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	s := server.NewServer(fmt.Sprintf("0.0.0.0:%s", port), server.NewHandler())
	if err := s.ListenAndServe(); err != nil {
		slog.Error("Listen and Serve", "err", err)
		os.Exit(1)
	}
}
