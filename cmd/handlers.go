package main

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CallbackReply func(CallbackQuery) Responder
type MsgReply func(Message) Responder

type Responders []Responder

// this makes a list of Responders also a Responder
func (r Responders) Respond(baseUrl string) {
	for _, responder := range r {
		responder.Respond(baseUrl)
	}
}

// TODO: do it with more generic type instead of Responder
func postRequest(baseUrl string, method string, r Responder) {
	data, err := json.Marshal(r)
	if err != nil {
		fmt.Println("failed to marshal: ", r)
		fmt.Println(err)
	}
	url := fmt.Sprintf("%s/%s", baseUrl, method)
	client := &http.Client{}
	res, err := client.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		// TODO: do smth if Telegram gives an error back
		fmt.Printf("error on post request: %s\n", err)
	}
	defer res.Body.Close()
	// TODO: define a new type for getting Telegram errors
	var j interface{}
	err = json.NewDecoder(res.Body).Decode(&j)
	if err != nil {
		fmt.Println("serializing error: ", err)
	}
	fmt.Println("response from telegram")
	fmt.Println(j)
}

type CommandHandler struct {
	db *pgxpool.Pool
}

func NewCommandHandler(db *pgxpool.Pool) CommandHandler {
	return CommandHandler{db: db}
}

