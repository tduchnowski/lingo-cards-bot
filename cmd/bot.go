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

type SendMsgOpts struct {
	ChatId      int64                 `json:"chat_id"`
	Text        string                `json:"text"`
	ReplyMarkup *InlineKeyboardMarkup `json:"reply_markup,omitempty"` // has to be a pointer for json serializer to skip it if empty
}

type CallbackHandler func(CallbackQuery) SendMsgOpts
type MsgHandler func(Message) SendMsgOpts

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
			switch {
			case update.CallbackQuery.Id != "" && b.callbackHandler != nil:
				reply := b.callbackHandler(update.CallbackQuery)
				b.sendMessage(reply)
				b.answerCallbackQuery(update.CallbackQuery.Id)
			case update.Msg.Id != 0:
				f, ok := b.commands[update.Msg.Text]
				if ok {
					reply := f(update.Msg)
					b.sendMessage(reply)
				}
			}
		}()
	}
}

func (b Bot) sendMessage(parameters SendMsgOpts) {
	data, err := json.Marshal(parameters)
	if err != nil {
		fmt.Println("sending failed on serializing message object to json")
	}
	url := fmt.Sprintf("%s/sendMessage", b.baseUrl)
	client := &http.Client{}
	res, err := client.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		fmt.Printf("error on post request: %s\n", err)
		// TODO: do smth if Telegram gives an error back
		var j interface{}
		err = json.NewDecoder(res.Body).Decode(&j)
		fmt.Println("send message response from telegram: ", j)
	}
}

func (b Bot) answerCallbackQuery(callbackQueryId string) {
	// this function needs to be called to stop buttons from
	// blinking after they're pressed by the user
	url := fmt.Sprintf("%s/answerCallbackQuery", b.baseUrl)
	data := fmt.Sprintf("{\"callback_query_id\":\"%s\"}", callbackQueryId)
	client := &http.Client{}
	res, _ := client.Post(url, "application/json", bytes.NewBufferString(data)) //I dont even care about this error, no big deal
	res.Body.Close()
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
