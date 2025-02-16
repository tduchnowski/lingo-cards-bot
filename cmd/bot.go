package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type CallbackHandler func(CallbackQuery) Responder
type MsgHandler func(Message) Responder

type Bot struct {
	User
	token           string
	baseUrl         string
	updates         chan Update // a channel for passing updates to the handler
	lastUpdateId    int64       // id of last processed update, needed for getUpdates() offset parameter
	commands        map[string]MsgHandler
	callbackHandler CallbackHandler
}

func (b Bot) AddCommand(name string, handler MsgHandler) {
	// simply add a command to Bot.commands map
	b.commands[name] = handler
}

func (b Bot) start(timeout int) {
	// start polling the Telegram API for updates using the /getUpdates method
	for {
		client := http.Client{Timeout: time.Duration(timeout) * time.Second}
		urlQuery := fmt.Sprintf("%s/getUpdates?timeout=%d&offset=%d", b.baseUrl, timeout, b.lastUpdateId+1)
		res, err := client.Get(urlQuery)
		if err != nil {
			// TODO: make it wait a few seconds and retry
			fmt.Println(err)
			continue
		}
		if res.StatusCode != 200 {
			fmt.Println(errors.New("/getUpdates request: " + res.Status))
		}
		body, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Println(err)
		}
		var ur UpdateResponse
		err = json.Unmarshal(body, &ur)
		if err != nil {
			fmt.Println(err)
		}
		for _, update := range ur.Updates {
			b.updates <- update
			b.lastUpdateId = update.Id
		}
		res.Body.Close()
	}
}

func (b Bot) handleUpdates() {
	for update := range b.updates {
		go func() {
			var response Responder
			switch {
			case update.CallbackQuery.Id != "" && b.callbackHandler != nil:
				response = b.callbackHandler(update.CallbackQuery)
				go update.CallbackQuery.answer(b.baseUrl)
			case update.Msg.Id != 0:
				commandHandler, ok := b.commands[update.Msg.Text]
				if ok {
					response = commandHandler(update.Msg)
				} else {
					response = SendMsg{}
				}
			}
			response.Respond(b.baseUrl)
		}()
	}
}

func createBot(token string) (Bot, error) {
	baseUrl := os.Getenv("TG_BOT_URL") + token
	getMeUrl := baseUrl + "/getMe"
	res, err := http.Get(getMeUrl)
	if err != nil {
		return Bot{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return Bot{}, errors.New("/getMe request: " + res.Status)
	}
	body, err := io.ReadAll(res.Body)
	var ur UserResponse
	err = json.Unmarshal(body, &ur)
	if err != nil {
		return Bot{}, err
	}
	updates := make(chan Update)
	commands := make(map[string]MsgHandler)
	bot := Bot{
		User:     ur.User,
		token:    token,
		baseUrl:  baseUrl,
		updates:  updates,
		commands: commands}
	return bot, nil
}
