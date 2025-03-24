// the point here is to simulate hundreds or thousands simultaneous requests
// and see how big is a memory consumption and how much time it takes to process
package benchmark_test

import (
	"context"
	"encoding/json"
	"fmt"
	"lang-learn-bot/database"
	"lang-learn-bot/handler"
	"lang-learn-bot/telegramapi"
	"log/slog"
	"math/rand"
	"os"
	"sync"

	"testing"
)

var cmdHandler handler.CommandHandler
var clbckHandler handler.CallbackHandler

var updates []telegramapi.Update

func TestMain(m *testing.M) {
	slog.SetLogLoggerLevel(slog.LevelError)
	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", "postgres", "postgres", "localhost", "5432", "bot")
	db, err := database.CreateConnection(dbUrl)
	if err != nil {
		return
	}
	defer db.Close()
	err = db.Ping(context.Background())
	if err != nil {
		return
	}
	cmdHandler = handler.NewCommandHandler(db)
	clbckHandler = handler.NewCallbackHandler(db)
	updates = generateUpdates(100)
	exitVal := m.Run()
	os.Exit(exitVal)
}

var reply handler.Responder

func BenchmarkHandlers(b *testing.B) {
	wg := sync.WaitGroup{}
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			update := updates[i%len(updates)]
			switch update.GetUpdateType() {
			case "message":
				reply = cmdHandler.GetResponder(update.Msg)
			case "callback":
				reply = clbckHandler.GetResponder(update.CallbackQuery)
			default:
				reply = handler.SendMsg{}
			}
		}()
	}
	wg.Wait()
}

func generateUpdates(size int) []telegramapi.Update {
	updateSl := make([]telegramapi.Update, size)
	for i := 0; i < size; i++ {
		var u telegramapi.Update
		if i%2 == 0 {
			u = randomMessageUpdate()
		} else {
			u = randomCallbackQueryUpdate()
		}
		updateSl[i] = u
	}
	return updateSl
}

func randomMessageUpdate() telegramapi.Update {
	msg := telegramapi.Message{
		Id:        1,
		MsgThread: 2,
		Text:      "/menu",
		Chat: telegramapi.Chat{
			Id:       9209,
			Username: "TB",
		},
	}
	return telegramapi.Update{
		Id:            1,
		Msg:           msg,
		CallbackQuery: telegramapi.CallbackQuery{},
	}
}

func randomCallbackQueryUpdate() telegramapi.Update {
	languages := []string{"pl", "ru", "es", "pt", "it"}
	data := handler.MenuCallbackData{
		Stage:      uint8(rand.Intn(4)),
		Difficulty: uint8(rand.Intn(3)),
		Language:   languages[rand.Intn(len(languages))],
	}
	d, err := json.Marshal(data)
	if err != nil {
		return telegramapi.Update{}
	}
	clbck := telegramapi.CallbackQuery{
		Id: "30",
		Msg: telegramapi.Message{
			Id: 3,
			Chat: telegramapi.Chat{
				Id: 33,
			},
		},
		Data: string(d),
	}
	return telegramapi.Update{
		Id:            2,
		Msg:           telegramapi.Message{},
		CallbackQuery: clbck,
	}
}
