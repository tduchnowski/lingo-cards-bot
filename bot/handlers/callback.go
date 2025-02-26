package handlers

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"lang-learn-bot/database"
	"lang-learn-bot/telegramapi"
	"log/slog"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MenuCallbackData struct {
	Stage      uint8  `json:"stage"`
	Language   string `json:"language"`
	Difficulty uint8  `json:"difficulty"`
}

type CallbackHandler struct {
	db *pgxpool.Pool
}

func NewCallbackHandler(db *pgxpool.Pool) CallbackHandler {
	return CallbackHandler{db}
}

func (callbackHandler CallbackHandler) GetResponder(cq telegramapi.CallbackQuery) Responder {
	responders := Responders{}
	if cq.Msg.Id != 0 {
		var callbackData MenuCallbackData
		err := json.Unmarshal([]byte(cq.Data), &callbackData)
		if err != nil {
			slog.Error(fmt.Sprintf("couldn't unmarshal %+v", cq))
			return SendMsg{}
		}
		switch callbackData.Stage {
		case 0:
			responders = append(responders, EditMsg{
				ChatId:      cq.Msg.Chat.Id,
				Text:        "Now choose the difficulty (how common or rare are the words)",
				ReplyMarkup: chooseLevelKeyboardMarkup(callbackData.Language),
				MsgId:       cq.Msg.Id})
		case 1:
			responders = append(responders,
				DeleteMsg{ChatId: cq.Msg.Chat.Id, MsgId: cq.Msg.Id},
				callbackHandler.nextWord(cq.Msg.Chat.Id, callbackData))
		case 2:
			responders = append(responders, callbackHandler.nextWord(cq.Msg.Chat.Id, callbackData))
		}
		return append(responders, AnswerCallback{CallbackQueryId: cq.Id})
	}
	return append(responders, SendMsg{}, AnswerCallback{CallbackQueryId: cq.Id})
}

func (callbackHandler CallbackHandler) nextWord(chatId int64, data MenuCallbackData) Responder {
	errorMsg := SendMsg{ChatId: chatId, Text: "Internal error. Try again later"}
	tableName := fmt.Sprintf("words_%s_%s", data.Language, strconv.Itoa(int(data.Difficulty)))
	// var rowCount int
	// query := fmt.Sprintf("SELECT COUNT(lemma) FROM %s", tableName)
	// err := callbackHandler.db.QueryRow(context.TODO(), query).Scan(&rowCount)
	// if err != nil {
	// 	slog.Error(fmt.Sprintf("during fetching "))
	// 	fmt.Println(err)
	// }
	// if rowCount == 0 {
	// 	return SendMsg{ChatId: chatId, Text: "no words for this language and difficulty level, yet. try again later or choose a different level"}
	// }
	query := fmt.Sprintf("SELECT * FROM %s ORDER BY RANDOM() LIMIT 1", tableName)
	rows, err := callbackHandler.db.Query(context.TODO(), query)
	if err != nil {
		slog.Error(fmt.Sprintf("couldn't retrieve rows from %s - menuCallbackData=%+v, error: %s", tableName, data, err.Error()))
		return errorMsg
	}
	words, err := pgx.CollectRows(rows, pgx.RowToStructByName[database.WordEntry])
	if err != nil {
		slog.Error(fmt.Sprintf("couldn't parse results from database into WordEntry struct - tableName=%s, menuCallbackData=%+v, error: %s", tableName, data, err.Error()))
		return errorMsg
	}
	if len(words) == 0 {
		return SendMsg{ChatId: chatId, Text: "no words for this language and difficulty level, yet. try again later or choose a different level"}
	}
	word := words[0]
	text := formatWordMsg(word)
	var parseMode string = "MarkdownV2"
	return SendMsg{ChatId: chatId, Text: text, ParseMode: &parseMode, ReplyMarkup: nextWordKeyboardMarkup(data)}
}

func formatWordMsg(word database.WordEntry) string {
	//check all nil pointers
	header := fmt.Sprintf("*%s*", word.Lemma)
	meaning := fmt.Sprintf("*meaning:* %s", word.LemmaMeaning)
	var sentences string
	if word.Sentences == nil {
		sentences = ""
	} else {
		sentences = fmt.Sprintf("*examples of sentences:*\n\n%s", formatExamples(*word.Sentences))
	}
	formattedWord := fmt.Sprintf("\n%s\n\n||%s\n\n%s||\n\n", header, meaning, sentences)
	formattedWord = excapeChars(formattedWord)
	return formattedWord
}

func formatExamples(examplesRaw string) string {
	// the examples in the database are of this form:
	// <example><sentence>...</sentence><translation>...</translation></example><example><sentence>...</sentence><translation></translation>...</example>
	// this function transforms it into a string like that:
	//
	// Examples of sentences
	// sentence 1
	// translation 1
	//
	// sentence 2
	// translation 2
	// ...

	xmlString := fmt.Sprintf("<examples>%s</examples>", examplesRaw) // have to wrap it in a root element for Unmarshaling
	var examples database.Examples
	err := xml.Unmarshal([]byte(xmlString), &examples)
	if err != nil {
		slog.Error(fmt.Sprintf("couldn't unmarshal example sentences. xmlString=%s, error: %s", xmlString, err.Error()))
		return "_no examples for this word yet_"
	}
	maxShownExamples := 3
	shuffleSlice(examples.Sentences, maxShownExamples)
	examplesSize := min(len(examples.Sentences), maxShownExamples)
	examplesFormatted := make([]string, examplesSize)
	for i := 0; i < examplesSize; i++ {
		examplesFormatted[i] = fmt.Sprintf("%s\n_%s_", examples.Sentences[i].Sentence, examples.Sentences[i].Translation)
	}
	return strings.Join(examplesFormatted, "\n\n")
}

func nextWordKeyboardMarkup(data MenuCallbackData) *telegramapi.InlineKeyboardMarkup {
	keyboard := make([][]telegramapi.InlineKeyboardButton, 1)
	data.Stage = 2
	dataJson, err := json.Marshal(data)
	if err != nil {
		slog.Error(fmt.Sprintf("couldn't unmarshal data: %+v, error: %s", data, err.Error()))
		fmt.Println(err)
	}
	keyboard[0] = []telegramapi.InlineKeyboardButton{{Text: "Next", CallbackData: string(dataJson)}}
	return &telegramapi.InlineKeyboardMarkup{InlineKeyboard: keyboard}
}
