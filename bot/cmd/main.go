package main

import (
	"context"
	"fmt"
	"lang-learn-bot/database"
	"lang-learn-bot/handler"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := LoadConfig("./cmd")
	if err != nil {
		slog.Error(err.Error())
	}
	if !ValidateConfig(&cfg) {
		slog.Error("Invalid config, field/s missing or empty")
	}
	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", cfg.Db.User, cfg.Db.Password, cfg.Db.Host, cfg.Db.Port, cfg.Db.Name)
	slog.Info(fmt.Sprintf("initializing the connection to the database: host=%s, port=%d, user=%s, database=%s", cfg.Db.Host, cfg.Db.Port, cfg.Db.User, cfg.Db.Name))
	db, err := database.CreateConnection(dbUrl)
	if err != nil {
		slog.Error("database connection failed")
		slog.Error(err.Error())
		slog.Info("exiting")
		return
	}
	defer db.Close()
	bot, err := CreateBot(cfg.Telegram)
	if err != nil {
		slog.Error(err.Error())
		slog.Info("exiting")
		return
	}
	bot.AddCmdHandler(handler.NewCommandHandler(db))
	bot.AddCallbackHandler(handler.NewCallbackHandler(db))
	ctx, cancel := context.WithCancel(context.Background())
	go bot.Start(ctx, 60)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	cancel()
	slog.Info("stopping application")
}
