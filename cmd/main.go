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
	bot.AddCommand("/start", start)
	bot.AddCommand("/menu", menu)
	bot.AddCommand("/about", about)
	bot.AddCommand("/help", help)
	bot.callbackHandler = menuCallback
	go bot.start(60)
	bot.handleUpdates()
}
