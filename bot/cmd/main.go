package main

import (
	"fmt"
	"lang-learn-bot/database"
	"log/slog"
	"os"
)

func main() {
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		slog.Error("No token provided. Exiting.")
		return
	}
	// db init
	dbHost := os.Getenv("BOT_DB_HOST")
	dbPort := os.Getenv("BOT_DB_PORT")
	dbUser := os.Getenv("BOT_DB_USER")
	dbPassword := os.Getenv("BOT_DB_PASSWORD")
	dbName := os.Getenv("BOT_DB_NAME")
	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", dbUser, dbPassword, dbHost, dbPort, dbName)
	slog.Info(fmt.Sprintf("Initializing the connection to the database: host=%s, port=%s, user=%s, database=%s", dbHost, dbPort, dbUser, dbName))
	//TODO: ping the db to be sure
	db, err := database.CreateConnection(dbUrl)
	if err != nil {
		slog.Error("Database connection failed")
		slog.Error(err.Error())
		slog.Error("Exiting")
		return
	}
	defer db.Close()
	bot, err := createBot(token)
	if err != nil {
		// TODO: actually wrong token is not the only reason why this may
		// return an error. could be simply that Telegram is down.
		// so make it retry in case of a timout
		slog.Error("Wrong token. Exiting")
		slog.Error(err.Error())
		return
	}
	slog.Info("Bot ready to fetch updates")
	bot.run(db, 60)
}
