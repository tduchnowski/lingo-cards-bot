package main

import (
	"fmt"
	"os"
)

// const telegramBaseUrl = "http://localhost:8000/bot"

func main() {
	// db init
	dbHost := os.Getenv("BOT_DB_HOST")
	dbPort := os.Getenv("BOT_DB_PORT")
	dbUser := os.Getenv("BOT_DB_USER")
	dbPassword := os.Getenv("BOT_DB_PASSWORD")
	dbName := os.Getenv("BOT_DB_NAME")
	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", dbUser, dbPassword, dbHost, dbPort, dbName)
	fmt.Println(dbUrl)
	db, err := createConnection(dbUrl)
	if err != nil {
		fmt.Println("db fail: ", err)
		return
	}
	defer db.Close()
	// bot token verification
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		fmt.Println("No token provided")
		return
	}
	bot, err := createBot(token)
	if err != nil {
		fmt.Println("error: couldn't create a bot", err)
		return
	}
	// bot handlers
	cmdHandler := NewCommandHandler()
	cmdHandler.AddCommand("/start", start)
	cmdHandler.AddCommand("/menu", menu)
	cmdHandler.AddCommand("/about", about)
	cmdHandler.AddCommand("/help", help)
	bot.commandHandler = cmdHandler
	bot.callbackHandler = CallbackHandler{db: db}
	go bot.start(60)
	bot.handleUpdates()
}