func (cmdHandler CommandHandler) GetResponder(msg Message) Responder {
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

type CallbackHandler struct {
	db *pgxpool.Pool
}

func NewCallbackHandler(db *pgxpool.Pool) (CallbackHandler, error) {
	// supportedLangs, err := getSupportedLanguages(db)
	// if err != nil {
	// 	fmt.Println("could not initiate CallbackHandler")
	// 	return CallbackHandler{}, err
	// }
	return CallbackHandler{db}, nil
}

func (callbackHandler CallbackHandler) GetResponder(cq CallbackQuery) Responder {
	if cq.Msg.Id != 0 {
		var callbackData MenuCallbackData
		err := json.Unmarshal([]byte(cq.Data), &callbackData)
		if err != nil {
			fmt.Println(err)
			return SendMsg{}
		}
		switch callbackData.Stage {
		case 0:
			return EditMsg{
				ChatId:      cq.Msg.Chat.Id,
				Text:        "Now choose the difficulty (how common or rare are the words)",
				ReplyMarkup: chooseLevelKeyboardMarkup(callbackData.Language),
				MsgId:       cq.Msg.Id}
		case 1:
			dm := DeleteMsg{ChatId: cq.Msg.Chat.Id, MsgId: cq.Msg.Id}
			sm := callbackHandler.nextWord(cq.Msg.Chat.Id, callbackData)
			return Responders{dm, sm}
		case 2:
			return callbackHandler.nextWord(cq.Msg.Chat.Id, callbackData)
		}
	}
	return SendMsg{}
}

func (callbackHandler CallbackHandler) nextWord(chatId int64, data MenuCallbackData) Responder {
	tableName := fmt.Sprintf("words_%s_%s", data.Language, strconv.Itoa(int(data.Difficulty)))
	var rowCount int
	query := fmt.Sprintf("SELECT COUNT(lemma) FROM %s", tableName)
	err := callbackHandler.db.QueryRow(context.TODO(), query).Scan(&rowCount)
	if err != nil {
		fmt.Println(err)
	}
	if rowCount == 0 {
		return SendMsg{ChatId: chatId, Text: "no words for this language and difficulty level, yet. try again later or choose a different level"}
	}
	query = fmt.Sprintf("SELECT * FROM %s ORDER BY RANDOM() LIMIT 1", tableName)
	rows, err := callbackHandler.db.Query(context.TODO(), query)
	if err != nil {
		fmt.Println(err)
		return SendMsg{}
	}
	words, err := pgx.CollectRows(rows, pgx.RowToStructByName[WordEntry])
	if err != nil {
		fmt.Println(err)
	}
	if len(words) == 0 {
		return SendMsg{ChatId: chatId, Text: "no words for this language and difficulty level, yet. try again later or choose a different level"}
	}
	word := words[0]
	text, err := formatWordMsg(word)
	// TODO: on error retry a few times to get a valid word
	if err != nil {
		return SendMsg{}
	}
	var parseMode string = "MarkdownV2"
	return SendMsg{ChatId: chatId, Text: text, ParseMode: &parseMode, ReplyMarkup: nextWordKeyboardMarkup(data)}
}

type SendMsg struct {
	ChatId      int64                 `json:"chat_id"`
	Text        string                `json:"text"`
	ParseMode   *string               `json:"parse_mode,omitempty"`   // pointer for omitempty to skip it if its not present
	ReplyMarkup *InlineKeyboardMarkup `json:"reply_markup,omitempty"` // pointer for omitempty to skip it if its not present
}

func (sm SendMsg) Respond(baseUrl string) {
	postRequest(baseUrl, "sendMessage", sm)
}

// TODO: there should be a way to embed SendMsg into this, its just one field more
type EditMsg struct {
	ChatId      int64                 `json:"chat_id"`
	Text        string                `json:"text"`
	ReplyMarkup *InlineKeyboardMarkup `json:"reply_markup,omitempty"` // has to be a pointer for json serializer to skip it if empty
	MsgId       int64                 `json:"message_id"`             // id of a message to edit
}

func (em EditMsg) Respond(baseUrl string) {
	postRequest(baseUrl, "editMessageText", em)
}

type DeleteMsg struct {
	ChatId int64 `json:"chat_id"`
	MsgId  int64 `json:"message_id"`
}

func (dm DeleteMsg) Respond(baseUrl string) {
	postRequest(baseUrl, "deleteMessage", dm)
}

type MenuCallbackData struct {
	Stage      uint8  `json:"stage"`
	Language   string `json:"language"`
	Difficulty uint8  `json:"difficulty"`
}

func start(msg Message, db *pgxpool.Pool) Responder {
	text := "Welcome"
	// TODO: saving a user to the database
	// check if this is a group or private chat
	username := msg.Chat.Username
	if username != "" {
		err := createOrUpdatePrivateChat(db, username, msg.Chat.Id)
		if err != nil {
			fmt.Printf("couldn't update user %s to in a database", username)
			fmt.Println(err)
		}
	}
	return SendMsg{ChatId: msg.Chat.Id, Text: text}
}

func menu(msg Message, db *pgxpool.Pool) Responder {
	text := "Choose a language"
	supportedLangs, err := getSupportedLanguages(db)
	if err != nil {
		fmt.Println(err)
		return SendMsg{}
	}
	return SendMsg{ChatId: msg.Chat.Id, Text: text, ReplyMarkup: chooseLanguageKeyboardMarkup(supportedLangs)}
}

func about(msg Message) Responder {
	text := "about this bot"
	return SendMsg{ChatId: msg.Chat.Id, Text: text}
}

func help(msg Message) Responder {
	text := "here are all the commands"
	return SendMsg{ChatId: msg.Chat.Id, Text: text}
}

func formatWordMsg(word WordEntry) (string, error) {
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
	return formattedWord, nil
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
	var examples Examples
	err := xml.Unmarshal([]byte(xmlString), &examples)
	if err != nil {
		fmt.Println(err)
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

func (cq CallbackQuery) answer(baseUrl string) {
	// this function needs to be called to stop buttons from
	// blinking after they're pressed by the user
	url := fmt.Sprintf("%s/answerCallbackQuery", baseUrl)
	data := fmt.Sprintf("{\"callback_query_id\":\"%s\"}", cq.Id)
	client := &http.Client{}
	res, _ := client.Post(url, "application/json", bytes.NewBufferString(data)) //I dont even care about this error, no big deal
	res.Body.Close()
}

func chooseLanguageKeyboardMarkup(languages []Language) *InlineKeyboardMarkup {
	buttonsNum := len(languages)
	keyboard := make([][]InlineKeyboardButton, buttonsNum)
	for i, lang := range languages {
		data := MenuCallbackData{Stage: 0, Language: lang.code}
		callbackData, err := json.Marshal(data)
		if err != nil {
			fmt.Println(err)
		}
		keyboard[i] = []InlineKeyboardButton{{Text: lang.fullName, CallbackData: string(callbackData)}}
		i++
	}
	return &InlineKeyboardMarkup{InlineKeyboard: keyboard}
}

func chooseLevelKeyboardMarkup(langCode string) *InlineKeyboardMarkup {
	buttonsNum := len(difficultyToString)
	keyboard := make([][]InlineKeyboardButton, buttonsNum)
	for i := 0; i < buttonsNum; i++ {
		data := MenuCallbackData{Stage: 1, Language: langCode, Difficulty: uint8(i)}
		callbackData, err := json.Marshal(data)
		if err != nil {
			fmt.Println(err)
		}
		keyboard[i] = []InlineKeyboardButton{{Text: difficultyToString[uint8(i)], CallbackData: string(callbackData)}}
	}
	return &InlineKeyboardMarkup{InlineKeyboard: keyboard}
}

func nextWordKeyboardMarkup(data MenuCallbackData) *InlineKeyboardMarkup {
	keyboard := make([][]InlineKeyboardButton, 1)
	data.Stage = 2
	dataJson, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err)
	}
	keyboard[0] = []InlineKeyboardButton{{Text: "Next", CallbackData: string(dataJson)}}
	return &InlineKeyboardMarkup{InlineKeyboard: keyboard}
}
