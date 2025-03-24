package handler

import (
	"context"
	"fmt"
	"lang-learn-bot/database"
	"lang-learn-bot/telegramapi"
	"log/slog"
	"os"
	"testing"
)

var cmdHandler CommandHandler
var callbackHandler CallbackHandler

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
	cmdHandler = NewCommandHandler(db)
	callbackHandler = NewCallbackHandler(db)
	exitVal := m.Run()
	os.Exit(exitVal)
}

// test responses to text messages that
// are not bot commands
func TestNonCommand(t *testing.T) {
	msg := telegramapi.Message{
		Id:   1,
		Text: "",
	}
	reply := cmdHandler.GetResponder(msg)
	if sm, ok := reply.(SendMsg); ok {
		if sm.Text != "" {
			t.Error("reply text for an empty message should be empty")
		}
	}
	msg.Text = "noncommand"
	reply = cmdHandler.GetResponder(msg)
	if sm, ok := reply.(SendMsg); ok {
		if sm.Text != "" {
			t.Error("reply text for a message that's not a command should be empty")
		}
	}
}

// test responses when the message text
// is a command but not the right one
// for the bot
func TestWrongCommand(t *testing.T) {
	msg := telegramapi.Message{
		Id:   1,
		Text: "/some_command",
	}
	reply := cmdHandler.GetResponder(msg)
	if sm, ok := reply.(SendMsg); ok {
		if sm.Text != "" {
			t.Error("reply text for a wrong command should be empty")
		}
	}
}

// test responses to weird callbacks
// and callback data
func TestMenuCallback(t *testing.T) {

}

// test formatting of different WordEntry
// structs
func TestNoSentenceFormatting(t *testing.T) {
	sentences := ""
	we := database.WordEntry{
		Lemma:        "lemma",
		LemmaMeaning: "meaning",
		Sentences:    &sentences,
	}
	formatted := formatWordMsg(we)
	correctRes := "\n*lemma*\n\n||*meaning:* meaning||\n\n"
	if formatted != correctRes {
		t.Errorf("formatted: %s != %s\n", formatted, correctRes)
	}
	we.Sentences = nil
	formatted = formatWordMsg(we)
	if formatted != correctRes {
		t.Errorf("formatted: %s != %s\n", formatted, correctRes)
	}
}

func TestMenuCommand(t *testing.T) {

}
