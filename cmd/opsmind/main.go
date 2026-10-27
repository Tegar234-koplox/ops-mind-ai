package main

import (
	"context"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"net/http"
	"opsmind/internal/agent"
	"opsmind/internal/config"
	"opsmind/internal/detection"
	"opsmind/internal/domain"
	"opsmind/internal/httpapi"
	"opsmind/internal/seed"
	"opsmind/internal/store"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if e := run(); e != nil {
		slog.Error("opsmind stopped", "reason", e.Error())
		os.Exit(1)
	}
}
func run() error {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(log)
	command := "serve"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}
	if command == "hash-password" {
		value := os.Getenv("OPERATOR_PASSWORD")
		if len(value) < 12 {
			return fmt.Errorf("OPERATOR_PASSWORD minimal 12 karakter")
		}
		hash, e := bcrypt.GenerateFromPassword([]byte(value), bcrypt.DefaultCost)
		if e != nil {
			return e
		}
		fmt.Println(string(hash))
		return nil
	}
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return fmt.Errorf("DATABASE_URL wajib diisi")
	}
	if command == "migrate" {
		return store.Migrate(dsn)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if e := store.Migrate(dsn); e != nil {
		return e
	}
	s, e := store.Open(ctx, dsn)
	if e != nil {
		return e
	}
	defer s.Pool.Close()
	if command == "seed" {
		created, e := seed.Run(ctx, s)
		if e != nil {
			return e
		}
		log.Info("synthetic dataset ready", "created", created, "seed", 20261027)
		return nil
	}
	c, e := config.Load()
	if e != nil {
		return e
	}
	if command == "detect" {
		engine := detection.Engine{Store: s, Config: c}
		v, e := engine.Execute(ctx, "", domain.AsOf)
		if e != nil {
			return e
		}
		fmt.Println(string(store.JSON(v)))
		return nil
	}
	if command != "serve" {
		return fmt.Errorf("command: serve | migrate | seed | detect | hash-password")
	}
	if c.SeedIfEmpty {
		if _, e = seed.Run(ctx, s); e != nil {
			return e
		}
	}
	api, e := httpapi.New(s, c, log)
	if e != nil {
		return e
	}
	runner := &agent.Runner{Store: s, Config: c, Logger: log}
	workerDone := make(chan struct{})
	go func() { defer close(workerDone); runner.Work(ctx) }()
	server := &http.Server{Addr: ":" + c.Port, Handler: api, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 16384}
	errCh := make(chan error, 1)
	go func() { errCh <- server.ListenAndServe() }()
	log.Info("opsmind listening", "port", c.Port, "mode", c.Mode, "synthetic", true)
	select {
	case <-ctx.Done():
	case e = <-errCh:
		if e != nil && e != http.ErrServerClosed {
			cancel()
			<-workerDone
			return e
		}
	}
	cancel()
	shutdownCtx, stop := context.WithTimeout(context.Background(), 10*time.Second)
	defer stop()
	e = server.Shutdown(shutdownCtx)
	<-workerDone
	return e
}
