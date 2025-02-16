package main

import (
	"fmt"
	"os"
)

// const telegramBaseUrl = "http://localhost:8000/bot"

func main() {
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		panic("No token provided")
	}
	bot, err := createBot(token)
	if err != nil {
		fmt.Println("error: couldn't create a bot", err)
		return
	}
	cmdHandler := NewCommandHandler()
	cmdHandler.AddCommand("/start", start)
	cmdHandler.AddCommand("/menu", menu)
	cmdHandler.AddCommand("/about", about)
	cmdHandler.AddCommand("/help", help)
	bot.commandHandler = cmdHandler
	bot.callbackHandler = CallbackHandler{}
	go bot.start(60)
	bot.handleUpdates()
}
