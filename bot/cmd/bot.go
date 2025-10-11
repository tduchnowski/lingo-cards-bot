package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"lang-learn-bot/handler"
	"lang-learn-bot/telegramapi"
	"log/slog"
	"net/http"
	"time"
)

type Bot struct {
	telegramapi.User
	token           string
	baseUrl         string
	cmdHandler      handler.CommandHandler
	callbackHandler handler.CallbackHandler
}

func (b *Bot) AddCmdHandler(cmdHandler handler.CommandHandler) {
	b.cmdHandler = cmdHandler
}

func (b *Bot) AddCallbackHandler(callbackHandler handler.CallbackHandler) {
	b.callbackHandler = callbackHandler
}

func (b Bot) Start(timeout int) {
	slog.Info("Starting the bot")
	lastId := int64(0)
	client := http.Client{Timeout: time.Duration(timeout) * time.Second}
	for {
		urlQuery := fmt.Sprintf("%s/getUpdates?timeout=%d&offset=%d", b.baseUrl, timeout, lastId+1)
		slog.Debug("fetching updates from Telegram")
		res, clientErr := client.Get(urlQuery)
		if clientErr != nil {
			slog.Error("error during update fetching" + clientErr.Error())
			time.Sleep(5 * time.Second)
			continue
		}
		if res.StatusCode != 200 {
			slog.Error(fmt.Sprintf("telegram responded with status: %s", res.Status))
			time.Sleep(5 * time.Second)
			continue
		}
		body, readErr := io.ReadAll(res.Body)
		if readErr != nil {
			slog.Error(readErr.Error())
			time.Sleep(5 * time.Second)
			continue
		}
		lastId = b.handleUpdateResponse(body)
		res.Body.Close()
	}
}

func (b Bot) handleUpdateResponse(updateBody []byte) int64 {
	var ur telegramapi.UpdateResponse
	err := json.Unmarshal(updateBody, &ur)
	if err != nil {
		slog.Error("handleUpdateResponse - " + err.Error())
		return 0
	}
	slog.Debug(fmt.Sprintf("received %d updates", len(ur.Updates)))
	var lastUpdateId int64
	for _, update := range ur.Updates {
		go func() {
			var reply handler.Responder
			switch update.GetUpdateType() {
			case "message":
				reply = b.cmdHandler.GetResponder(update.Msg)
			case "callback":
				reply = b.callbackHandler.GetResponder(update.CallbackQuery)
			default:
				reply = handler.SendMsg{}
			}
			reply.Respond(b.baseUrl)
		}()
		lastUpdateId = update.Id
	}
	return lastUpdateId
}

// verifies token and then creates Bot struct and returns it
func CreateBot(telegramConfig TelegramConfig) (Bot, error) {
	baseUrl := telegramConfig.BotApiUrl + telegramConfig.BotToken
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
	var ur telegramapi.UserResponse
	err = json.Unmarshal(body, &ur)
	if err != nil {
		return Bot{}, err
	}
	bot := Bot{
		User:    ur.User,
		token:   telegramConfig.BotToken,
		baseUrl: baseUrl,
	}
	return bot, nil
}
