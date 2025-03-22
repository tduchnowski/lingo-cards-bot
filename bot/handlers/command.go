package handlers

import (
	"fmt"
	"lang-learn-bot/telegramapi"
	"log/slog"

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
	var resp Responder
	switch msg.Text {
	case "/start":
		resp = start(msg)
	case "/menu":
		resp = menu(msg, cmdHandler.db)
	case "/about":
		resp = about(msg)
	case "/help":
		resp = help(msg)
	}
	go func() {
		username := msg.Chat.Username
		if resp != nil {
			slog.Info(fmt.Sprintf("user activity -- command: %s, user: %s", msg.Text, username))
		}
		if username != "" {
			createOrUpdatePrivateChat(cmdHandler.db, username, msg.Chat.Id)
		}
	}()
	if resp != nil {
		return resp
	}
	return SendMsg{}
}

func start(msg telegramapi.Message) Responder {
	text := "Hello\n\nThis bot's purpose is to serve you word cards with translations and examples of sentences in which the word can appear. I made an effort for all the words to be in their root form (lemma).\n\nAfter you pick a language and frequency (most frequent, common, or less frequent out of the words in a bot's database) with a /menu command, the bot will start serving you word cards, with a translation and examples hidden.\n\nIf you are interested in where the words and examples come from, you can choose /about command.\n\nI hope you have fun and learn something new."
	return SendMsg{ChatId: msg.Chat.Id, Text: text}
}

func menu(msg telegramapi.Message, db *pgxpool.Pool) Responder {
	text := "Pick a language of your interest:"
	supportedLangs, err := getSupportedLanguages(db)
	if err != nil {
		return SendMsg{}
	}
	return SendMsg{ChatId: msg.Chat.Id, Text: text, ReplyMarkup: chooseLanguageKeyboardMarkup(supportedLangs)}
}

func about(msg telegramapi.Message) Responder {
	text := "The frequency lists I used were based on the words taken from OpenSubtitles.org. Since those lists take words as they appear in text, they contain many words not in their root form (conjugated verbs etc.) or words that don't add much value for learning (peoples' names for example). I created a lemma database and examples with Chat GPT (gpt-4o-mini model). For every lemma, its count is a sum of counts of the words that have that lemma as its root.\n\nIf you have any suggestions, feel free to message @CosmicBuddy."
	return SendMsg{ChatId: msg.Chat.Id, Text: text}
}

func help(msg telegramapi.Message) Responder {
	text := "Commands:\n\n/menu - choose a language and frequency and start exploring words\n\n/about - information about this bot"
	return SendMsg{ChatId: msg.Chat.Id, Text: text}
}
