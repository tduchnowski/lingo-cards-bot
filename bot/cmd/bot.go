package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"lang-learn-bot/handler"
	"lang-learn-bot/telegramapi"
	"log/slog"
	"net/http"
)

type Bot struct {
	telegramapi.User
	token           string
	baseUrl         string
	cmdHandler      handler.CommandHandler
	callbackHandler handler.CallbackHandler
	sem             chan struct{}
}

func (b *Bot) AddCmdHandler(cmdHandler handler.CommandHandler) {
	b.cmdHandler = cmdHandler
}

func (b *Bot) AddCallbackHandler(callbackHandler handler.CallbackHandler) {
	b.callbackHandler = callbackHandler
}

func (b Bot) Start(ctx context.Context) {
	slog.Info("Starting bot")
	updateChan := make(chan telegramapi.Update)
	defer close(updateChan)
	go ListenForUpdates(ctx, updateChan, 8080)
	for {
		select {
		case <-ctx.Done():
			return
		case update := <-updateChan:
			b.handleUpdate(update)
		}
	}
}

func (b Bot) handleUpdate(update telegramapi.Update) {
	b.sem <- struct{}{}
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
		<-b.sem
	}()
}

func CreateBot(telegramConfig TelegramConfig, maxWorkers int) (Bot, error) {
	baseUrl := telegramConfig.BotApiUrl + telegramConfig.BotToken
	botUser, err := ValidateBot(baseUrl)
	if err != nil || SetWebhook(baseUrl, telegramConfig.WebhookUrl) != nil {
		return Bot{}, err
	}
	workerSem := make(chan struct{}, maxWorkers)
	bot := Bot{
		User:    botUser.User,
		token:   telegramConfig.BotToken,
		baseUrl: baseUrl,
		sem:     workerSem,
	}
	return bot, nil
}

func ValidateBot(botBaseUrl string) (telegramapi.UserResponse, error) {
	var botUser telegramapi.UserResponse
	getMeUrl := botBaseUrl + "/getMe"
	res, err := http.Get(getMeUrl)
	if err != nil {
		return botUser, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return botUser, errors.New("/getMe request: " + res.Status)
	}
	body, err := io.ReadAll(res.Body)
	err = json.Unmarshal(body, &botUser)
	if err != nil {
		return botUser, err
	}
	return botUser, nil
}

func SetWebhook(botBaseUrl, webhookUrl string) error {
	// https://api.telegram.org/bot<token>/setWebhook?url=https://myserver.com/webhook
	setWebhookUrl := fmt.Sprintf("%s/setWebhook?url=%s", botBaseUrl, webhookUrl)
	res, err := http.Get(setWebhookUrl)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return errors.New("/getMe request: " + res.Status)
	}
	body, err := io.ReadAll(res.Body)
	fmt.Println(string(body))
	var setWebhookResponse telegramapi.SetWebhookResponse
	err = json.Unmarshal(body, &setWebhookResponse)
	if err != nil {
		return err
	}
	if !setWebhookResponse.Ok || !setWebhookResponse.Result {
		return fmt.Errorf("couldn't set webhook; description:%s", setWebhookResponse.Description)
	}
	return nil
}
