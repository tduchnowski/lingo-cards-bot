package handlers

import (
	"lang-learn-bot/telegramapi"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MsgReply func(telegramapi.Message) Responder

type CommandHandler struct {
	db *pgxpool.Pool
}

func NewCommandHandler(db *pgxpool.Pool) CommandHandler {
	return CommandHandler{db: db}
}

func (cmdHandler CommandHandler) GetResponder(msg telegramapi.Message) Responder {
	switch msg.Text {
	case "/start":
		return start(msg, cmdHandler.db)
	case "/menu":
		return menu(msg, cmdHandler.db)
	case "/about":
		return about(msg)
	case "/help":
		return help(msg)
	}
	return SendMsg{}
}

func start(msg telegramapi.Message, db *pgxpool.Pool) Responder {
	text := "Welcome"
	username := msg.Chat.Username
	if username != "" {
		createOrUpdatePrivateChat(db, username, msg.Chat.Id)
	}
	return SendMsg{ChatId: msg.Chat.Id, Text: text}
}

func menu(msg telegramapi.Message, db *pgxpool.Pool) Responder {
	text := "Choose a language"
	supportedLangs, err := getSupportedLanguages(db)
	if err != nil {
		return SendMsg{}
	}
	return SendMsg{ChatId: msg.Chat.Id, Text: text, ReplyMarkup: chooseLanguageKeyboardMarkup(supportedLangs)}
}

func about(msg telegramapi.Message) Responder {
	text := "about this bot"
	return SendMsg{ChatId: msg.Chat.Id, Text: text}
}

func help(msg telegramapi.Message) Responder {
	text := "here are all the commands"
	return SendMsg{ChatId: msg.Chat.Id, Text: text}
}
