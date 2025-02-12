package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Bot struct {
	User
	token        string
	baseUrl      string
	updates      chan Update // a channel for passing updates to the handler
	lastUpdateId int64       // id of last processed update, needed for getUpdates() offset parameter
	commands     map[string]MsgHandler
	// queryCallbacks map[string]CallbackHandler
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
		}
		defer res.Body.Close()
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
	}
}

func (b Bot) handleUpdates() error {
	for update := range b.updates {
		switch {
		case update.CallbackQuery.Id != "":
			fmt.Println("Its a callback")
		case update.Msg.Id != 0:
			f, ok := b.commands[update.Msg.Text]
			if ok {
				go func() {
					reply := f(update.Msg)
					b.sendMessage(reply)
				}()
			}
		}
	}
	return nil
}

func (b Bot) sendMessage(parameters SendMsgOpts) {
	data, err := json.Marshal(parameters)
	if err != nil {
		fmt.Println("sending failed on serializing message object to json")
	}
	url := fmt.Sprintf("%s/sendMessage", b.baseUrl)
	fmt.Println(bytes.NewReader(data))
	client := &http.Client{}
	res, err := client.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		fmt.Printf("error on post request: %s", err)
	}

	var j interface{}
	err = json.NewDecoder(res.Body).Decode(&j)
	fmt.Println("send message response from telegram: ", j)
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

type CallbackHandler func(CallbackQuery) SendMsgOpts
type MsgHandler func(Message) SendMsgOpts
type SendMsgOpts struct {
	ChatId      int64                 `json:"chat_id"`
	Text        string                `json:"text"`
	ReplyMarkup *InlineKeyboardMarkup `json:"reply_markup,omitempty"` // has to be a pointer for json serializer to skip it if empty
}
