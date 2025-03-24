package main

import (
	"fmt"
	"lang-learn-bot/database"
	"lang-learn-bot/handler"
	"log/slog"
	"os"
)

func main() {
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		slog.Error("no token provided. exiting.")
		return
	}
	// db init
	dbHost := os.Getenv("BOT_DB_HOST")
	dbPort := os.Getenv("BOT_DB_PORT")
	dbUser := os.Getenv("BOT_DB_USER")
	dbPassword := os.Getenv("BOT_DB_PASSWORD")
	dbName := os.Getenv("BOT_DB_NAME")
	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", dbUser, dbPassword, dbHost, dbPort, dbName)
	slog.Info(fmt.Sprintf("initializing the connection to the database: host=%s, port=%s, user=%s, database=%s", dbHost, dbPort, dbUser, dbName))
	//TODO: ping the db to be sure
	db, err := database.CreateConnection(dbUrl)
	if err != nil {
		slog.Error("database connection failed")
		slog.Error(err.Error())
		slog.Error("exiting")
		return
	}
	defer db.Close()
	bot, err := createBot(token)
	if err != nil {
		slog.Error(err.Error())
		slog.Error("exiting")
		return
	}
	bot.addCmdHandler(handler.NewCommandHandler(db))
	bot.addCallbackHandler(handler.NewCallbackHandler(db))
	slog.Info("bot ready to fetch updates")
	bot.start(60)
}
