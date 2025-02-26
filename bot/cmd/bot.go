package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"lang-learn-bot/handlers"
	"lang-learn-bot/telegramapi"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Bot struct {
	telegramapi.User
	token              string
	baseUrl            string
	updateResponseChan chan []byte
	lastUpdateIdChan   chan int64
}

func (b Bot) run(db *pgxpool.Pool, timeout int) {
	var wg sync.WaitGroup
	wg.Add(2)
	b.startFetchingUpdates(timeout, &wg)
	b.handleUpdateResponses(db, &wg)
	wg.Wait()
}

func (b *Bot) startFetchingUpdates(timeout int, wg *sync.WaitGroup) {
	b.updateResponseChan = make(chan []byte)
	b.lastUpdateIdChan = make(chan int64)
	lastUpdateId := int64(0)
	go func(lastId int64) {
		defer wg.Done()
		for {
			client := http.Client{Timeout: time.Duration(timeout) * time.Second}
			urlQuery := fmt.Sprintf("%s/getUpdates?timeout=%d&offset=%d", b.baseUrl, timeout, lastId+1)
			slog.Info("Fetching updates from Telegram")
			res, clientErr := client.Get(urlQuery)
			if clientErr != nil {
				slog.Error("Error during update fetching" + clientErr.Error())
				time.Sleep(5 * time.Second)
				continue
			}
			if res.StatusCode != 200 {
				slog.Error(fmt.Sprintf("Telegram responded with status: %s", res.Status))
				time.Sleep(5 * time.Second)
				continue
			}
			body, readErr := io.ReadAll(res.Body)
			if readErr != nil {
				slog.Error(readErr.Error())
				time.Sleep(5 * time.Second)
				continue
			}
			b.updateResponseChan <- body
			lastId = <-b.lastUpdateIdChan
			res.Body.Close()
		}
	}(lastUpdateId)
}

func (b *Bot) handleUpdateResponses(db *pgxpool.Pool, wg *sync.WaitGroup) {
	cmdHandler := handlers.NewCommandHandler(db)
	callbackHandler := handlers.NewCallbackHandler(db)
	go func() {
		defer wg.Done()
		var lastProcessedUpdateId int64 = int64(0)
		for updateBody := range b.updateResponseChan {
			var ur telegramapi.UpdateResponse
			err := json.Unmarshal(updateBody, &ur)
			if err != nil {
				slog.Error(err.Error())
				continue
			}
			slog.Info(fmt.Sprintf("Received %d updates", len(ur.Updates)))
			for _, update := range ur.Updates {
				lastProcessedUpdateId = update.Id
				go func() {
					var reply handlers.Responder
					switch update.GetUpdateType() {
					case "message":
						reply = cmdHandler.GetResponder(update.Msg)
					case "callback":
						reply = callbackHandler.GetResponder(update.CallbackQuery)
					default:
						reply = handlers.SendMsg{}
					}
					reply.Respond(b.baseUrl)
				}()
			}
			b.lastUpdateIdChan <- lastProcessedUpdateId
		}
	}()
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
	var ur telegramapi.UserResponse
	err = json.Unmarshal(body, &ur)
	if err != nil {
		return Bot{}, err
	}
	bot := Bot{
		User:    ur.User,
		token:   token,
		baseUrl: baseUrl,
	}
	return bot, nil
}
